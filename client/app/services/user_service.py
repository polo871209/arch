"""Business logic service for user operations."""

import logging
import sys
from pathlib import Path

import grpc

from ..core.exceptions import grpc_to_http_exception
from ..grpc_client import AsyncUserGRPCClient
from ..models import (
    MessageResponse,
    UserCreate,
    UserListResponse,
    UserResponse,
    UserUpdate,
)

client_dir = Path(__file__).parent.parent.parent
sys.path.insert(0, str(client_dir))

from proto.user_pb2 import (  # noqa: E402
    CreateUserRequest,
    DeleteUserRequest,
    GetUserRequest,
    ListUsersRequest,
    UpdateUserRequest,
    User,
)

logger = logging.getLogger(__name__)


class UserService:
    def __init__(self, grpc_client: AsyncUserGRPCClient) -> None:
        self.grpc_client = grpc_client

    async def create_user(self, user_data: UserCreate) -> UserResponse:
        try:
            request = CreateUserRequest(
                name=user_data.name,
                email=user_data.email,
                age=user_data.age,
            )
            response = await self.grpc_client.stub.CreateUser(
                request, metadata=self.grpc_client.get_metadata()
            )
            return self._grpc_user_to_pydantic(response.user)
        except grpc.RpcError as e:
            logger.error(f"gRPC error creating user: {e}")
            raise grpc_to_http_exception(e) from e

    async def get_user(self, user_id: str) -> UserResponse:
        try:
            request = GetUserRequest(id=user_id)
            response = await self.grpc_client.stub.GetUser(
                request, metadata=self.grpc_client.get_metadata()
            )
            return self._grpc_user_to_pydantic(response.user)
        except grpc.RpcError as e:
            logger.error(f"gRPC error getting user {user_id}: {e}")
            raise grpc_to_http_exception(e) from e

    async def update_user(self, user_id: str, user_data: UserUpdate) -> UserResponse:
        try:
            request = UpdateUserRequest(
                id=user_id,
                name=user_data.name or "",
                email=user_data.email or "",
                age=user_data.age or 0,
            )
            response = await self.grpc_client.stub.UpdateUser(
                request, metadata=self.grpc_client.get_metadata()
            )
            return self._grpc_user_to_pydantic(response.user)
        except grpc.RpcError as e:
            logger.error(f"gRPC error updating user {user_id}: {e}")
            raise grpc_to_http_exception(e) from e

    async def delete_user(self, user_id: str) -> MessageResponse:
        try:
            request = DeleteUserRequest(id=user_id)
            response = await self.grpc_client.stub.DeleteUser(
                request, metadata=self.grpc_client.get_metadata()
            )
            return MessageResponse(message=response.message)
        except grpc.RpcError as e:
            logger.error(f"gRPC error deleting user {user_id}: {e}")
            raise grpc_to_http_exception(e) from e

    async def list_users(self, page: int = 1, limit: int = 10) -> UserListResponse:
        try:
            request = ListUsersRequest(page=page, limit=limit)
            response = await self.grpc_client.stub.ListUsers(
                request, metadata=self.grpc_client.get_metadata()
            )

            users = [self._grpc_user_to_pydantic(user) for user in response.users]
            return UserListResponse(
                users=users,
                total=response.total,
                message=response.message,
            )
        except grpc.RpcError as e:
            logger.error(f"gRPC error listing users: {e}")
            raise grpc_to_http_exception(e) from e

    def _grpc_user_to_pydantic(self, grpc_user: User) -> UserResponse:
        return UserResponse(
            id=grpc_user.id,
            name=grpc_user.name,
            email=grpc_user.email,
            age=grpc_user.age,
            created_at=grpc_user.created_at,
            updated_at=grpc_user.updated_at,
        )
