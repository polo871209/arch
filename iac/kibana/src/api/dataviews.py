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

    def as_payload(self) -> dict[str, Any]:
        dv: dict[str, Any] = {"name": self.name}
        # Derive a reasonable default title from name if not provided.
        # Kibana requires 'title' to resolve indices. Use a slug-like transform.
        derived_title = None
        if not self.title and self.name:
            import re

            # Lowercase, replace non-word with underscores, collapse repeats, strip underscores
            slug = re.sub(r"\W+", "_", self.name.lower())
            slug = re.sub(r"_+", "_", slug).strip("_")
            derived_title = slug or self.name
        dv["title"] = self.title or derived_title
        if self.timeFieldName:
            dv["timeFieldName"] = self.timeFieldName
        if self.allowNoIndex:
            dv["allowNoIndex"] = True
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

    # --- Raw endpoints ---
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
        r.raise_for_status()
        return r.json()

    def get(self, view_id: str) -> dict[str, Any]:
        encoded = httpx.URL("/").copy_with(path=f"/{view_id}").path.lstrip("/")
        r = self.http.get(self.GET.format(id=encoded))
        r.raise_for_status()
        return r.json()

    def update(
        self, view_id: str, spec: DataViewSpec, *, refresh_fields: bool = False
    ) -> dict[str, Any]:
        payload: dict[str, Any] = {"data_view": spec.as_payload()}
        if refresh_fields:
            payload["refresh_fields"] = True
        encoded = httpx.URL("/").copy_with(path=f"/{view_id}").path.lstrip("/")
        r = self.http.post(self.UPDATE.format(id=encoded), json=payload, xsrf=True)
        r.raise_for_status()
        return r.json()

    def delete(self, view_id: str) -> None:
        encoded = httpx.URL("/").copy_with(path=f"/{view_id}").path.lstrip("/")
        r = self.http.delete(self.DELETE.format(id=encoded), xsrf=True)
        if r.status_code not in (200, 202, 204, 404):
            r.raise_for_status()

    # --- Declarative sync ---
    def sync(self, desired: list[DataViewSpec]) -> dict[str, Any]:
        """Make Kibana match the desired list of DataViewSpec.

        Behavior:
        - Upsert any views present in desired (create override then update when ID known).
        - Delete any existing views not present in desired (name match).
        Notes:
        - If a spec has no title, we will try to update an existing view by name only; if none exists, it is skipped.
        Returns summary with created/updated/deleted/skipped lists.
        """
        existing_by_name = {dv.get("name"): dv for dv in self.list_views() if isinstance(dv, dict)}

        desired_names = {d.name for d in desired}

        created: list[dict[str, Any]] = []
        updated: list[dict[str, Any]] = []
        deleted: list[dict[str, Any]] = []
        skipped: list[dict[str, Any]] = []

        for spec in desired:
            if not spec.title and spec.name not in existing_by_name:
                skipped.append({"reason": "missing_title", "name": spec.name})
                continue

            resp = self.create(spec, override=True)
            dv_obj = resp.get("data_view", {}) if isinstance(resp, dict) else {}
            if (view_id := dv_obj.get("id")):
                updated.append(self.update(view_id, spec, refresh_fields=False))
            else:
                created.append(resp)

        for name, dv in existing_by_name.items():
            if name and name not in desired_names and (view_id := dv.get("id")):
                self.delete(view_id)
                deleted.append({"id": view_id, "name": name})

        return {"created": created, "updated": updated, "deleted": deleted, "skipped": skipped}
