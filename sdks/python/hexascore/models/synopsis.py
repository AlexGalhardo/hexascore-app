from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar, cast

from attrs import define as _attrs_define

from ..models.synopsis_locale import SynopsisLocale

T = TypeVar("T", bound="Synopsis")


@_attrs_define
class Synopsis:
    """Synopsis with the license you must honour when showing it.

    Attributes:
        text (str):
        locale (SynopsisLocale): Locale the text is written in.
        license_ (None | str): License of the text (e.g. `CC BY-SA 4.0`). Attribute it when you display the synopsis.
        source_url (None | str): Where the text comes from (for attribution).
    """

    text: str
    locale: SynopsisLocale
    license_: None | str
    source_url: None | str

    def to_dict(self) -> dict[str, Any]:
        text = self.text

        locale = self.locale.value

        license_: None | str
        license_ = self.license_

        source_url: None | str
        source_url = self.source_url

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "text": text,
                "locale": locale,
                "license": license_,
                "source_url": source_url,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        text = d.pop("text")

        locale = SynopsisLocale(d.pop("locale"))

        def _parse_license_(data: object) -> None | str:
            if data is None:
                return data
            return cast(None | str, data)

        license_ = _parse_license_(d.pop("license"))

        def _parse_source_url(data: object) -> None | str:
            if data is None:
                return data
            return cast(None | str, data)

        source_url = _parse_source_url(d.pop("source_url"))

        synopsis = cls(
            text=text,
            locale=locale,
            license_=license_,
            source_url=source_url,
        )

        return synopsis
