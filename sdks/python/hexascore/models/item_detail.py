from __future__ import annotations

import datetime
from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar, cast
from uuid import UUID

from attrs import define as _attrs_define

from ..models.item_detail_locale import ItemDetailLocale
from ..models.item_detail_type import ItemDetailType

if TYPE_CHECKING:
    from ..models.item_detail_categories import ItemDetailCategories
    from ..models.item_detail_external_ids import ItemDetailExternalIds
    from ..models.item_detail_trailers_item import ItemDetailTrailersItem
    from ..models.synopsis import Synopsis
    from ..models.time_to_beat import TimeToBeat


T = TypeVar("T", bound="ItemDetail")


@_attrs_define
class ItemDetail:
    """Full item: metadata, scores, external ids, trailers and synopsis.

    Attributes:
        id (UUID): Stable item id.
        type_ (ItemDetailType): Kind of item. List endpoints use the plural path (`/games`, `/movies`, `/series`,
            `/books`).
        slug (str): URL-safe name, unique per type.
        title (str): Title in the requested locale, English when missing.
        release_date (datetime.date | None): First release (YYYY-MM-DD).
        year (int | None):
        score (int | None): Community score 0–100 (mean of the six categories). `null` until the item has 5 ratings.
        categories (ItemDetailCategories): Average per category, 0–100, `null` under 5 ratings. Keys by type — `game`:
            `gameplay`, `fun`, `sound`, `technical`, `visual`, `storytelling`; `movie`: `screenplay`, `acting`, `direction`,
            `cinematography`, `soundtrack`, `entertainment`; `series`: `screenplay`, `acting`, `characters`, `pacing`,
            `production`, `entertainment`; `book`: `plot`, `characters`, `writing`, `pacing`, `originality`, `impact`.
        rating_count (int): Number of community ratings.
        genres (list[str]):
        platforms (list[str]): Platforms (games only; empty for other types).
        cover_url (None | str):
        url (str): Public page of the item on the website, in the requested locale.
        locale (ItemDetailLocale):
        translation_fallback (bool): True when the requested locale had no title or synopsis and English was used.
        synopsis (None | Synopsis):
        external_ids (ItemDetailExternalIds): Identifiers of the item in other databases. Only known ids are present.
        trailers (list[ItemDetailTrailersItem]): YouTube video ids; embed as `https://www.youtube-
            nocookie.com/embed/{youtube_id}`.
        creators (list[str]): Developers, directors, showrunners or authors.
        countries (list[str]):
        runtime_minutes (int | None):
        pages (int | None):
        seasons (int | None):
        episodes (int | None):
        time_to_beat (None | TimeToBeat): Games only; `null` for other types.
    """

    id: UUID
    type_: ItemDetailType
    slug: str
    title: str
    release_date: datetime.date | None
    year: int | None
    score: int | None
    categories: ItemDetailCategories
    rating_count: int
    genres: list[str]
    platforms: list[str]
    cover_url: None | str
    url: str
    locale: ItemDetailLocale
    translation_fallback: bool
    synopsis: None | Synopsis
    external_ids: ItemDetailExternalIds
    trailers: list[ItemDetailTrailersItem]
    creators: list[str]
    countries: list[str]
    runtime_minutes: int | None
    pages: int | None
    seasons: int | None
    episodes: int | None
    time_to_beat: None | TimeToBeat

    def to_dict(self) -> dict[str, Any]:
        from ..models.synopsis import Synopsis  # noqa: PLC0415
        from ..models.time_to_beat import TimeToBeat  # noqa: PLC0415

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

        locale = self.locale.value

        translation_fallback = self.translation_fallback

        synopsis: dict[str, Any] | None
        if isinstance(self.synopsis, Synopsis):
            synopsis = self.synopsis.to_dict()
        else:
            synopsis = self.synopsis

        external_ids = self.external_ids.to_dict()

        trailers = []
        for trailers_item_data in self.trailers:
            trailers_item = trailers_item_data.to_dict()
            trailers.append(trailers_item)

        creators = self.creators

        countries = self.countries

        runtime_minutes: int | None
        runtime_minutes = self.runtime_minutes

        pages: int | None
        pages = self.pages

        seasons: int | None
        seasons = self.seasons

        episodes: int | None
        episodes = self.episodes

        time_to_beat: dict[str, Any] | None
        if isinstance(self.time_to_beat, TimeToBeat):
            time_to_beat = self.time_to_beat.to_dict()
        else:
            time_to_beat = self.time_to_beat

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
                "locale": locale,
                "translation_fallback": translation_fallback,
                "synopsis": synopsis,
                "external_ids": external_ids,
                "trailers": trailers,
                "creators": creators,
                "countries": countries,
                "runtime_minutes": runtime_minutes,
                "pages": pages,
                "seasons": seasons,
                "episodes": episodes,
                "time_to_beat": time_to_beat,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.item_detail_categories import ItemDetailCategories  # noqa: PLC0415
        from ..models.item_detail_external_ids import ItemDetailExternalIds  # noqa: PLC0415
        from ..models.item_detail_trailers_item import ItemDetailTrailersItem  # noqa: PLC0415
        from ..models.synopsis import Synopsis  # noqa: PLC0415
        from ..models.time_to_beat import TimeToBeat  # noqa: PLC0415

        d = dict(src_dict)
        id = UUID(d.pop("id"))

        type_ = ItemDetailType(d.pop("type"))

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

        categories = ItemDetailCategories.from_dict(d.pop("categories"))

        rating_count = d.pop("rating_count")

        genres = cast(list[str], d.pop("genres"))

        platforms = cast(list[str], d.pop("platforms"))

        def _parse_cover_url(data: object) -> None | str:
            if data is None:
                return data
            return cast(None | str, data)

        cover_url = _parse_cover_url(d.pop("cover_url"))

        url = d.pop("url")

        locale = ItemDetailLocale(d.pop("locale"))

        translation_fallback = d.pop("translation_fallback")

        def _parse_synopsis(data: object) -> None | Synopsis:
            if data is None:
                return data
            try:
                if not isinstance(data, dict):
                    raise TypeError()
                synopsis_type_0 = Synopsis.from_dict(data)

                return synopsis_type_0
            except (TypeError, ValueError, AttributeError, KeyError):
                pass
            return cast(None | Synopsis, data)

        synopsis = _parse_synopsis(d.pop("synopsis"))

        external_ids = ItemDetailExternalIds.from_dict(d.pop("external_ids"))

        trailers = []
        _trailers = d.pop("trailers")
        for trailers_item_data in _trailers:
            trailers_item = ItemDetailTrailersItem.from_dict(trailers_item_data)

            trailers.append(trailers_item)

        creators = cast(list[str], d.pop("creators"))

        countries = cast(list[str], d.pop("countries"))

        def _parse_runtime_minutes(data: object) -> int | None:
            if data is None:
                return data
            return cast(int | None, data)

        runtime_minutes = _parse_runtime_minutes(d.pop("runtime_minutes"))

        def _parse_pages(data: object) -> int | None:
            if data is None:
                return data
            return cast(int | None, data)

        pages = _parse_pages(d.pop("pages"))

        def _parse_seasons(data: object) -> int | None:
            if data is None:
                return data
            return cast(int | None, data)

        seasons = _parse_seasons(d.pop("seasons"))

        def _parse_episodes(data: object) -> int | None:
            if data is None:
                return data
            return cast(int | None, data)

        episodes = _parse_episodes(d.pop("episodes"))

        def _parse_time_to_beat(data: object) -> None | TimeToBeat:
            if data is None:
                return data
            try:
                if not isinstance(data, dict):
                    raise TypeError()
                time_to_beat_type_0 = TimeToBeat.from_dict(data)

                return time_to_beat_type_0
            except (TypeError, ValueError, AttributeError, KeyError):
                pass
            return cast(None | TimeToBeat, data)

        time_to_beat = _parse_time_to_beat(d.pop("time_to_beat"))

        item_detail = cls(
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
            locale=locale,
            translation_fallback=translation_fallback,
            synopsis=synopsis,
            external_ids=external_ids,
            trailers=trailers,
            creators=creators,
            countries=countries,
            runtime_minutes=runtime_minutes,
            pages=pages,
            seasons=seasons,
            episodes=episodes,
            time_to_beat=time_to_beat,
        )

        return item_detail
