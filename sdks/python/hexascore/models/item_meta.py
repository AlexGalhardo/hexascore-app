from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, cast

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

T = TypeVar("T", bound="ItemMeta")


@_attrs_define
class ItemMeta:
    """
    Attributes:
        omitted_fields (list[str] | Unset):
    """

    omitted_fields: list[str] | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        omitted_fields: list[str] | Unset = UNSET
        if not isinstance(self.omitted_fields, Unset):
            omitted_fields = self.omitted_fields

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update({})
        if omitted_fields is not UNSET:
            field_dict["omittedFields"] = omitted_fields

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        omitted_fields = cast(list[str], d.pop("omittedFields", UNSET))

        item_meta = cls(
            omitted_fields=omitted_fields,
        )

        item_meta.additional_properties = d
        return item_meta

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
