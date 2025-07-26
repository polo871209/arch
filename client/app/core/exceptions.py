"""Custom exception classes for the application."""

from fastapi import HTTPException


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
    import grpc

    if not isinstance(grpc_error, grpc.RpcError):
        return HTTPException(status_code=500, detail="Internal server error")

    status_code = grpc_error.code()
    detail = grpc_error.details()

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
