from __future__ import annotations

import datetime
from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar, cast
from uuid import UUID

from attrs import define as _attrs_define

from ..models.item_summary_type import ItemSummaryType

if TYPE_CHECKING:
    from ..models.item_summary_categories import ItemSummaryCategories


T = TypeVar("T", bound="ItemSummary")


@_attrs_define
class ItemSummary:
    """Compact item used in lists and search results.

    Attributes:
        id (UUID): Stable item id.
        type_ (ItemSummaryType): Kind of item. List endpoints use the plural path (`/games`, `/movies`, `/series`,
            `/books`).
        slug (str): URL-safe name, unique per type.
        title (str): Title in the requested locale, English when missing.
        release_date (datetime.date | None): First release (YYYY-MM-DD).
        year (int | None):
        score (int | None): Community score 0–100 (mean of the six categories). `null` until the item has 5 ratings.
        categories (ItemSummaryCategories): Average per category, 0–100, `null` under 5 ratings. Keys by type — `game`:
            `gameplay`, `fun`, `sound`, `technical`, `visual`, `storytelling`; `movie`: `screenplay`, `acting`, `direction`,
            `cinematography`, `soundtrack`, `entertainment`; `series`: `screenplay`, `acting`, `characters`, `pacing`,
            `production`, `entertainment`; `book`: `plot`, `characters`, `writing`, `pacing`, `originality`, `impact`.
        rating_count (int): Number of community ratings.
        genres (list[str]):
        platforms (list[str]): Platforms (games only; empty for other types).
        cover_url (None | str):
        url (str): Public page of the item on the website, in the requested locale.
    """

    id: UUID
    type_: ItemSummaryType
    slug: str
    title: str
    release_date: datetime.date | None
    year: int | None
    score: int | None
    categories: ItemSummaryCategories
    rating_count: int
    genres: list[str]
    platforms: list[str]
    cover_url: None | str
    url: str

    def to_dict(self) -> dict[str, Any]:
        id = str(self.id)

        type_ = self.type_.value

        slug = self.slug

        title = self.title

        release_date: None | str
        if isinstance(self.release_date, datetime.date):
            release_date = self.release_date.isoformat()
        else:
            release_date = self.release_date

        year: int | None
        year = self.year

        score: int | None
        score = self.score

        categories = self.categories.to_dict()

        rating_count = self.rating_count

        genres = self.genres

        platforms = self.platforms

        cover_url: None | str
        cover_url = self.cover_url

        url = self.url

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "id": id,
                "type": type_,
                "slug": slug,
                "title": title,
                "release_date": release_date,
                "year": year,
                "score": score,
                "categories": categories,
                "rating_count": rating_count,
                "genres": genres,
                "platforms": platforms,
                "cover_url": cover_url,
                "url": url,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.item_summary_categories import ItemSummaryCategories  # noqa: PLC0415

        d = dict(src_dict)
        id = UUID(d.pop("id"))

        type_ = ItemSummaryType(d.pop("type"))

        slug = d.pop("slug")

        title = d.pop("title")

        def _parse_release_date(data: object) -> datetime.date | None:
            if data is None:
                return data
            try:
                if not isinstance(data, str):
                    raise TypeError()
                release_date_type_0 = datetime.date.fromisoformat(data)

                return release_date_type_0
            except (TypeError, ValueError, AttributeError, KeyError):
                pass
            return cast(datetime.date | None, data)

        release_date = _parse_release_date(d.pop("release_date"))

        def _parse_year(data: object) -> int | None:
            if data is None:
                return data
            return cast(int | None, data)

        year = _parse_year(d.pop("year"))

        def _parse_score(data: object) -> int | None:
            if data is None:
                return data
            return cast(int | None, data)

        score = _parse_score(d.pop("score"))

        categories = ItemSummaryCategories.from_dict(d.pop("categories"))

        rating_count = d.pop("rating_count")

        genres = cast(list[str], d.pop("genres"))

        platforms = cast(list[str], d.pop("platforms"))

        def _parse_cover_url(data: object) -> None | str:
            if data is None:
                return data
            return cast(None | str, data)

        cover_url = _parse_cover_url(d.pop("cover_url"))

        url = d.pop("url")

        item_summary = cls(
            id=id,
            type_=type_,
            slug=slug,
            title=title,
            release_date=release_date,
            year=year,
            score=score,
            categories=categories,
            rating_count=rating_count,
            genres=genres,
            platforms=platforms,
            cover_url=cover_url,
            url=url,
        )

        return item_summary
