"""Business logic service for user operations."""

import logging
import sys
from pathlib import Path

import grpc

# Add the client directory to Python path for proto imports
client_dir = Path(__file__).parent.parent.parent  
sys.path.insert(0, str(client_dir))

# Import protobuf classes  
from proto import (  # noqa: E402
    CreateUserRequest,
    DeleteUserRequest,
    GetUserRequest,
    ListUsersRequest,
    UpdateUserRequest,
    User,
)

# Local imports
from ..core.exceptions import grpc_to_http_exception  # noqa: E402
from ..grpc_client import UserGRPCClient  # noqa: E402
from ..models import (  # noqa: E402
    MessageResponse,
    UserCreate,
    UserListResponse,
    UserResponse,
    UserUpdate,
)

logger = logging.getLogger(__name__)


class UserService:
    """Service class for user business logic and gRPC communication."""

    def __init__(self, grpc_client: UserGRPCClient) -> None:
        """Initialize service with gRPC client."""
        self.grpc_client = grpc_client

    async def create_user(self, user_data: UserCreate) -> UserResponse:
        """Create a new user."""
        try:
            request = CreateUserRequest(
                name=user_data.name,
                email=user_data.email,
                age=user_data.age,
            )
            response = self.grpc_client.stub.CreateUser(request)
            return self._grpc_user_to_pydantic(response.user)
        except grpc.RpcError as e:
            logger.error(f"gRPC error creating user: {e}")
            raise grpc_to_http_exception(e)

    async def get_user(self, user_id: str) -> UserResponse:
        """Get user by ID."""
        try:
            request = GetUserRequest(id=user_id)
            response = self.grpc_client.stub.GetUser(request)
            return self._grpc_user_to_pydantic(response.user)
        except grpc.RpcError as e:
            logger.error(f"gRPC error getting user {user_id}: {e}")
            raise grpc_to_http_exception(e)

    async def update_user(self, user_id: str, user_data: UserUpdate) -> UserResponse:
        """Update an existing user."""
        try:
            request = UpdateUserRequest(
                id=user_id,
                name=user_data.name or "",
                email=user_data.email or "",
                age=user_data.age or 0,
            )
            response = self.grpc_client.stub.UpdateUser(request)
            return self._grpc_user_to_pydantic(response.user)
        except grpc.RpcError as e:
            logger.error(f"gRPC error updating user {user_id}: {e}")
            raise grpc_to_http_exception(e)

    async def delete_user(self, user_id: str) -> MessageResponse:
        """Delete a user by ID."""
        try:
            request = DeleteUserRequest(id=user_id)
            response = self.grpc_client.stub.DeleteUser(request)
            return MessageResponse(message=response.message)
        except grpc.RpcError as e:
            logger.error(f"gRPC error deleting user {user_id}: {e}")
            raise grpc_to_http_exception(e)

    async def list_users(self, page: int = 1, limit: int = 10) -> UserListResponse:
        """List users with pagination."""
        try:
            request = ListUsersRequest(page=page, limit=limit)
            response = self.grpc_client.stub.ListUsers(request)

            users = [self._grpc_user_to_pydantic(user) for user in response.users]
            return UserListResponse(
                users=users,
                total=response.total,
                message=response.message,
            )
        except grpc.RpcError as e:
            logger.error(f"gRPC error listing users: {e}")
            raise grpc_to_http_exception(e)

    def _grpc_user_to_pydantic(self, grpc_user: User) -> UserResponse:
        """Convert gRPC User message to Pydantic UserResponse."""
        return UserResponse(
            id=grpc_user.id,
            name=grpc_user.name,
            email=grpc_user.email,
            age=grpc_user.age,
            created_at=grpc_user.created_at,
            updated_at=grpc_user.updated_at,
        )

