from typing import List, Optional
from google.protobuf.message import Message

class User(Message):
    id: str
    name: str
    email: str
    age: int
    created_at: int
    updated_at: int
    
    def __init__(
        self,
        *,
        id: str = ...,
        name: str = ...,
        email: str = ...,
        age: int = ...,
        created_at: int = ...,
        updated_at: int = ...,
    ) -> None: ...

class CreateUserRequest(Message):
    name: str
    email: str
    age: int
    
    def __init__(
        self,
        *,
        name: str = ...,
        email: str = ...,
        age: int = ...,
    ) -> None: ...

class CreateUserResponse(Message):
    user: User
    message: str
    
    def __init__(
        self,
        *,
        user: Optional[User] = ...,
        message: str = ...,
    ) -> None: ...

class GetUserRequest(Message):
    id: str
    
    def __init__(
        self,
        *,
        id: str = ...,
    ) -> None: ...

class GetUserResponse(Message):
    user: User
    message: str
    
    def __init__(
        self,
        *,
        user: Optional[User] = ...,
        message: str = ...,
    ) -> None: ...

class UpdateUserRequest(Message):
    id: str
    name: str
    email: str
    age: int
    
    def __init__(
        self,
        *,
        id: str = ...,
        name: str = ...,
        email: str = ...,
        age: int = ...,
    ) -> None: ...

class UpdateUserResponse(Message):
    user: User
    message: str
    
    def __init__(
        self,
        *,
        user: Optional[User] = ...,
        message: str = ...,
    ) -> None: ...

class DeleteUserRequest(Message):
    id: str
    
    def __init__(
        self,
        *,
        id: str = ...,
    ) -> None: ...

class DeleteUserResponse(Message):
    message: str
    
    def __init__(
        self,
        *,
        message: str = ...,
    ) -> None: ...

class ListUsersRequest(Message):
    page: int
    limit: int
    
    def __init__(
        self,
        *,
        page: int = ...,
        limit: int = ...,
    ) -> None: ...

class ListUsersResponse(Message):
    users: List[User]
    total: int
    message: str
    
    def __init__(
        self,
        *,
        users: Optional[List[User]] = ...,
        total: int = ...,
        message: str = ...,
    ) -> None: ...