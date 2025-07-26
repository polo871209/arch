"""Pydantic models for user data validation and serialization."""

from pydantic import BaseModel, EmailStr, Field


class UserBase(BaseModel):
    """Base user model with common fields."""

    name: str = Field(..., min_length=1, max_length=100, description="User's full name")
    email: EmailStr = Field(..., description="User's email address")
    age: int = Field(..., ge=0, le=150, description="User's age")


class UserCreate(UserBase):
    """Model for creating a new user."""

    pass


class UserUpdate(BaseModel):
    """Model for updating an existing user (all fields optional)."""

    name: str | None = Field(
        None, min_length=1, max_length=100, description="User's full name"
    )
    email: EmailStr | None = Field(None, description="User's email address")
    age: int | None = Field(None, ge=0, le=150, description="User's age")


class UserResponse(UserBase):
    """Model for user response data."""

    id: str = Field(..., description="Unique user identifier")
    created_at: int = Field(..., description="Creation timestamp")
    updated_at: int = Field(..., description="Last update timestamp")

    model_config = {"from_attributes": True}


class UserListResponse(BaseModel):
    """Model for paginated user list response."""

    users: list[UserResponse] = Field(..., description="List of users")
    total: int = Field(..., ge=0, description="Total number of users")
    message: str = Field(..., description="Response message")


class MessageResponse(BaseModel):
    """Model for simple message responses."""

    message: str = Field(..., description="Response message")


class HealthResponse(BaseModel):
    """Model for health check response."""

    status: str = Field(..., description="Service health status")
    service: str = Field(..., description="Service name")
    version: str = Field(..., description="Service version")
