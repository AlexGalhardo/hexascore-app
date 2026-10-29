from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, cast

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..models.usage_quota_period import UsageQuotaPeriod
from ..types import UNSET, Unset

T = TypeVar("T", bound="UsageQuota")


@_attrs_define
class UsageQuota:
    """
    Attributes:
        limit (int | Unset):
        period (UsageQuotaPeriod | Unset):
        remaining (int | Unset):
        resets_at (None | str | Unset):
        used (int | Unset):
    """

    limit: int | Unset = UNSET
    period: UsageQuotaPeriod | Unset = UNSET
    remaining: int | Unset = UNSET
    resets_at: None | str | Unset = UNSET
    used: int | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        limit = self.limit

        period: str | Unset = UNSET
        if not isinstance(self.period, Unset):
            period = self.period.value

        remaining = self.remaining

        resets_at: None | str | Unset
        if isinstance(self.resets_at, Unset):
            resets_at = UNSET
        else:
            resets_at = self.resets_at

        used = self.used

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update({})
        if limit is not UNSET:
            field_dict["limit"] = limit
        if period is not UNSET:
            field_dict["period"] = period
        if remaining is not UNSET:
            field_dict["remaining"] = remaining
        if resets_at is not UNSET:
            field_dict["resets_at"] = resets_at
        if used is not UNSET:
            field_dict["used"] = used

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        limit = d.pop("limit", UNSET)

        _period = d.pop("period", UNSET)
        period: UsageQuotaPeriod | Unset
        if isinstance(_period, Unset):
            period = UNSET
        else:
            period = UsageQuotaPeriod(_period)

        remaining = d.pop("remaining", UNSET)

        def _parse_resets_at(data: object) -> None | str | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(None | str | Unset, data)

        resets_at = _parse_resets_at(d.pop("resets_at", UNSET))

        used = d.pop("used", UNSET)

        usage_quota = cls(
            limit=limit,
            period=period,
            remaining=remaining,
            resets_at=resets_at,
            used=used,
        )

        usage_quota.additional_properties = d
        return usage_quota

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
