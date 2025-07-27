import logging
from collections.abc import AsyncIterator
from contextlib import asynccontextmanager

from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse

from .api import health_router, users_router
from .core.config import settings

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:
    yield  # Add startup/shutdown logic if needed


def create_app() -> FastAPI:
    app = FastAPI(
        title=settings.app_name,
        description="FastAPI client for User gRPC service",
        version=settings.app_version,
        lifespan=lifespan,
        docs_url="/docs",
        redoc_url="/redoc",
    )

    @app.exception_handler(Exception)
    async def global_exception_handler(
        request: Request, exc: Exception
    ) -> JSONResponse:
        logger.error(
            f"Unhandled exception on {request.method} {request.url}: {exc}",
            exc_info=True,
        )
        return JSONResponse(
            status_code=500, content={"detail": f"Internal server error: {str(exc)}"}
        )

    app.include_router(health_router)
    app.include_router(users_router, prefix="/v1")

    return app


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
