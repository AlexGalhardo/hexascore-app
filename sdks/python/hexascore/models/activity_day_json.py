from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

T = TypeVar("T", bound="ActivityDayJSON")


@_attrs_define
class ActivityDayJSON:
    """
    Attributes:
        day (str):
        list_adds (int | Unset):
        ratings_new (int | Unset):
        ratings_updated (int | Unset):
        search_clicks (int | Unset):
        status_completed (int | Unset):
        status_dropped (int | Unset):
        status_in_progress (int | Unset):
        status_want (int | Unset):
        suggestions (int | Unset):
        views (int | Unset):
    """

    day: str
    list_adds: int | Unset = UNSET
    ratings_new: int | Unset = UNSET
    ratings_updated: int | Unset = UNSET
    search_clicks: int | Unset = UNSET
    status_completed: int | Unset = UNSET
    status_dropped: int | Unset = UNSET
    status_in_progress: int | Unset = UNSET
    status_want: int | Unset = UNSET
    suggestions: int | Unset = UNSET
    views: int | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        day = self.day

        list_adds = self.list_adds

        ratings_new = self.ratings_new

        ratings_updated = self.ratings_updated

        search_clicks = self.search_clicks

        status_completed = self.status_completed

        status_dropped = self.status_dropped

        status_in_progress = self.status_in_progress

        status_want = self.status_want

        suggestions = self.suggestions

        views = self.views

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "day": day,
            }
        )
        if list_adds is not UNSET:
            field_dict["listAdds"] = list_adds
        if ratings_new is not UNSET:
            field_dict["ratingsNew"] = ratings_new
        if ratings_updated is not UNSET:
            field_dict["ratingsUpdated"] = ratings_updated
        if search_clicks is not UNSET:
            field_dict["searchClicks"] = search_clicks
        if status_completed is not UNSET:
            field_dict["statusCompleted"] = status_completed
        if status_dropped is not UNSET:
            field_dict["statusDropped"] = status_dropped
        if status_in_progress is not UNSET:
            field_dict["statusInProgress"] = status_in_progress
        if status_want is not UNSET:
            field_dict["statusWant"] = status_want
        if suggestions is not UNSET:
            field_dict["suggestions"] = suggestions
        if views is not UNSET:
            field_dict["views"] = views

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        day = d.pop("day")

        list_adds = d.pop("listAdds", UNSET)

        ratings_new = d.pop("ratingsNew", UNSET)

        ratings_updated = d.pop("ratingsUpdated", UNSET)

        search_clicks = d.pop("searchClicks", UNSET)

        status_completed = d.pop("statusCompleted", UNSET)

        status_dropped = d.pop("statusDropped", UNSET)

        status_in_progress = d.pop("statusInProgress", UNSET)

        status_want = d.pop("statusWant", UNSET)

        suggestions = d.pop("suggestions", UNSET)

        views = d.pop("views", UNSET)

        activity_day_json = cls(
            day=day,
            list_adds=list_adds,
            ratings_new=ratings_new,
            ratings_updated=ratings_updated,
            search_clicks=search_clicks,
            status_completed=status_completed,
            status_dropped=status_dropped,
            status_in_progress=status_in_progress,
            status_want=status_want,
            suggestions=suggestions,
            views=views,
        )

        activity_day_json.additional_properties = d
        return activity_day_json

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
