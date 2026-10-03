from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, cast

from attrs import define as _attrs_define

T = TypeVar("T", bound="TimeToBeatCompletionist")


@_attrs_define
class TimeToBeatCompletionist:
    """
    Attributes:
        median_hours (float | None): Median hours reported by players who completed the game (1 decimal). `null` under 3
            reports.
        reports (int): Number of players who reported this style.
    """

    median_hours: float | None
    reports: int

    def to_dict(self) -> dict[str, Any]:
        median_hours: float | None
        median_hours = self.median_hours

        reports = self.reports

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "median_hours": median_hours,
                "reports": reports,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)

        def _parse_median_hours(data: object) -> float | None:
            if data is None:
                return data
            return cast(float | None, data)

        median_hours = _parse_median_hours(d.pop("median_hours"))

        reports = d.pop("reports")

        time_to_beat_completionist = cls(
            median_hours=median_hours,
            reports=reports,
        )

        return time_to_beat_completionist
