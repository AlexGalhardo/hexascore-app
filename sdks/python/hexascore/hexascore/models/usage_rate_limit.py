from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar

from attrs import define as _attrs_define

T = TypeVar("T", bound="UsageRateLimit")


@_attrs_define
class UsageRateLimit:
    """
    Attributes:
        requests_per_second (int):
    """

    requests_per_second: int

    def to_dict(self) -> dict[str, Any]:
        requests_per_second = self.requests_per_second

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "requests_per_second": requests_per_second,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        requests_per_second = d.pop("requests_per_second")

        usage_rate_limit = cls(
            requests_per_second=requests_per_second,
        )

        return usage_rate_limit
