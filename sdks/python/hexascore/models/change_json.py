from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, cast

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..models.change_json_action import ChangeJSONAction
from ..models.change_json_item_type import ChangeJSONItemType
from ..types import UNSET, Unset

T = TypeVar("T", bound="ChangeJSON")


@_attrs_define
class ChangeJSON:
    """
    Attributes:
        action (ChangeJSONAction):
        changed_at (str):
        id (str):
        item_id (str):
        item_type (ChangeJSONItemType):
        fields (list[str] | Unset):
        merged_into (None | str | Unset):
    """

    action: ChangeJSONAction
    changed_at: str
    id: str
    item_id: str
    item_type: ChangeJSONItemType
    fields: list[str] | Unset = UNSET
    merged_into: None | str | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        action = self.action.value

        changed_at = self.changed_at

        id = self.id

        item_id = self.item_id

        item_type = self.item_type.value

        fields: list[str] | Unset = UNSET
        if not isinstance(self.fields, Unset):
            fields = self.fields

        merged_into: None | str | Unset
        if isinstance(self.merged_into, Unset):
            merged_into = UNSET
        else:
            merged_into = self.merged_into

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "action": action,
                "changedAt": changed_at,
                "id": id,
                "itemId": item_id,
                "itemType": item_type,
            }
        )
        if fields is not UNSET:
            field_dict["fields"] = fields
        if merged_into is not UNSET:
            field_dict["mergedInto"] = merged_into

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        action = ChangeJSONAction(d.pop("action"))

        changed_at = d.pop("changedAt")

        id = d.pop("id")

        item_id = d.pop("itemId")

        item_type = ChangeJSONItemType(d.pop("itemType"))

        fields = cast(list[str], d.pop("fields", UNSET))

        def _parse_merged_into(data: object) -> None | str | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(None | str | Unset, data)

        merged_into = _parse_merged_into(d.pop("mergedInto", UNSET))

        change_json = cls(
            action=action,
            changed_at=changed_at,
            id=id,
            item_id=item_id,
            item_type=item_type,
            fields=fields,
            merged_into=merged_into,
        )

        change_json.additional_properties = d
        return change_json

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
