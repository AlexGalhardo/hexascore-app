from enum import StrEnum


class GetBookLocale(StrEnum):
    EN = "en"
    ES = "es"
    PT_BR = "pt-BR"

    def __str__(self) -> str:
        return str(self.value)
