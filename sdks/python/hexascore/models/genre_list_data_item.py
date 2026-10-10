from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar

from attrs import define as _attrs_define

T = TypeVar("T", bound="GenreListDataItem")


@_attrs_define
class GenreListDataItem:
    """
    Attributes:
        slug (str): Exact value to pass as `genre`.
        count (int): Items of this type with the genre.
    """

    slug: str
    count: int

    def to_dict(self) -> dict[str, Any]:
        slug = self.slug

        count = self.count

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "slug": slug,
                "count": count,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        slug = d.pop("slug")

        count = d.pop("count")

        genre_list_data_item = cls(
            slug=slug,
            count=count,
        )

        return genre_list_data_item
