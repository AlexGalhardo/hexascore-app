from enum import StrEnum


class GetExportType(StrEnum):
    BOOKS = "books"
    GAMES = "games"
    MOVIES = "movies"
    SERIES = "series"

    def __str__(self) -> str:
        return str(self.value)
