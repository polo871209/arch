"""gRPC client for User service."""

import logging
import sys
from pathlib import Path
from typing import Optional

import grpc

# Add the client directory to Python path for proto imports
client_dir = Path(__file__).parent.parent.parent
sys.path.insert(0, str(client_dir))

# Import protobuf stub
from proto import UserServiceStub  # noqa: E402

# Local imports
from ..core.config import settings  # noqa: E402
from ..core.exceptions import (  # noqa: E402
    GRPCClientError,
    GRPCServiceUnavailableError,
)

logger = logging.getLogger(__name__)


class UserGRPCClient:
    """gRPC client for User service with connection management."""

    def __init__(self, host: Optional[str] = None, port: Optional[int] = None) -> None:
        """Initialize gRPC client with optional host/port override."""
        self._host = host or settings.grpc_host
        self._port = port or settings.grpc_port
        self._address = f"{self._host}:{self._port}"
        self._channel: Optional[grpc.Channel] = None
        self._stub: Optional[UserServiceStub] = None

    def __enter__(self) -> "UserGRPCClient":
        """Context manager entry."""
        self.connect()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb) -> None:
        """Context manager exit."""
        self.close()

    def connect(self) -> None:
        """Establish connection to gRPC server."""
        try:
            self._channel = grpc.insecure_channel(self._address)
            self._stub = UserServiceStub(self._channel)
            logger.info(f"Connected to gRPC server at {self._address}")
        except Exception as e:
            logger.error(f"Failed to connect to gRPC server at {self._address}: {e}")
            raise GRPCServiceUnavailableError(f"Failed to connect to gRPC server: {e}")

    def close(self) -> None:
        """Close gRPC connection."""
        if self._channel:
            self._channel.close()
            self._channel = None
            self._stub = None
            logger.info("gRPC connection closed")

    @property
    def stub(self) -> UserServiceStub:
        """Get the gRPC stub, connecting if necessary."""
        if not self._stub:
            self.connect()
        if not self._stub:
            raise GRPCClientError("gRPC client not connected")
        return self._stub

    @property
    def is_connected(self) -> bool:
        """Check if client is connected."""
        return self._stub is not None

    def health_check(self) -> bool:
        """Perform a basic health check by attempting to connect."""
        try:
            if not self.is_connected:
                self.connect()
            return True
        except Exception:
            return False
