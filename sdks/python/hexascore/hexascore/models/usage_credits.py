from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar

from attrs import define as _attrs_define

T = TypeVar("T", bound="UsageCredits")


@_attrs_define
class UsageCredits:
    """
    Attributes:
        balance (int): Prepaid requests used after the quota runs out (paid plans).
    """

    balance: int

    def to_dict(self) -> dict[str, Any]:
        balance = self.balance

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "balance": balance,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        balance = d.pop("balance")

        usage_credits = cls(
            balance=balance,
        )

        return usage_credits
