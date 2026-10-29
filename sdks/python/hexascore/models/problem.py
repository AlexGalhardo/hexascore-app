from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.problem_errors import ProblemErrors


T = TypeVar("T", bound="Problem")


@_attrs_define
class Problem:
    """
    Attributes:
        detail (str | Unset):
        errors (ProblemErrors | Unset):
        instance (str | Unset):
        request_id (str | Unset):
        status (int | Unset):
        title (str | Unset):
        type_ (str | Unset):
    """

    detail: str | Unset = UNSET
    errors: ProblemErrors | Unset = UNSET
    instance: str | Unset = UNSET
    request_id: str | Unset = UNSET
    status: int | Unset = UNSET
    title: str | Unset = UNSET
    type_: str | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        detail = self.detail

        errors: dict[str, Any] | Unset = UNSET
        if not isinstance(self.errors, Unset):
            errors = self.errors.to_dict()

        instance = self.instance

        request_id = self.request_id

        status = self.status

        title = self.title

        type_ = self.type_

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update({})
        if detail is not UNSET:
            field_dict["detail"] = detail
        if errors is not UNSET:
            field_dict["errors"] = errors
        if instance is not UNSET:
            field_dict["instance"] = instance
        if request_id is not UNSET:
            field_dict["request_id"] = request_id
        if status is not UNSET:
            field_dict["status"] = status
        if title is not UNSET:
            field_dict["title"] = title
        if type_ is not UNSET:
            field_dict["type"] = type_

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.problem_errors import ProblemErrors  # noqa: PLC0415

        d = dict(src_dict)
        detail = d.pop("detail", UNSET)

        _errors = d.pop("errors", UNSET)
        errors: ProblemErrors | Unset
        if isinstance(_errors, Unset):
            errors = UNSET
        else:
            errors = ProblemErrors.from_dict(_errors)

        instance = d.pop("instance", UNSET)

        request_id = d.pop("request_id", UNSET)

        status = d.pop("status", UNSET)

        title = d.pop("title", UNSET)

        type_ = d.pop("type", UNSET)

        problem = cls(
            detail=detail,
            errors=errors,
            instance=instance,
            request_id=request_id,
            status=status,
            title=title,
            type_=type_,
        )

        problem.additional_properties = d
        return problem

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
