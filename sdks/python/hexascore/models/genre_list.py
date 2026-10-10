from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar

from attrs import define as _attrs_define

if TYPE_CHECKING:
    from ..models.genre_list_data_item import GenreListDataItem


T = TypeVar("T", bound="GenreList")


@_attrs_define
class GenreList:
    """Genres of a type, most common first.

    Attributes:
        data (list[GenreListDataItem]):
    """

    data: list[GenreListDataItem]

    def to_dict(self) -> dict[str, Any]:
        data = []
        for data_item_data in self.data:
            data_item = data_item_data.to_dict()
            data.append(data_item)

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "data": data,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.genre_list_data_item import GenreListDataItem  # noqa: PLC0415

        d = dict(src_dict)
        data = []
        _data = d.pop("data")
        for data_item_data in _data:
            data_item = GenreListDataItem.from_dict(data_item_data)

            data.append(data_item)

        genre_list = cls(
            data=data,
        )

        return genre_list
