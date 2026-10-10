from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar

from attrs import define as _attrs_define

if TYPE_CHECKING:
    from ..models.time_to_beat_completionist import TimeToBeatCompletionist
    from ..models.time_to_beat_extras import TimeToBeatExtras
    from ..models.time_to_beat_main import TimeToBeatMain


T = TypeVar("T", bound="TimeToBeat")


@_attrs_define
class TimeToBeat:
    """Crowdsourced time to beat per play style: main story, main + extras, completionist.

    Attributes:
        main (TimeToBeatMain):
        extras (TimeToBeatExtras):
        completionist (TimeToBeatCompletionist):
    """

    main: TimeToBeatMain
    extras: TimeToBeatExtras
    completionist: TimeToBeatCompletionist

    def to_dict(self) -> dict[str, Any]:
        main = self.main.to_dict()

        extras = self.extras.to_dict()

        completionist = self.completionist.to_dict()

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "main": main,
                "extras": extras,
                "completionist": completionist,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.time_to_beat_completionist import TimeToBeatCompletionist  # noqa: PLC0415
        from ..models.time_to_beat_extras import TimeToBeatExtras  # noqa: PLC0415
        from ..models.time_to_beat_main import TimeToBeatMain  # noqa: PLC0415

        d = dict(src_dict)
        main = TimeToBeatMain.from_dict(d.pop("main"))

        extras = TimeToBeatExtras.from_dict(d.pop("extras"))

        completionist = TimeToBeatCompletionist.from_dict(d.pop("completionist"))

        time_to_beat = cls(
            main=main,
            extras=extras,
            completionist=completionist,
        )

        return time_to_beat
