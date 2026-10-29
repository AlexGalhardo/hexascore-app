from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar, cast

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.entity_json_external_ids import EntityJSONExternalIds
    from ..models.entity_json_provenance import EntityJSONProvenance
    from ..models.item_meta import ItemMeta
    from ..models.synopsis import Synopsis


T = TypeVar("T", bound="EntityJSON")


@_attrs_define
class EntityJSON:
    """
    Attributes:
        id (str):
        kind (str):
        locale (str):
        slug (str):
        born_on (str | Unset):
        code (str | Unset):
        description (None | Synopsis | Unset):
        died_on (str | Unset):
        external_ids (EntityJSONExternalIds | Unset):
        founded_year (int | None | Unset):
        meta (ItemMeta | Unset):
        name (str | Unset):
        provenance (EntityJSONProvenance | Unset):
        wikidata_qid (str | Unset):
    """

    id: str
    kind: str
    locale: str
    slug: str
    born_on: str | Unset = UNSET
    code: str | Unset = UNSET
    description: None | Synopsis | Unset = UNSET
    died_on: str | Unset = UNSET
    external_ids: EntityJSONExternalIds | Unset = UNSET
    founded_year: int | None | Unset = UNSET
    meta: ItemMeta | Unset = UNSET
    name: str | Unset = UNSET
    provenance: EntityJSONProvenance | Unset = UNSET
    wikidata_qid: str | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        from ..models.synopsis import Synopsis  # noqa: PLC0415

        id = self.id

        kind = self.kind

        locale = self.locale

        slug = self.slug

        born_on = self.born_on

        code = self.code

        description: dict[str, Any] | None | Unset
        if isinstance(self.description, Unset):
            description = UNSET
        elif isinstance(self.description, Synopsis):
            description = self.description.to_dict()
        else:
            description = self.description

        died_on = self.died_on

        external_ids: dict[str, Any] | Unset = UNSET
        if not isinstance(self.external_ids, Unset):
            external_ids = self.external_ids.to_dict()

        founded_year: int | None | Unset
        if isinstance(self.founded_year, Unset):
            founded_year = UNSET
        else:
            founded_year = self.founded_year

        meta: dict[str, Any] | Unset = UNSET
        if not isinstance(self.meta, Unset):
            meta = self.meta.to_dict()

        name = self.name

        provenance: dict[str, Any] | Unset = UNSET
        if not isinstance(self.provenance, Unset):
            provenance = self.provenance.to_dict()

        wikidata_qid = self.wikidata_qid

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "kind": kind,
                "locale": locale,
                "slug": slug,
            }
        )
        if born_on is not UNSET:
            field_dict["bornOn"] = born_on
        if code is not UNSET:
            field_dict["code"] = code
        if description is not UNSET:
            field_dict["description"] = description
        if died_on is not UNSET:
            field_dict["diedOn"] = died_on
        if external_ids is not UNSET:
            field_dict["externalIds"] = external_ids
        if founded_year is not UNSET:
            field_dict["foundedYear"] = founded_year
        if meta is not UNSET:
            field_dict["meta"] = meta
        if name is not UNSET:
            field_dict["name"] = name
        if provenance is not UNSET:
            field_dict["provenance"] = provenance
        if wikidata_qid is not UNSET:
            field_dict["wikidataQid"] = wikidata_qid

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.entity_json_external_ids import EntityJSONExternalIds  # noqa: PLC0415
        from ..models.entity_json_provenance import EntityJSONProvenance  # noqa: PLC0415
        from ..models.item_meta import ItemMeta  # noqa: PLC0415
        from ..models.synopsis import Synopsis  # noqa: PLC0415

        d = dict(src_dict)
        id = d.pop("id")

        kind = d.pop("kind")

        locale = d.pop("locale")

        slug = d.pop("slug")

        born_on = d.pop("bornOn", UNSET)

        code = d.pop("code", UNSET)

        def _parse_description(data: object) -> None | Synopsis | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            try:
                if not isinstance(data, dict):
                    raise TypeError()
                description_type_0 = Synopsis.from_dict(data)

                return description_type_0
            except (TypeError, ValueError, AttributeError, KeyError):
                pass
            return cast(None | Synopsis | Unset, data)

        description = _parse_description(d.pop("description", UNSET))

        died_on = d.pop("diedOn", UNSET)

        _external_ids = d.pop("externalIds", UNSET)
        external_ids: EntityJSONExternalIds | Unset
        if isinstance(_external_ids, Unset):
            external_ids = UNSET
        else:
            external_ids = EntityJSONExternalIds.from_dict(_external_ids)

        def _parse_founded_year(data: object) -> int | None | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(int | None | Unset, data)

        founded_year = _parse_founded_year(d.pop("foundedYear", UNSET))

        _meta = d.pop("meta", UNSET)
        meta: ItemMeta | Unset
        if isinstance(_meta, Unset):
            meta = UNSET
        else:
            meta = ItemMeta.from_dict(_meta)

        name = d.pop("name", UNSET)

        _provenance = d.pop("provenance", UNSET)
        provenance: EntityJSONProvenance | Unset
        if isinstance(_provenance, Unset):
            provenance = UNSET
        else:
            provenance = EntityJSONProvenance.from_dict(_provenance)

        wikidata_qid = d.pop("wikidataQid", UNSET)

        entity_json = cls(
            id=id,
            kind=kind,
            locale=locale,
            slug=slug,
            born_on=born_on,
            code=code,
            description=description,
            died_on=died_on,
            external_ids=external_ids,
            founded_year=founded_year,
            meta=meta,
            name=name,
            provenance=provenance,
            wikidata_qid=wikidata_qid,
        )

        entity_json.additional_properties = d
        return entity_json

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
