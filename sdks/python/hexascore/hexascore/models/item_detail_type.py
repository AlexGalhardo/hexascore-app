from enum import StrEnum


class ItemDetailType(StrEnum):
    BOOK = "book"
    GAME = "game"
    MOVIE = "movie"
    SERIES = "series"

    def __str__(self) -> str:
        return str(self.value)
