from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, cast

from attrs import define as _attrs_define

T = TypeVar("T", bound="ItemDetailTrailersItem")


@_attrs_define
class ItemDetailTrailersItem:
    """
    Attributes:
        youtube_id (str):
        title (None | str):
    """

    youtube_id: str
    title: None | str

    def to_dict(self) -> dict[str, Any]:
        youtube_id = self.youtube_id

        title: None | str
        title = self.title

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "youtube_id": youtube_id,
                "title": title,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        youtube_id = d.pop("youtube_id")

        def _parse_title(data: object) -> None | str:
            if data is None:
                return data
            return cast(None | str, data)

        title = _parse_title(d.pop("title"))

        item_detail_trailers_item = cls(
            youtube_id=youtube_id,
            title=title,
        )

        return item_detail_trailers_item
