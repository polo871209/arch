import grpc
from typing import Optional, Sequence, Tuple
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
)

# Type aliases for gRPC types
Metadata = Sequence[Tuple[str, str]]

class UserServiceStub:
    def __init__(self, channel: grpc.Channel) -> None: ...
    
    def CreateUser(
        self, 
        request: CreateUserRequest,
        timeout: Optional[float] = ...,
        metadata: Optional[Metadata] = ...,
        credentials: Optional[grpc.CallCredentials] = ...,
        wait_for_ready: Optional[bool] = ...,
        compression: Optional[grpc.Compression] = ...,
    ) -> CreateUserResponse: ...
    
    def GetUser(
        self, 
        request: GetUserRequest,
        timeout: Optional[float] = ...,
        metadata: Optional[Metadata] = ...,
        credentials: Optional[grpc.CallCredentials] = ...,  
        wait_for_ready: Optional[bool] = ...,
        compression: Optional[grpc.Compression] = ...,
    ) -> GetUserResponse: ...
    
    def UpdateUser(
        self, 
        request: UpdateUserRequest,
        timeout: Optional[float] = ...,
        metadata: Optional[Metadata] = ...,
        credentials: Optional[grpc.CallCredentials] = ...,
        wait_for_ready: Optional[bool] = ...,
        compression: Optional[grpc.Compression] = ...,
    ) -> UpdateUserResponse: ...
    
    def DeleteUser(
        self, 
        request: DeleteUserRequest,
        timeout: Optional[float] = ...,
        metadata: Optional[Metadata] = ...,
        credentials: Optional[grpc.CallCredentials] = ...,
        wait_for_ready: Optional[bool] = ...,
        compression: Optional[grpc.Compression] = ...,
    ) -> DeleteUserResponse: ...
    
    def ListUsers(
        self, 
        request: ListUsersRequest,
        timeout: Optional[float] = ...,
        metadata: Optional[Metadata] = ...,
        credentials: Optional[grpc.CallCredentials] = ...,
        wait_for_ready: Optional[bool] = ...,
        compression: Optional[grpc.Compression] = ...,
    ) -> ListUsersResponse: ...