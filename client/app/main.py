"""Main FastAPI application module."""

import logging
from collections.abc import AsyncIterator
from contextlib import asynccontextmanager

from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse

from .api import health_router, users_router
from .core.config import settings

# Configure logging
logging.basicConfig(level=logging.DEBUG)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:
    """Manage application lifespan with proper resource cleanup."""
    # Startup
    yield
    # Shutdown - cleanup resources if needed


def create_app() -> FastAPI:
    """Create and configure FastAPI application."""
    app = FastAPI(
        title=settings.app_name,
        description="Modern FastAPI client for User gRPC service using Python 3.13",
        version=settings.app_version,
        lifespan=lifespan,
        docs_url="/docs",
        redoc_url="/redoc",
    )

    # Global exception handler
    @app.exception_handler(Exception)
    async def global_exception_handler(request: Request, exc: Exception) -> JSONResponse:
        """Handle all unhandled exceptions."""
        logger.error(f"Unhandled exception on {request.method} {request.url}: {exc}", exc_info=True)
        
        # Print to stdout so Docker can capture it
        print(f"ERROR: {request.method} {request.url} - {type(exc).__name__}: {exc}")
        import traceback
        traceback.print_exc()
        
        return JSONResponse(
            status_code=500,
            content={"detail": f"Internal server error: {str(exc)}"}
        )

    # Include routers
    app.include_router(health_router)
    app.include_router(users_router, prefix="/v1")

    return app


# Create app instance
app = create_app()


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(
        "app.main:app",
        host=settings.api_host,
        port=settings.api_port,
        reload=settings.api_reload,
        log_level=settings.log_level.lower(),
    )
