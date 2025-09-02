from pydantic import BaseModel


class HealthResponse(BaseModel):
    status: str
    service: str
    version: str
    grpc_user_status: str
