import logging
from collections.abc import AsyncIterator
from contextlib import asynccontextmanager

from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse
from opentelemetry import trace
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.sdk.trace import TracerProvider

from .api import health_router, users_router
from .core.config import settings
from .grpc_client.client import AsyncUserGRPCClient

logger = logging.getLogger(__name__)

trace.set_tracer_provider(TracerProvider())


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:
    logger.info("Starting up FastAPI app...")
    grpc_client = AsyncUserGRPCClient()
    await grpc_client.connect()
    app.state.grpc_client = grpc_client
    logger.info("gRPC client connected.")

    try:
        yield
    finally:
        # Shutdown
        logger.info("Shutting down FastAPI app...")
        await grpc_client.close()
        logger.info("gRPC client connection closed.")


def create_app() -> FastAPI:
    app = FastAPI(
        title=settings.app_name,
        description="FastAPI client for User gRPC service",
        version=settings.app_version,
        lifespan=lifespan,
        docs_url="/docs",
        redoc_url="/redoc",
    )
    FastAPIInstrumentor.instrument_app(app)

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
