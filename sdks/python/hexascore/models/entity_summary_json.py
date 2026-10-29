from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

T = TypeVar("T", bound="EntitySummaryJSON")


@_attrs_define
class EntitySummaryJSON:
    """
    Attributes:
        id (str):
        kind (str):
        slug (str):
        item_count (int | Unset):
        name (str | Unset):
    """

    id: str
    kind: str
    slug: str
    item_count: int | Unset = UNSET
    name: str | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        id = self.id

        kind = self.kind

        slug = self.slug

        item_count = self.item_count

        name = self.name

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "kind": kind,
                "slug": slug,
            }
        )
        if item_count is not UNSET:
            field_dict["itemCount"] = item_count
        if name is not UNSET:
            field_dict["name"] = name

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        id = d.pop("id")

        kind = d.pop("kind")

        slug = d.pop("slug")

        item_count = d.pop("itemCount", UNSET)

        name = d.pop("name", UNSET)

        entity_summary_json = cls(
            id=id,
            kind=kind,
            slug=slug,
            item_count=item_count,
            name=name,
        )

        entity_summary_json.additional_properties = d
        return entity_summary_json

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
