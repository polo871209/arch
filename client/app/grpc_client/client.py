import logging
import sys
from pathlib import Path
from typing import Dict, Optional

from grpc import aio

from ..core.config import settings
from ..core.exceptions import (
    GRPCClientError,
    GRPCServiceUnavailableError,
)
from ..core.tracing import create_grpc_metadata

client_dir = Path(__file__).parent.parent.parent
sys.path.insert(0, str(client_dir))

from proto.user_pb2_grpc import UserServiceStub  # noqa: E402

logger = logging.getLogger(__name__)


class AsyncUserGRPCClient:
    def __init__(self) -> None:
        self._host = settings.grpc_host
        self._port = settings.grpc_port
        self._address = f"{self._host}:{self._port}"
        self._channel: aio.Channel | None = None
        self._stub: UserServiceStub | None = None
        self._additional_headers: Dict[str, str] = {}

    def set_additional_headers(self, headers: Dict[str, str]) -> None:
        """Set additional headers (non-tracing) for gRPC calls."""
        self._additional_headers = headers

    def get_metadata(self, context_headers: Optional[Dict[str, str]] = None) -> list:
        """
        Get gRPC metadata with trace context and additional headers.

        Args:
            context_headers: Headers from request context (for trace propagation)

        Returns:
            List of metadata tuples for gRPC call
        """
        metadata = []

        # Add trace context headers if provided
        if context_headers:
            metadata.extend(create_grpc_metadata(context_headers))

        # Add any additional headers
        if self._additional_headers:
            metadata.extend(create_grpc_metadata(self._additional_headers))

        return metadata

    async def __aenter__(self) -> "AsyncUserGRPCClient":
        """Async context manager entry."""
        await self.connect()
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb) -> None:
        """Async context manager exit."""
        await self.close()

    async def connect(self) -> None:
        try:
            self._channel = aio.insecure_channel(self._address)
            self._stub = UserServiceStub(self._channel)
            # Try to make a dummy call to ensure connection is healthy
            await self._channel.channel_ready()
            logger.debug(f"Connected to gRPC server at {self._address}")
        except Exception as e:
            logger.error(f"Failed to connect to gRPC server at {self._address}: {e}")
            raise GRPCServiceUnavailableError(
                f"Failed to connect to gRPC server: {e}"
            ) from e

    async def close(self) -> None:
        if self._channel:
            await self._channel.close()
            self._channel = None
            self._stub = None
            logger.info("gRPC connection closed")

    @property
    def stub(self) -> UserServiceStub:
        """Get the gRPC stub, connecting if necessary."""
        if not self._stub:
            raise GRPCClientError(
                "gRPC client not connected. Use 'await connect()' first."
            )
        return self._stub

    @property
    def is_connected(self) -> bool:
        """Check if client is connected."""
        return self._stub is not None and self._channel is not None

    async def health_check(self) -> bool:
        """Perform a basic health check."""
        try:
            await self.connect()
            return True
        except Exception as e:
            logger.warning(f"Health check failed: {e}")
            return False
