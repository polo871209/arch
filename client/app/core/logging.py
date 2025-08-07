import logging
import sys

from opentelemetry.instrumentation.logging import LoggingInstrumentor

from .config import settings


def setup_logging():
    """Configure unified logging with OpenTelemetry trace context correlation"""
    LoggingInstrumentor().instrument(set_logging_format=True)

    formatter = logging.Formatter(
        "%(asctime)s %(levelname)s %(otelTraceID)s %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
    )

    handler = logging.StreamHandler(sys.stdout)
    handler.setFormatter(formatter)

    loggers = ["", "uvicorn", "uvicorn.error", "fastapi"]
    level = getattr(logging, settings.log_level.upper())

    for logger_name in loggers:
        logger = logging.getLogger(logger_name)
        logger.handlers.clear()
        logger.addHandler(handler)
        logger.setLevel(level)
        if logger_name:
            logger.propagate = False

    logging.getLogger("uvicorn.access").disabled = True


def create_access_log_middleware():
    access_logger = logging.getLogger("app.access")

    async def access_log_middleware(request, call_next):
        response = await call_next(request)

        client = request.client.host if request.client else "unknown"
        access_logger.info(
            f'{client} "{request.method} {request.url.path}" {response.status_code}'
        )

        return response

    return access_log_middleware
