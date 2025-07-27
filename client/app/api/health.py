from fastapi import APIRouter, Request

from ..core.config import settings
from ..models import HealthResponse

router = APIRouter(tags=["health"])


@router.get(
    "/health",
    response_model=HealthResponse,
    summary="Health check",
    description="Check the health status of the API service",
)
async def health_check(request: Request) -> HealthResponse:
    grpc_client = request.app.state.grpc_client
    grpc_user_status = "healthy" if await grpc_client.health_check() else "unhealthy"

    return HealthResponse(
        status="healthy",
        service=settings.app_name,
        version=settings.app_version,
        grpc_user_status=grpc_user_status,
    )
