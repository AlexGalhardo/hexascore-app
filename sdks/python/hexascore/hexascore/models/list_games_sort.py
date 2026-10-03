from enum import StrEnum


class ListGamesSort(StrEnum):
    POPULARITY = "popularity"
    POPULARITY_DESC = "-popularity"
    RATING_COUNT = "rating_count"
    RATING_COUNT_DESC = "-rating_count"
    RELEASE_DATE = "release_date"
    RELEASE_DATE_DESC = "-release_date"
    SCORE = "score"
    SCORE_DESC = "-score"

    def __str__(self) -> str:
        return str(self.value)
