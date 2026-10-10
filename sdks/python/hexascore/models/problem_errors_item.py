from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar

from attrs import define as _attrs_define

T = TypeVar("T", bound="ProblemErrorsItem")


@_attrs_define
class ProblemErrorsItem:
    """
    Attributes:
        path (str):
        message (str):
    """

    path: str
    message: str

    def to_dict(self) -> dict[str, Any]:
        path = self.path

        message = self.message

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "path": path,
                "message": message,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        path = d.pop("path")

        message = d.pop("message")

        problem_errors_item = cls(
            path=path,
            message=message,
        )

        return problem_errors_item
