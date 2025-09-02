from pydantic import Field, computed_field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Application configuration with environment variable validation."""

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False,
        extra="forbid",
    )

    # Server settings
    grpc_host: str = Field(default="localhost", description="gRPC server host")
    grpc_port: int = Field(default=50051, description="gRPC server port")

    # API settings
    api_host: str = Field(default="0.0.0.0", description="API server host")
    api_port: int = Field(default=8000, description="API server port")
    api_reload: bool = Field(default=False, description="Enable API auto-reload")

    # Application settings
    app_name: str = Field(default="User gRPC Client", description="Application name")
    app_version: str = Field(default="0.1.0", description="Application version")
    debug: bool = Field(default=False, description="Enable debug mode")

    # Logging
    log_level: str = Field(default="INFO", description="Logging level")

    @computed_field  # type: ignore[prop-decorator]
    @property
    def grpc_address(self) -> str:
        """Get the complete gRPC server address."""
        return f"{self.grpc_host}:{self.grpc_port}"


# Create settings instance - will read from environment variables or use defaults
settings = Settings()
