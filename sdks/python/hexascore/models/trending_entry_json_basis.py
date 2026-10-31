from enum import StrEnum


class TrendingEntryJSONBasis(StrEnum):
    PRIOR = "prior"
    TREND = "trend"

    def __str__(self) -> str:
        return str(self.value)
