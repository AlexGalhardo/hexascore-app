from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, cast

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

T = TypeVar("T", bound="Meta")


@_attrs_define
class Meta:
    """
    Attributes:
        last_page (int | Unset):
        next_cursor (str | Unset):
        page (int | Unset):
        per_page (int | Unset):
        total (int | None | Unset):
    """

    last_page: int | Unset = UNSET
    next_cursor: str | Unset = UNSET
    page: int | Unset = UNSET
    per_page: int | Unset = UNSET
    total: int | None | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        last_page = self.last_page

        next_cursor = self.next_cursor

        page = self.page

        per_page = self.per_page

        total: int | None | Unset
        if isinstance(self.total, Unset):
            total = UNSET
        else:
            total = self.total

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update({})
        if last_page is not UNSET:
            field_dict["last_page"] = last_page
        if next_cursor is not UNSET:
            field_dict["next_cursor"] = next_cursor
        if page is not UNSET:
            field_dict["page"] = page
        if per_page is not UNSET:
            field_dict["per_page"] = per_page
        if total is not UNSET:
            field_dict["total"] = total

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        last_page = d.pop("last_page", UNSET)

        next_cursor = d.pop("next_cursor", UNSET)

        page = d.pop("page", UNSET)

        per_page = d.pop("per_page", UNSET)

        def _parse_total(data: object) -> int | None | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(int | None | Unset, data)

        total = _parse_total(d.pop("total", UNSET))

        meta = cls(
            last_page=last_page,
            next_cursor=next_cursor,
            page=page,
            per_page=per_page,
            total=total,
        )

        meta.additional_properties = d
        return meta

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
