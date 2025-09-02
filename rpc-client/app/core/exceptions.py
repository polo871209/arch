from typing import Protocol

from fastapi import HTTPException


class GRPCError(Protocol):
    """Protocol for gRPC errors that have code and details methods."""

    def code(self) -> int: ...
    def details(self) -> str: ...


class GRPCClientError(Exception):
    """Base exception for gRPC client errors."""

    pass


class GRPCServiceUnavailableError(GRPCClientError):
    """Raised when gRPC service is unavailable."""

    pass


class GRPCTimeoutError(GRPCClientError):
    """Raised when gRPC request times out."""

    pass


def grpc_to_http_exception(grpc_error: Exception) -> HTTPException:
    """Convert gRPC errors to HTTP exceptions."""
    from typing import cast

    import grpc

    if not isinstance(grpc_error, grpc.RpcError):
        return HTTPException(status_code=500, detail="Internal server error")

    # Check if the error has the required methods (some RpcError subclasses don't)
    if not (hasattr(grpc_error, "code") and hasattr(grpc_error, "details")):
        return HTTPException(status_code=500, detail="Internal server error")

    # Cast to our protocol to satisfy type checker
    typed_error = cast(GRPCError, grpc_error)
    status_code = typed_error.code()
    detail = typed_error.details()

    match status_code:
        case grpc.StatusCode.NOT_FOUND:
            return HTTPException(status_code=404, detail=detail)
        case grpc.StatusCode.ALREADY_EXISTS:
            return HTTPException(status_code=409, detail=detail)
        case grpc.StatusCode.INVALID_ARGUMENT:
            return HTTPException(status_code=400, detail=detail)
        case grpc.StatusCode.UNAVAILABLE:
            return HTTPException(status_code=503, detail="gRPC service unavailable")
        case grpc.StatusCode.DEADLINE_EXCEEDED:
            return HTTPException(status_code=504, detail="Request timeout")
        case grpc.StatusCode.PERMISSION_DENIED:
            return HTTPException(status_code=403, detail="Permission denied")
        case grpc.StatusCode.UNAUTHENTICATED:
            return HTTPException(status_code=401, detail="Authentication required")
        case _:
            return HTTPException(status_code=500, detail="Internal server error")
