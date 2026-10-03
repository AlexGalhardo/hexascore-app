"""Contains all the data models used in inputs/outputs"""

from .genre_list import GenreList
from .genre_list_data_item import GenreListDataItem
from .get_book_locale import GetBookLocale
from .get_game_locale import GetGameLocale
from .get_movie_locale import GetMovieLocale
from .get_series_locale import GetSeriesLocale
from .item_detail import ItemDetail
from .item_detail_categories import ItemDetailCategories
from .item_detail_external_ids import ItemDetailExternalIds
from .item_detail_locale import ItemDetailLocale
from .item_detail_trailers_item import ItemDetailTrailersItem
from .item_detail_type import ItemDetailType
from .item_list import ItemList
from .item_summary import ItemSummary
from .item_summary_categories import ItemSummaryCategories
from .item_summary_type import ItemSummaryType
from .list_books_locale import ListBooksLocale
from .list_books_sort import ListBooksSort
from .list_games_locale import ListGamesLocale
from .list_games_sort import ListGamesSort
from .list_movies_locale import ListMoviesLocale
from .list_movies_sort import ListMoviesSort
from .list_series_locale import ListSeriesLocale
from .list_series_sort import ListSeriesSort
from .lookup_item_locale import LookupItemLocale
from .lookup_item_source import LookupItemSource
from .problem import Problem
from .problem_code import ProblemCode
from .problem_errors_item import ProblemErrorsItem
from .problem_plan import ProblemPlan
from .search_items_locale import SearchItemsLocale
from .search_items_type import SearchItemsType
from .search_result import SearchResult
from .synopsis import Synopsis
from .synopsis_locale import SynopsisLocale
from .time_to_beat import TimeToBeat
from .time_to_beat_completionist import TimeToBeatCompletionist
from .time_to_beat_extras import TimeToBeatExtras
from .time_to_beat_main import TimeToBeatMain
from .usage import Usage
from .usage_credits import UsageCredits
from .usage_plan import UsagePlan
from .usage_quota import UsageQuota
from .usage_quota_period import UsageQuotaPeriod
from .usage_rate_limit import UsageRateLimit

__all__ = (
    "GenreList",
    "GenreListDataItem",
    "GetBookLocale",
    "GetGameLocale",
    "GetMovieLocale",
    "GetSeriesLocale",
    "ItemDetail",
    "ItemDetailCategories",
    "ItemDetailExternalIds",
    "ItemDetailLocale",
    "ItemDetailTrailersItem",
    "ItemDetailType",
    "ItemList",
    "ItemSummary",
    "ItemSummaryCategories",
    "ItemSummaryType",
    "ListBooksLocale",
    "ListBooksSort",
    "ListGamesLocale",
    "ListGamesSort",
    "ListMoviesLocale",
    "ListMoviesSort",
    "ListSeriesLocale",
    "ListSeriesSort",
    "LookupItemLocale",
    "LookupItemSource",
    "Problem",
    "ProblemCode",
    "ProblemErrorsItem",
    "ProblemPlan",
    "SearchItemsLocale",
    "SearchItemsType",
    "SearchResult",
    "Synopsis",
    "SynopsisLocale",
    "TimeToBeat",
    "TimeToBeatCompletionist",
    "TimeToBeatExtras",
    "TimeToBeatMain",
    "Usage",
    "UsageCredits",
    "UsagePlan",
    "UsageQuota",
    "UsageQuotaPeriod",
    "UsageRateLimit",
)
