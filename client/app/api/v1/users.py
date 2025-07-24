"""User API endpoints version 1."""

from typing import Annotated

from fastapi import APIRouter, Depends, Query

from ...models import UserCreate, UserUpdate, UserResponse, UserListResponse, MessageResponse
from ...services import UserService
from ...grpc_client import UserGRPCClient


router = APIRouter(tags=["users"])


def get_user_service() -> UserService:
    """Dependency to get UserService instance."""
    grpc_client = UserGRPCClient()
    return UserService(grpc_client)


@router.post(
    "/users",
    response_model=UserResponse,
    status_code=201,
    summary="Create a new user",
    description="Create a new user with name, email, and age",
)
async def create_user(
    user: UserCreate,
    user_service: Annotated[UserService, Depends(get_user_service)]
) -> UserResponse:
    """Create a new user."""
    return await user_service.create_user(user)


@router.get(
    "/users/{user_id}",
    response_model=UserResponse,
    summary="Get user by ID",
    description="Retrieve a user by their unique identifier",
)
async def get_user(
    user_id: str,
    user_service: Annotated[UserService, Depends(get_user_service)]
) -> UserResponse:
    """Get a user by ID."""
    return await user_service.get_user(user_id)


@router.put(
    "/users/{user_id}",
    response_model=UserResponse,
    summary="Update user",
    description="Update an existing user's information",
)
async def update_user(
    user_id: str,
    user: UserUpdate,
    user_service: Annotated[UserService, Depends(get_user_service)]
) -> UserResponse:
    """Update a user."""
    return await user_service.update_user(user_id, user)


@router.delete(
    "/users/{user_id}",
    response_model=MessageResponse,
    summary="Delete user",
    description="Delete a user by their unique identifier",
)
async def delete_user(
    user_id: str,
    user_service: Annotated[UserService, Depends(get_user_service)]
) -> MessageResponse:
    """Delete a user."""
    return await user_service.delete_user(user_id)


@router.get(
    "/users",
    response_model=UserListResponse,
    summary="List users",
    description="List users with pagination support",
)
async def list_users(
    user_service: Annotated[UserService, Depends(get_user_service)],
    page: Annotated[int, Query(ge=1, description="Page number")] = 1,
    limit: Annotated[int, Query(ge=1, le=100, description="Items per page")] = 10,
) -> UserListResponse:
    """List users with pagination."""
    return await user_service.list_users(page, limit)