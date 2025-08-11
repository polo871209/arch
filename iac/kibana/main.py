import json
import os
from pathlib import Path
from typing import Any

import httpx

from src.api.dataviews import DataViewsAPI, DataViewSpec
from src.api.http import KibanaAuth, KibanaHTTP
from src.config import AppConfig

CONFIG_FILE = Path(os.getenv("KIBANA_IAC_CONFIG", "kibana.iac.json"))


def run_sync(cfg: AppConfig) -> dict[str, Any]:
    auth = KibanaAuth(base_url=cfg.base_url, api_key=cfg.api_key, space_id=cfg.space_id)
    with KibanaHTTP(auth, timeout=cfg.timeout, http2=cfg.http2, verify=cfg.verify_ssl) as http:
        api = DataViewsAPI(http)

        desired_specs = [
            DataViewSpec(
                name=str(item.get("name")),
                title=item.get("title"),
                timeFieldName=item.get("timeFieldName"),
                allowNoIndex=bool(item.get("allowNoIndex", False)),
                fieldFormats=item.get("fieldFormats"),
                refresh_fields=bool(item.get("refresh_fields", False)),
            )
            for item in cfg.data_views
            if isinstance(item, dict) and item.get("name")
        ]

        summary = api.sync(desired_specs)
        return summary


def main() -> None:
    cfg = AppConfig.load(CONFIG_FILE)
    if not cfg.base_url or not cfg.api_key or cfg.api_key.startswith("<"):
        raise SystemExit(
            "kibana.base_url and kibana.api_key are required in config file (set a real API key)"
        )

    try:
        summary = run_sync(cfg)
        print(json.dumps(summary, indent=2))
    except httpx.HTTPStatusError as e:
        sc = e.response.status_code
        if sc == 401:
            raise SystemExit(
                "Unauthorized (401): API key invalid or missing required scheme. "
                "Set kibana.api_key (e.g., 'ApiKey <token>' or raw token) and try again."
            ) from e
        if sc == 403:
            raise SystemExit(
                "Forbidden (403): API key lacks required Kibana privileges for Data Views endpoints."
            ) from e
        raise SystemExit(f"HTTP error {sc}: {e}") from e
    except httpx.HTTPError as e:
        raise SystemExit(f"HTTP client error: {e}") from e
    except Exception as e:
        raise SystemExit(str(e)) from e


if __name__ == "__main__":
    main()
