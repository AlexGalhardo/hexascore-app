from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar, cast

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.item_meta import ItemMeta


T = TypeVar("T", bound="EditionJSON")


@_attrs_define
class EditionJSON:
    """
    Attributes:
        id (str):
        format_ (None | str | Unset):
        isbn13 (None | str | Unset):
        language (str | Unset):
        meta (ItemMeta | Unset):
        page_count (int | None | Unset):
        publication_year (int | None | Unset):
        publisher (None | str | Unset):
        subtitle (None | str | Unset):
        title (None | str | Unset):
    """

    id: str
    format_: None | str | Unset = UNSET
    isbn13: None | str | Unset = UNSET
    language: str | Unset = UNSET
    meta: ItemMeta | Unset = UNSET
    page_count: int | None | Unset = UNSET
    publication_year: int | None | Unset = UNSET
    publisher: None | str | Unset = UNSET
    subtitle: None | str | Unset = UNSET
    title: None | str | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        id = self.id

        format_: None | str | Unset
        if isinstance(self.format_, Unset):
            format_ = UNSET
        else:
            format_ = self.format_

        isbn13: None | str | Unset
        if isinstance(self.isbn13, Unset):
            isbn13 = UNSET
        else:
            isbn13 = self.isbn13

        language = self.language

        meta: dict[str, Any] | Unset = UNSET
        if not isinstance(self.meta, Unset):
            meta = self.meta.to_dict()

        page_count: int | None | Unset
        if isinstance(self.page_count, Unset):
            page_count = UNSET
        else:
            page_count = self.page_count

        publication_year: int | None | Unset
        if isinstance(self.publication_year, Unset):
            publication_year = UNSET
        else:
            publication_year = self.publication_year

        publisher: None | str | Unset
        if isinstance(self.publisher, Unset):
            publisher = UNSET
        else:
            publisher = self.publisher

        subtitle: None | str | Unset
        if isinstance(self.subtitle, Unset):
            subtitle = UNSET
        else:
            subtitle = self.subtitle

        title: None | str | Unset
        if isinstance(self.title, Unset):
            title = UNSET
        else:
            title = self.title

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
            }
        )
        if format_ is not UNSET:
            field_dict["format"] = format_
        if isbn13 is not UNSET:
            field_dict["isbn13"] = isbn13
        if language is not UNSET:
            field_dict["language"] = language
        if meta is not UNSET:
            field_dict["meta"] = meta
        if page_count is not UNSET:
            field_dict["pageCount"] = page_count
        if publication_year is not UNSET:
            field_dict["publicationYear"] = publication_year
        if publisher is not UNSET:
            field_dict["publisher"] = publisher
        if subtitle is not UNSET:
            field_dict["subtitle"] = subtitle
        if title is not UNSET:
            field_dict["title"] = title

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.item_meta import ItemMeta  # noqa: PLC0415

        d = dict(src_dict)
        id = d.pop("id")

        def _parse_format_(data: object) -> None | str | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(None | str | Unset, data)

        format_ = _parse_format_(d.pop("format", UNSET))

        def _parse_isbn13(data: object) -> None | str | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(None | str | Unset, data)

        isbn13 = _parse_isbn13(d.pop("isbn13", UNSET))

        language = d.pop("language", UNSET)

        _meta = d.pop("meta", UNSET)
        meta: ItemMeta | Unset
        if isinstance(_meta, Unset):
            meta = UNSET
        else:
            meta = ItemMeta.from_dict(_meta)

        def _parse_page_count(data: object) -> int | None | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(int | None | Unset, data)

        page_count = _parse_page_count(d.pop("pageCount", UNSET))

        def _parse_publication_year(data: object) -> int | None | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(int | None | Unset, data)

        publication_year = _parse_publication_year(d.pop("publicationYear", UNSET))

        def _parse_publisher(data: object) -> None | str | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(None | str | Unset, data)

        publisher = _parse_publisher(d.pop("publisher", UNSET))

        def _parse_subtitle(data: object) -> None | str | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(None | str | Unset, data)

        subtitle = _parse_subtitle(d.pop("subtitle", UNSET))

        def _parse_title(data: object) -> None | str | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(None | str | Unset, data)

        title = _parse_title(d.pop("title", UNSET))

        edition_json = cls(
            id=id,
            format_=format_,
            isbn13=isbn13,
            language=language,
            meta=meta,
            page_count=page_count,
            publication_year=publication_year,
            publisher=publisher,
            subtitle=subtitle,
            title=title,
        )

        edition_json.additional_properties = d
        return edition_json

    @property
    def additional_keys(self) -> list[str]:
        return list(self.additional_properties.keys())

    def __getitem__(self, key: str) -> Any:
        return self.additional_properties[key]

    def __setitem__(self, key: str, value: Any) -> None:
        self.additional_properties[key] = value

    def __delitem__(self, key: str) -> None:
        del self.additional_properties[key]

    def __contains__(self, key: str) -> bool:
        return key in self.additional_properties
