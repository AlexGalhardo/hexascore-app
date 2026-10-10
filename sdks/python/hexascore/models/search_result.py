from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar

from attrs import define as _attrs_define

if TYPE_CHECKING:
    from ..models.item_summary import ItemSummary


T = TypeVar("T", bound="SearchResult")


@_attrs_define
class SearchResult:
    """Best matches first. Search is not paginated.

    Attributes:
        data (list[ItemSummary]):
    """

    data: list[ItemSummary]

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
        from ..models.item_summary import ItemSummary  # noqa: PLC0415

        d = dict(src_dict)
        data = []
        _data = d.pop("data")
        for data_item_data in _data:
            data_item = ItemSummary.from_dict(data_item_data)

            data.append(data_item)

        search_result = cls(
            data=data,
        )

        return search_result
