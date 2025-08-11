import json
from dataclasses import dataclass
from pathlib import Path
from typing import Any


@dataclass
class AppConfig:
    base_url: str
    api_key: str
    space_id: str | None
    timeout: float
    http2: bool
    verify_ssl: bool
    data_views: list[dict[str, Any]]

    @staticmethod
    def load(path: Path) -> "AppConfig":
        if not path.exists():
            template = {
                "kibana": {
                    "base_url": "http://kibana:5601",
                    "api_key": "<your kibana api key>",
                    "space_id": None,
                    "timeout": 20.0,
                    "http2": True,
                    "verify_ssl": False,
                },
                "data_views": [
                    {
                        "name": "Kibana Sample Data eCommerce",
                        "title": "kibana_sample_data_ecommerce",
                    }
                ],
            }
            path.write_text(json.dumps(template, indent=2))
            raise SystemExit(
                f"Config template created at {path}. Please review and set kibana.api_key, then re-run."
            )

        cfg = json.loads(path.read_text())
        kib = cfg.get("kibana", {})
        data_views = cfg.get("data_views", [])

        return AppConfig(
            base_url=str(kib.get("base_url", "")).rstrip("/"),
            api_key=str(kib.get("api_key", "")),
            space_id=kib.get("space_id"),
            timeout=float(kib.get("timeout", 20.0)),
            http2=bool(kib.get("http2", True)),
            verify_ssl=bool(kib.get("verify_ssl", False)),
            data_views=data_views if isinstance(data_views, list) else [],
        )
