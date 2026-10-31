from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, cast

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

T = TypeVar("T", bound="WebhookCreated")


@_attrs_define
class WebhookCreated:
    """
    Attributes:
        created_at (str):
        events (list[str]):
        id (str):
        item_types (list[str]):
        secret (str):
        url (str):
        disabled (bool | Unset):
    """

    created_at: str
    events: list[str]
    id: str
    item_types: list[str]
    secret: str
    url: str
    disabled: bool | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        created_at = self.created_at

        events = self.events

        id = self.id

        item_types = self.item_types

        secret = self.secret

        url = self.url

        disabled = self.disabled

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "createdAt": created_at,
                "events": events,
                "id": id,
                "itemTypes": item_types,
                "secret": secret,
                "url": url,
            }
        )
        if disabled is not UNSET:
            field_dict["disabled"] = disabled

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        created_at = d.pop("createdAt")

        events = cast(list[str], d.pop("events"))

        id = d.pop("id")

        item_types = cast(list[str], d.pop("itemTypes"))

        secret = d.pop("secret")

        url = d.pop("url")

        disabled = d.pop("disabled", UNSET)

        webhook_created = cls(
            created_at=created_at,
            events=events,
            id=id,
            item_types=item_types,
            secret=secret,
            url=url,
            disabled=disabled,
        )

        webhook_created.additional_properties = d
        return webhook_created

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
