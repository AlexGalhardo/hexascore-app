from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar

from attrs import define as _attrs_define

from ..models.usage_plan import UsagePlan

if TYPE_CHECKING:
    from ..models.usage_credits import UsageCredits
    from ..models.usage_quota import UsageQuota
    from ..models.usage_rate_limit import UsageRateLimit


T = TypeVar("T", bound="Usage")


@_attrs_define
class Usage:
    """Quota, credits and rate limit of the key's account. This call counts as one request.

    Attributes:
        plan (UsagePlan):
        quota (UsageQuota):
        credits_ (UsageCredits):
        rate_limit (UsageRateLimit):
    """

    plan: UsagePlan
    quota: UsageQuota
    credits_: UsageCredits
    rate_limit: UsageRateLimit

    def to_dict(self) -> dict[str, Any]:
        plan = self.plan.value

        quota = self.quota.to_dict()

        credits_ = self.credits_.to_dict()

        rate_limit = self.rate_limit.to_dict()

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "plan": plan,
                "quota": quota,
                "credits": credits_,
                "rate_limit": rate_limit,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.usage_credits import UsageCredits  # noqa: PLC0415
        from ..models.usage_quota import UsageQuota  # noqa: PLC0415
        from ..models.usage_rate_limit import UsageRateLimit  # noqa: PLC0415

        d = dict(src_dict)
        plan = UsagePlan(d.pop("plan"))

        quota = UsageQuota.from_dict(d.pop("quota"))

        credits_ = UsageCredits.from_dict(d.pop("credits"))

        rate_limit = UsageRateLimit.from_dict(d.pop("rate_limit"))

        usage = cls(
            plan=plan,
            quota=quota,
            credits_=credits_,
            rate_limit=rate_limit,
        )

        return usage
