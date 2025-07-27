"""Health check endpoints."""

from fastapi import APIRouter

from ..core.config import settings
from ..grpc_client import UserGRPCClient
from ..models import HealthResponse

router = APIRouter(tags=["health"])


@router.get(
    "/health",
    response_model=HealthResponse,
    summary="Health check",
    description="Check the health status of the API service",
)
async def health_check() -> HealthResponse:
    user_grpc_client = UserGRPCClient()
    grpc_user_status = "unhealthy"

    if user_grpc_client.health_check():
        grpc_user_status = "healthy"

    return HealthResponse(
        status="healthy",
        service=settings.app_name,
        version=settings.app_version,
        grpc_user_status=grpc_user_status,
    )
