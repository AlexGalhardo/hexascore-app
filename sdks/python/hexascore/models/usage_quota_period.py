from enum import StrEnum


class UsageQuotaPeriod(StrEnum):
    LIFETIME = "lifetime"
    MONTH = "month"

    def __str__(self) -> str:
        return str(self.value)
