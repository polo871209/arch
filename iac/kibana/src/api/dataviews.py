import re
from dataclasses import dataclass
from typing import Any

import httpx

from .http import KibanaHTTP


@dataclass
class DataViewSpec:
    name: str
    title: str | None = None
    timeFieldName: str | None = None
    allowNoIndex: bool = False
    fieldFormats: dict[str, Any] | None = None
    refresh_fields: bool = False

    def as_payload(self) -> dict[str, Any]:
        dv: dict[str, Any] = {"name": self.name}
        # Derive a reasonable default title from name if not provided
        if not self.title and self.name:
            # Lowercase, replace non-word with underscores, collapse repeats, strip underscores
            slug = re.sub(r"\W+", "_", self.name.lower())
            slug = re.sub(r"_+", "_", slug).strip("_")
            derived_title = slug or self.name
        else:
            derived_title = None

        dv["title"] = self.title or derived_title
        if self.timeFieldName:
            dv["timeFieldName"] = self.timeFieldName
        if self.allowNoIndex:
            dv["allowNoIndex"] = True
        if self.fieldFormats:
            dv["fieldFormats"] = self.fieldFormats
        return dv


class DataViewsAPI:
    """Data Views CRUD wrapper and declarative sync."""

    LIST = "/api/data_views"
    CREATE = "/api/data_views/data_view"
    GET = "/api/data_views/data_view/{id}"
    UPDATE = "/api/data_views/data_view/{id}"
    DELETE = "/api/data_views/data_view/{id}"

    def __init__(self, http: KibanaHTTP):
        self.http = http

    def _encode_id(self, view_id: str) -> str:
        """URL encode the view ID safely."""
        return httpx.URL("/").copy_with(path=f"/{view_id}").path.lstrip("/")

    def list_views(self) -> list[dict[str, Any]]:
        r = self.http.get(self.LIST)
        r.raise_for_status()
        body = r.json()
        return body.get("data_view", []) if isinstance(body, dict) else []

    def create(self, spec: DataViewSpec, *, override: bool = False) -> dict[str, Any]:
        payload: dict[str, Any] = {"data_view": spec.as_payload()}
        if override:
            payload["override"] = True
        r = self.http.post(self.CREATE, json=payload, xsrf=True)
        if r.status_code >= 400:
            raise httpx.HTTPStatusError(
                f"Create failed: {r.status_code} {r.text}", request=r.request, response=r
            )
        return r.json()

    def get(self, view_id: str) -> dict[str, Any]:
        r = self.http.get(self.GET.format(id=self._encode_id(view_id)))
        r.raise_for_status()
        return r.json()

    def update(
        self, view_id: str, spec: DataViewSpec, *, refresh_fields: bool = False
    ) -> dict[str, Any]:
        payload: dict[str, Any] = {"data_view": spec.as_payload()}
        if refresh_fields:
            payload["refresh_fields"] = True
        r = self.http.post(self.UPDATE.format(id=self._encode_id(view_id)), json=payload, xsrf=True)
        r.raise_for_status()
        return r.json()

    def delete(self, view_id: str) -> None:
        r = self.http.delete(self.DELETE.format(id=self._encode_id(view_id)), xsrf=True)
        if r.status_code not in (200, 202, 204, 404):
            r.raise_for_status()

    def set_default(self, view_id: str, *, force: bool = True) -> dict[str, Any]:
        payload = {"data_view_id": view_id, "force": force}
        r = self.http.post("/api/data_views/default", json=payload, xsrf=True)
        r.raise_for_status()
        return r.json()

    def sync(self, desired: list[DataViewSpec]) -> dict[str, Any]:
        """Make Kibana match the desired list of DataViewSpec.

        Upserts any views present in desired and deletes any existing views not present.
        Returns summary with created/updated/deleted/skipped lists.
        """
        existing_by_name = {dv.get("name"): dv for dv in self.list_views() if isinstance(dv, dict)}
        desired_names = {d.name for d in desired}

        created: list[dict[str, Any]] = []
        updated: list[dict[str, Any]] = []
        deleted: list[dict[str, Any]] = []
        skipped: list[dict[str, Any]] = []

        first_view_id: str | None = None
        for spec in desired:
            if not spec.title and spec.name not in existing_by_name:
                skipped.append({"reason": "missing_title", "name": spec.name})
                continue

            resp = self.create(spec, override=True)
            dv_obj = resp.get("data_view", {}) if isinstance(resp, dict) else {}
            if view_id := dv_obj.get("id"):
                updated.append(self.update(view_id, spec, refresh_fields=spec.refresh_fields))
                if first_view_id is None:
                    first_view_id = view_id
            else:
                created.append(resp)
                if first_view_id is None and (maybe_id := dv_obj.get("id")):
                    first_view_id = maybe_id

        for name, dv in existing_by_name.items():
            if name and name not in desired_names and (view_id := dv.get("id")):
                self.delete(view_id)
                deleted.append({"id": view_id, "name": name})

        # Set the first desired data view as default
        if first_view_id:
            try:
                self.set_default(first_view_id, force=True)
            except Exception as e:  # noqa: BLE001
                skipped.append({"reason": "set_default_failed", "error": str(e)})

        return {"created": created, "updated": updated, "deleted": deleted, "skipped": skipped}
