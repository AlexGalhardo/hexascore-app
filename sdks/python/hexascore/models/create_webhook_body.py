from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, cast

from attrs import define as _attrs_define
from attrs import field as _attrs_field

T = TypeVar("T", bound="CreateWebhookBody")


@_attrs_define
class CreateWebhookBody:
    """
    Attributes:
        events (list[str]):
        item_types (list[str]):
        url (str):
    """

    events: list[str]
    item_types: list[str]
    url: str
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        events = self.events

        item_types = self.item_types

        url = self.url

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "events": events,
                "itemTypes": item_types,
                "url": url,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        events = cast(list[str], d.pop("events"))

        item_types = cast(list[str], d.pop("itemTypes"))

        url = d.pop("url")

        create_webhook_body = cls(
            events=events,
            item_types=item_types,
            url=url,
        )

        create_webhook_body.additional_properties = d
        return create_webhook_body

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
