from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar

from attrs import define as _attrs_define

from ..models.problem_code import ProblemCode
from ..models.problem_plan import ProblemPlan
from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.problem_errors_item import ProblemErrorsItem


T = TypeVar("T", bound="Problem")


@_attrs_define
class Problem:
    """RFC 9457 problem details (`application/problem+json`).

    Attributes:
        type_ (str): Link to the documentation of this error.
        title (str):
        status (int):
        detail (str): What went wrong and how to fix it.
        code (ProblemCode): Stable machine-readable error code.
        request_id (str): Quote it when contacting support.
        errors (list[ProblemErrorsItem] | Unset): Invalid parameters (only on `bad_request`).
        plan (ProblemPlan | Unset):
        limit (int | Unset):
    """

    type_: str
    title: str
    status: int
    detail: str
    code: ProblemCode
    request_id: str
    errors: list[ProblemErrorsItem] | Unset = UNSET
    plan: ProblemPlan | Unset = UNSET
    limit: int | Unset = UNSET

    def to_dict(self) -> dict[str, Any]:
        type_ = self.type_

        title = self.title

        status = self.status

        detail = self.detail

        code = self.code.value

        request_id = self.request_id

        errors: list[dict[str, Any]] | Unset = UNSET
        if not isinstance(self.errors, Unset):
            errors = []
            for errors_item_data in self.errors:
                errors_item = errors_item_data.to_dict()
                errors.append(errors_item)

        plan: str | Unset = UNSET
        if not isinstance(self.plan, Unset):
            plan = self.plan.value

        limit = self.limit

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "type": type_,
                "title": title,
                "status": status,
                "detail": detail,
                "code": code,
                "request_id": request_id,
            }
        )
        if errors is not UNSET:
            field_dict["errors"] = errors
        if plan is not UNSET:
            field_dict["plan"] = plan
        if limit is not UNSET:
            field_dict["limit"] = limit

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.problem_errors_item import ProblemErrorsItem  # noqa: PLC0415

        d = dict(src_dict)
        type_ = d.pop("type")

        title = d.pop("title")

        status = d.pop("status")

        detail = d.pop("detail")

        code = ProblemCode(d.pop("code"))

        request_id = d.pop("request_id")

        _errors = d.pop("errors", UNSET)
        errors: list[ProblemErrorsItem] | Unset = UNSET
        if _errors is not UNSET:
            errors = []
            for errors_item_data in _errors:
                errors_item = ProblemErrorsItem.from_dict(errors_item_data)

                errors.append(errors_item)

        _plan = d.pop("plan", UNSET)
        plan: ProblemPlan | Unset
        if isinstance(_plan, Unset):
            plan = UNSET
        else:
            plan = ProblemPlan(_plan)

        limit = d.pop("limit", UNSET)

        problem = cls(
            type_=type_,
            title=title,
            status=status,
            detail=detail,
            code=code,
            request_id=request_id,
            errors=errors,
            plan=plan,
            limit=limit,
        )

        return problem
