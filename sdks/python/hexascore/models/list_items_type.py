from enum import StrEnum


class ListItemsType(StrEnum):
    BOOKS = "books"
    GAMES = "games"
    MOVIES = "movies"
    SERIES = "series"

    def __str__(self) -> str:
        return str(self.value)
