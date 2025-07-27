import logging
import sys
from pathlib import Path
from typing import Optional

import grpc

from ..core.config import settings
from ..core.exceptions import (
    GRPCClientError,
    GRPCServiceUnavailableError,
)

client_dir = Path(__file__).parent.parent.parent
sys.path.insert(0, str(client_dir))

from proto.user_pb2_grpc import UserServiceStub  # noqa: E402

logger = logging.getLogger(__name__)


class UserGRPCClient:
    def __init__(self) -> None:
        self._host = settings.grpc_host
        self._port = settings.grpc_port
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
        try:
            self._channel = grpc.insecure_channel(self._address)
            self._stub = UserServiceStub(self._channel)
            logger.debug(f"Connected to gRPC server at {self._address}")
        except Exception as e:
            logger.error(f"Failed to connect to gRPC server at {self._address}: {e}")
            raise GRPCServiceUnavailableError(f"Failed to connect to gRPC server: {e}")

    def close(self) -> None:
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
