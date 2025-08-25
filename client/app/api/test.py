"""Test API endpoints."""

from typing import Annotated

from fastapi import APIRouter, Depends, Path, Request

from ..grpc_client import AsyncUserGRPCClient
from ..models import MessageResponse
from ..services import TestService

router = APIRouter(tags=["test"])


def get_test_service(request: Request) -> TestService:
    grpc_client: AsyncUserGRPCClient = request.app.state.grpc_client
    return TestService(grpc_client)


@router.get(
    "/test-error/{status_code}",
    response_model=MessageResponse,
    summary="Test error endpoint",
    description="Test endpoint that returns the specified HTTP status code with trace_id",
    responses={
        400: {"description": "Bad Request"},
        401: {"description": "Unauthorized"},
        403: {"description": "Forbidden"},
        404: {"description": "Not Found"},
        409: {"description": "Conflict"},
        429: {"description": "Too Many Requests"},
        500: {"description": "Internal Server Error"},
        501: {"description": "Not Implemented"},
        503: {"description": "Service Unavailable"},
        504: {"description": "Gateway Timeout"},
    },
)
async def test_error(
    status_code: Annotated[
        str, Path(description="HTTP status code to return (e.g., '404', '500')")
    ],
    test_service: Annotated[TestService, Depends(get_test_service)],
) -> MessageResponse:
    """Test endpoint that returns the specified HTTP status code."""
    return await test_service.test_error(status_code)
