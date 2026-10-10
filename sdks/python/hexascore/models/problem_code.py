from enum import StrEnum


class ProblemCode(StrEnum):
    BAD_REQUEST = "bad_request"
    FORBIDDEN = "forbidden"
    INTERNAL_ERROR = "internal_error"
    NOT_FOUND = "not_found"
    PAYMENT_REQUIRED = "payment_required"
    RATE_LIMITED = "rate_limited"
    SERVICE_UNAVAILABLE = "service_unavailable"
    UNAUTHORIZED = "unauthorized"

    def __str__(self) -> str:
        return str(self.value)
