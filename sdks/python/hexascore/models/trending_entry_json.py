from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..models.trending_entry_json_basis import TrendingEntryJSONBasis
from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.item_json import ItemJSON


T = TypeVar("T", bound="TrendingEntryJSON")


@_attrs_define
class TrendingEntryJSON:
    """
    Attributes:
        basis (TrendingEntryJSONBasis):
        rank (int):
        item (ItemJSON | Unset):
        trend_score (float | Unset):
    """

    basis: TrendingEntryJSONBasis
    rank: int
    item: ItemJSON | Unset = UNSET
    trend_score: float | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        basis = self.basis.value

        rank = self.rank

        item: dict[str, Any] | Unset = UNSET
        if not isinstance(self.item, Unset):
            item = self.item.to_dict()

        trend_score = self.trend_score

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "basis": basis,
                "rank": rank,
            }
        )
        if item is not UNSET:
            field_dict["item"] = item
        if trend_score is not UNSET:
            field_dict["trendScore"] = trend_score

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.item_json import ItemJSON  # noqa: PLC0415

        d = dict(src_dict)
        basis = TrendingEntryJSONBasis(d.pop("basis"))

        rank = d.pop("rank")

        _item = d.pop("item", UNSET)
        item: ItemJSON | Unset
        if isinstance(_item, Unset):
            item = UNSET
        else:
            item = ItemJSON.from_dict(_item)

        trend_score = d.pop("trendScore", UNSET)

        trending_entry_json = cls(
            basis=basis,
            rank=rank,
            item=item,
            trend_score=trend_score,
        )

        trending_entry_json.additional_properties = d
        return trending_entry_json

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
