from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.usage_quota import UsageQuota
    from ..models.usage_rate_limit import UsageRateLimit


T = TypeVar("T", bound="UsageJSON")


@_attrs_define
class UsageJSON:
    """
    Attributes:
        plan (str):
        quota (UsageQuota | Unset):
        rate_limit (UsageRateLimit | Unset):
    """

    plan: str
    quota: UsageQuota | Unset = UNSET
    rate_limit: UsageRateLimit | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        plan = self.plan

        quota: dict[str, Any] | Unset = UNSET
        if not isinstance(self.quota, Unset):
            quota = self.quota.to_dict()

        rate_limit: dict[str, Any] | Unset = UNSET
        if not isinstance(self.rate_limit, Unset):
            rate_limit = self.rate_limit.to_dict()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "plan": plan,
            }
        )
        if quota is not UNSET:
            field_dict["quota"] = quota
        if rate_limit is not UNSET:
            field_dict["rate_limit"] = rate_limit

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.usage_quota import UsageQuota  # noqa: PLC0415
        from ..models.usage_rate_limit import UsageRateLimit  # noqa: PLC0415

        d = dict(src_dict)
        plan = d.pop("plan")

        _quota = d.pop("quota", UNSET)
        quota: UsageQuota | Unset
        if isinstance(_quota, Unset):
            quota = UNSET
        else:
            quota = UsageQuota.from_dict(_quota)

        _rate_limit = d.pop("rate_limit", UNSET)
        rate_limit: UsageRateLimit | Unset
        if isinstance(_rate_limit, Unset):
            rate_limit = UNSET
        else:
            rate_limit = UsageRateLimit.from_dict(_rate_limit)

        usage_json = cls(
            plan=plan,
            quota=quota,
            rate_limit=rate_limit,
        )

        usage_json.additional_properties = d
        return usage_json

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
