from dataclasses import dataclass
from typing import Any

import httpx


@dataclass
class KibanaAuth:
    base_url: str
    api_key: str
    space_id: str | None = None


class KibanaHTTP:
    """Minimal HTTP helper for Kibana API requests."""

    def __init__(
        self, auth: KibanaAuth, *, timeout: float = 20.0, http2: bool = True, verify: bool = True
    ):
        self.auth = auth
        self._client = httpx.Client(timeout=timeout, http2=http2, verify=verify)

    @staticmethod
    def _auth_value(raw: str) -> str:
        """Return normalized Authorization header value."""
        if not raw:
            return raw
        val = raw.strip()
        lowered = val.lower()
        if lowered.startswith(("apikey ", "bearer ", "basic ")):
            return val
        return f"ApiKey {val}"

    def _url(self, path: str) -> str:
        path = path if path.startswith("/") else "/" + path
        if self.auth.space_id:
            encoded_space = httpx.QueryParams({}).encode_component(self.auth.space_id)
            return f"{self.auth.base_url}/s/{encoded_space}/{path.lstrip('/')}"
        return f"{self.auth.base_url}{path}"

    def _headers(
        self, *, xsrf: bool = False, extra: dict[str, str] | None = None
    ) -> dict[str, str]:
        headers: dict[str, str] = {"Authorization": self._auth_value(self.auth.api_key)}
        if xsrf:
            headers["kbn-xsrf"] = "true"
        if extra:
            headers.update(extra)
        return headers

    def get(self, path: str) -> httpx.Response:
        return self._client.get(self._url(path), headers=self._headers())

    def post(
        self, path: str, json: dict[str, Any] | None = None, *, xsrf: bool = True
    ) -> httpx.Response:
        return self._client.post(self._url(path), headers=self._headers(xsrf=xsrf), json=json)

    def delete(self, path: str, *, xsrf: bool = True) -> httpx.Response:
        return self._client.delete(self._url(path), headers=self._headers(xsrf=xsrf))

    def close(self) -> None:
        self._client.close()

    def __enter__(self) -> "KibanaHTTP":
        return self

    def __exit__(self, exc_type, exc, tb) -> None:
        self.close()
