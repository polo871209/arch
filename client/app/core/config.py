import os


class Settings:
    def __init__(self) -> None:
        # Server settings
        self.grpc_host: str = os.getenv("GRPC_HOST", "localhost")
        self.grpc_port: int = int(os.getenv("GRPC_PORT", "50051"))

        # API settings
        self.api_host: str = os.getenv("API_HOST", "0.0.0.0")
        self.api_port: int = int(os.getenv("API_PORT", "8000"))
        self.api_reload: bool = os.getenv("API_RELOAD", "true").lower() == "true"

        # Application settings
        self.app_name: str = os.getenv("APP_NAME", "User Management API")
        self.app_version: str = os.getenv("APP_VERSION", "1.0.0")
        self.debug: bool = os.getenv("DEBUG", "false").lower() == "true"

        # Logging
        self.log_level: str = os.getenv("LOG_LEVEL", "INFO")

    @property
    def grpc_address(self) -> str:
        """Get the complete gRPC server address."""
        return f"{self.grpc_host}:{self.grpc_port}"


settings = Settings()
