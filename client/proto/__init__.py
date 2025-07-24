"""Protobuf generated modules for gRPC communication."""

# Re-export protobuf classes for easier imports
from .user_pb2 import (
    CreateUserRequest,
    CreateUserResponse,
    DeleteUserRequest,
    DeleteUserResponse,
    GetUserRequest,
    GetUserResponse,
    ListUsersRequest,
    ListUsersResponse,
    UpdateUserRequest,
    UpdateUserResponse,
    User,
)
from .user_pb2_grpc import UserServiceStub

__all__ = [
    "CreateUserRequest",
    "CreateUserResponse",
    "DeleteUserRequest", 
    "DeleteUserResponse",
    "GetUserRequest",
    "GetUserResponse",
    "ListUsersRequest",
    "ListUsersResponse",
    "UpdateUserRequest",
    "UpdateUserResponse",
    "User",
    "UserServiceStub",
]