"""Type stubs for protobuf generated modules."""

from .user_pb2 import (
    CreateUserRequest as CreateUserRequest,
    CreateUserResponse as CreateUserResponse,
    DeleteUserRequest as DeleteUserRequest,
    DeleteUserResponse as DeleteUserResponse,
    GetUserRequest as GetUserRequest,
    GetUserResponse as GetUserResponse,
    ListUsersRequest as ListUsersRequest,
    ListUsersResponse as ListUsersResponse,
    UpdateUserRequest as UpdateUserRequest,
    UpdateUserResponse as UpdateUserResponse,
    User as User,
)
from .user_pb2_grpc import UserServiceStub as UserServiceStub

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