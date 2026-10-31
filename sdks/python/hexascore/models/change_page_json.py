from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.change_json import ChangeJSON
    from ..models.links import Links
    from ..models.meta import Meta


T = TypeVar("T", bound="ChangePageJSON")


@_attrs_define
class ChangePageJSON:
    """
    Attributes:
        data (list[ChangeJSON] | Unset):
        links (Links | Unset):
        meta (Meta | Unset):
    """

    data: list[ChangeJSON] | Unset = UNSET
    links: Links | Unset = UNSET
    meta: Meta | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        data: list[dict[str, Any]] | Unset = UNSET
        if not isinstance(self.data, Unset):
            data = []
            for data_item_data in self.data:
                data_item = data_item_data.to_dict()
                data.append(data_item)

        links: dict[str, Any] | Unset = UNSET
        if not isinstance(self.links, Unset):
            links = self.links.to_dict()

        meta: dict[str, Any] | Unset = UNSET
        if not isinstance(self.meta, Unset):
            meta = self.meta.to_dict()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update({})
        if data is not UNSET:
            field_dict["data"] = data
        if links is not UNSET:
            field_dict["links"] = links
        if meta is not UNSET:
            field_dict["meta"] = meta

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.change_json import ChangeJSON  # noqa: PLC0415
        from ..models.links import Links  # noqa: PLC0415
        from ..models.meta import Meta  # noqa: PLC0415

        d = dict(src_dict)
        _data = d.pop("data", UNSET)
        data: list[ChangeJSON] | Unset = UNSET
        if _data is not UNSET:
            data = []
            for data_item_data in _data:
                data_item = ChangeJSON.from_dict(data_item_data)

                data.append(data_item)

        _links = d.pop("links", UNSET)
        links: Links | Unset
        if isinstance(_links, Unset):
            links = UNSET
        else:
            links = Links.from_dict(_links)

        _meta = d.pop("meta", UNSET)
        meta: Meta | Unset
        if isinstance(_meta, Unset):
            meta = UNSET
        else:
            meta = Meta.from_dict(_meta)

        change_page_json = cls(
            data=data,
            links=links,
            meta=meta,
        )

        change_page_json.additional_properties = d
        return change_page_json

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
