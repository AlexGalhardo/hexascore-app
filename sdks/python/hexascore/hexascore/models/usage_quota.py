from __future__ import annotations

import datetime
from collections.abc import Mapping
from typing import Any, TypeVar, cast

from attrs import define as _attrs_define

from ..models.usage_quota_period import UsageQuotaPeriod

T = TypeVar("T", bound="UsageQuota")


@_attrs_define
class UsageQuota:
    """
    Attributes:
        limit (int):
        used (int):
        remaining (int):
        period (UsageQuotaPeriod): `lifetime` on Free (100 requests in total); `month` on paid plans (UTC calendar
            month).
        resets_at (datetime.datetime | None): Next reset; `null` for lifetime quotas.
    """

    limit: int
    used: int
    remaining: int
    period: UsageQuotaPeriod
    resets_at: datetime.datetime | None

    def to_dict(self) -> dict[str, Any]:
        limit = self.limit

        used = self.used

        remaining = self.remaining

        period = self.period.value

        resets_at: None | str
        if isinstance(self.resets_at, datetime.datetime):
            resets_at = self.resets_at.isoformat()
        else:
            resets_at = self.resets_at

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "limit": limit,
                "used": used,
                "remaining": remaining,
                "period": period,
                "resets_at": resets_at,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        limit = d.pop("limit")

        used = d.pop("used")

        remaining = d.pop("remaining")

        period = UsageQuotaPeriod(d.pop("period"))

        def _parse_resets_at(data: object) -> datetime.datetime | None:
            if data is None:
                return data
            try:
                if not isinstance(data, str):
                    raise TypeError()
                resets_at_type_0 = datetime.datetime.fromisoformat(data)

                return resets_at_type_0
            except (TypeError, ValueError, AttributeError, KeyError):
                pass
            return cast(datetime.datetime | None, data)

        resets_at = _parse_resets_at(d.pop("resets_at"))

        usage_quota = cls(
            limit=limit,
            used=used,
            remaining=remaining,
            period=period,
            resets_at=resets_at,
        )

        return usage_quota
