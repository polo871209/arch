from typing import Dict, Optional

from fastapi import Request
from opentelemetry import propagate
from opentelemetry.propagators.b3 import B3MultiFormat

# https://github.com/istio/istio/blob/release-1.26/samples/bookinfo/src/productpage/productpage.py

# Initialize OpenTelemetry B3 propagator for Istio compatibility
propagator = B3MultiFormat()
propagate.set_global_textmap(propagator)


def get_forward_headers(request: Request) -> Dict[str, str]:
    """
    Extract and propagate tracing headers from FastAPI request.

    This function extracts trace context from incoming headers and creates
    a dictionary of headers to forward to downstream services, following
    the Istio best practices for distributed tracing.
    """
    headers = {}

    # Extract B3 trace context and inject into headers dict
    # Convert request headers to lowercase dict for case-insensitive lookup
    request_headers = {k.lower(): v for k, v in request.headers.items()}
    ctx = propagator.extract(carrier=request_headers)
    propagator.inject(headers, ctx)

    # Additional headers that need manual propagation for Istio
    # Keep this in sync with Istio documentation
    additional_headers = [
        # All applications should propagate x-request-id
        "x-request-id",
        # W3C Trace Context (compatible with OpenCensus and Stackdriver)
        "traceparent",
        "tracestate",
        # Cloud trace context
        "x-cloud-trace-context",
    ]

    # Propagate additional headers
    for header_name in additional_headers:
        header_value = request.headers.get(header_name)
        if header_value:
            headers[header_name] = header_value

    return headers


def create_grpc_metadata(headers: Optional[Dict[str, str]] = None) -> list:
    """Create gRPC metadata tuples from tracing headers."""
    if not headers:
        return []

    return [(key, value) for key, value in headers.items()]
