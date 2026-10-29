from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

T = TypeVar("T", bound="Synopsis")


@_attrs_define
class Synopsis:
    """
    Attributes:
        licence (str | Unset):
        source_url (str | Unset):
        text (str | Unset):
    """

    licence: str | Unset = UNSET
    source_url: str | Unset = UNSET
    text: str | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        licence = self.licence

        source_url = self.source_url

        text = self.text

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update({})
        if licence is not UNSET:
            field_dict["licence"] = licence
        if source_url is not UNSET:
            field_dict["sourceUrl"] = source_url
        if text is not UNSET:
            field_dict["text"] = text

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        licence = d.pop("licence", UNSET)

        source_url = d.pop("sourceUrl", UNSET)

        text = d.pop("text", UNSET)

        synopsis = cls(
            licence=licence,
            source_url=source_url,
            text=text,
        )

        synopsis.additional_properties = d
        return synopsis

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
