from enum import StrEnum


class ProblemPlan(StrEnum):
    BUSINESS = "business"
    FREE = "free"
    PRO = "pro"

    def __str__(self) -> str:
        return str(self.value)
