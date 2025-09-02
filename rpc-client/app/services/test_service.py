"""Business logic service for test operations."""

import logging
import sys
from pathlib import Path

import grpc

from ..core.exceptions import grpc_to_http_exception
from ..grpc_client import AsyncUserGRPCClient
from ..models import MessageResponse

client_dir = Path(__file__).parent.parent.parent
sys.path.insert(0, str(client_dir))

from proto.user_pb2 import TestErrorRequest  # noqa: E402

logger = logging.getLogger(__name__)


class TestService:
    def __init__(self, grpc_client: AsyncUserGRPCClient) -> None:
        self.grpc_client = grpc_client

    async def test_error(self, status_code: str) -> MessageResponse:
        try:
            request = TestErrorRequest(status_code=status_code)
            response = await self.grpc_client.stub.TestError(request)
            return MessageResponse(message=response.message)
        except grpc.RpcError as e:
            logger.error(f"gRPC error from test error endpoint: {e}")
            raise grpc_to_http_exception(e) from e
