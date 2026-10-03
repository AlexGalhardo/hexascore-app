from enum import StrEnum


class LookupItemSource(StrEnum):
    IGDB = "igdb"
    IMDB = "imdb"
    ISBN13 = "isbn13"
    OPENLIBRARY = "openlibrary"
    STEAM = "steam"
    TMDB = "tmdb"
    WIKIDATA = "wikidata"

    def __str__(self) -> str:
        return str(self.value)
