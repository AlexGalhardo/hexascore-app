from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, cast

from attrs import define as _attrs_define
from attrs import field as _attrs_field

T = TypeVar("T", bound="ItemDetailCategories")


@_attrs_define
class ItemDetailCategories:
    """Average per category, 0–100, `null` under 5 ratings. Keys by type — `game`: `gameplay`, `fun`, `sound`, `technical`,
    `visual`, `storytelling`; `movie`: `screenplay`, `acting`, `direction`, `cinematography`, `soundtrack`,
    `entertainment`; `series`: `screenplay`, `acting`, `characters`, `pacing`, `production`, `entertainment`; `book`:
    `plot`, `characters`, `writing`, `pacing`, `originality`, `impact`.

    """

    additional_properties: dict[str, int | None] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:

        field_dict: dict[str, Any] = {}
        for prop_name, prop in self.additional_properties.items():
            field_dict[prop_name] = prop

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        item_detail_categories = cls()

        additional_properties = {}
        for prop_name, prop_dict in d.items():

            def _parse_additional_property(data: object) -> int | None:
                if data is None:
                    return data
                return cast(int | None, data)

            additional_property = _parse_additional_property(prop_dict)

            additional_properties[prop_name] = additional_property

        item_detail_categories.additional_properties = additional_properties
        return item_detail_categories

    @property
    def additional_keys(self) -> list[str]:
        return list(self.additional_properties.keys())

    def __getitem__(self, key: str) -> int | None:
        return self.additional_properties[key]

    def __setitem__(self, key: str, value: int | None) -> None:
        self.additional_properties[key] = value

    def __delitem__(self, key: str) -> None:
        del self.additional_properties[key]

    def __contains__(self, key: str) -> bool:
        return key in self.additional_properties
