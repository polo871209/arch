import os


class Settings:
    def __init__(self) -> None:
        # Server settings
        self.grpc_host: str = self._require_env("GRPC_HOST")
        self.grpc_port: int = int(self._require_env("GRPC_PORT"))

        # API settings
        self.api_host: str = self._require_env("API_HOST")
        self.api_port: int = int(self._require_env("API_PORT"))
        self.api_reload: bool = self._require_env("API_RELOAD").lower() == "true"

        # Application settings
        self.app_name: str = self._require_env("APP_NAME")
        self.app_version: str = self._require_env("APP_VERSION")
        self.debug: bool = self._require_env("DEBUG").lower() == "true"

        # Logging
        self.log_level: str = self._require_env("LOG_LEVEL")

    def _require_env(self, key: str) -> str:
        """Get environment variable or raise error if not set."""
        value = os.getenv(key)
        if value is None:
            raise ValueError(f"Environment variable {key} is required but not set")
        return value

    @property
    def grpc_address(self) -> str:
        """Get the complete gRPC server address."""
        return f"{self.grpc_host}:{self.grpc_port}"


settings = Settings()
