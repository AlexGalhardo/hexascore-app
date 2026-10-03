from enum import StrEnum


class UsagePlan(StrEnum):
    BUSINESS = "business"
    FREE = "free"
    PRO = "pro"

    def __str__(self) -> str:
        return str(self.value)
