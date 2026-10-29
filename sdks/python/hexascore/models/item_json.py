from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar, cast

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.entity_ref import EntityRef
    from ..models.item_json_external_ids import ItemJSONExternalIds
    from ..models.item_json_provenance import ItemJSONProvenance
    from ..models.item_json_score_by_category import ItemJSONScoreByCategory
    from ..models.item_meta import ItemMeta
    from ..models.synopsis import Synopsis


T = TypeVar("T", bound="ItemJSON")


@_attrs_define
class ItemJSON:
    """
    Attributes:
        id (str):
        locale (str):
        slug (str):
        type_ (str):
        confidence (None | str | Unset):
        entities (list[EntityRef] | Unset):
        external_ids (ItemJSONExternalIds | Unset):
        formula_version (int | None | Unset):
        meta (ItemMeta | Unset):
        page_count (int | None | Unset):
        provenance (ItemJSONProvenance | Unset):
        rating_count (int | Unset):
        release_date (str | Unset):
        release_precision (str | Unset):
        runtime_minutes (int | None | Unset):
        score (float | None | Unset):
        score_by_category (ItemJSONScoreByCategory | Unset):
        synopsis (None | Synopsis | Unset):
        title (str | Unset):
        wikidata_qid (str | Unset):
    """

    id: str
    locale: str
    slug: str
    type_: str
    confidence: None | str | Unset = UNSET
    entities: list[EntityRef] | Unset = UNSET
    external_ids: ItemJSONExternalIds | Unset = UNSET
    formula_version: int | None | Unset = UNSET
    meta: ItemMeta | Unset = UNSET
    page_count: int | None | Unset = UNSET
    provenance: ItemJSONProvenance | Unset = UNSET
    rating_count: int | Unset = UNSET
    release_date: str | Unset = UNSET
    release_precision: str | Unset = UNSET
    runtime_minutes: int | None | Unset = UNSET
    score: float | None | Unset = UNSET
    score_by_category: ItemJSONScoreByCategory | Unset = UNSET
    synopsis: None | Synopsis | Unset = UNSET
    title: str | Unset = UNSET
    wikidata_qid: str | Unset = UNSET
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        from ..models.synopsis import Synopsis  # noqa: PLC0415

        id = self.id

        locale = self.locale

        slug = self.slug

        type_ = self.type_

        confidence: None | str | Unset
        if isinstance(self.confidence, Unset):
            confidence = UNSET
        else:
            confidence = self.confidence

        entities: list[dict[str, Any]] | Unset = UNSET
        if not isinstance(self.entities, Unset):
            entities = []
            for entities_item_data in self.entities:
                entities_item = entities_item_data.to_dict()
                entities.append(entities_item)

        external_ids: dict[str, Any] | Unset = UNSET
        if not isinstance(self.external_ids, Unset):
            external_ids = self.external_ids.to_dict()

        formula_version: int | None | Unset
        if isinstance(self.formula_version, Unset):
            formula_version = UNSET
        else:
            formula_version = self.formula_version

        meta: dict[str, Any] | Unset = UNSET
        if not isinstance(self.meta, Unset):
            meta = self.meta.to_dict()

        page_count: int | None | Unset
        if isinstance(self.page_count, Unset):
            page_count = UNSET
        else:
            page_count = self.page_count

        provenance: dict[str, Any] | Unset = UNSET
        if not isinstance(self.provenance, Unset):
            provenance = self.provenance.to_dict()

        rating_count = self.rating_count

        release_date = self.release_date

        release_precision = self.release_precision

        runtime_minutes: int | None | Unset
        if isinstance(self.runtime_minutes, Unset):
            runtime_minutes = UNSET
        else:
            runtime_minutes = self.runtime_minutes

        score: float | None | Unset
        if isinstance(self.score, Unset):
            score = UNSET
        else:
            score = self.score

        score_by_category: dict[str, Any] | Unset = UNSET
        if not isinstance(self.score_by_category, Unset):
            score_by_category = self.score_by_category.to_dict()

        synopsis: dict[str, Any] | None | Unset
        if isinstance(self.synopsis, Unset):
            synopsis = UNSET
        elif isinstance(self.synopsis, Synopsis):
            synopsis = self.synopsis.to_dict()
        else:
            synopsis = self.synopsis

        title = self.title

        wikidata_qid = self.wikidata_qid

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "locale": locale,
                "slug": slug,
                "type": type_,
            }
        )
        if confidence is not UNSET:
            field_dict["confidence"] = confidence
        if entities is not UNSET:
            field_dict["entities"] = entities
        if external_ids is not UNSET:
            field_dict["externalIds"] = external_ids
        if formula_version is not UNSET:
            field_dict["formulaVersion"] = formula_version
        if meta is not UNSET:
            field_dict["meta"] = meta
        if page_count is not UNSET:
            field_dict["pageCount"] = page_count
        if provenance is not UNSET:
            field_dict["provenance"] = provenance
        if rating_count is not UNSET:
            field_dict["ratingCount"] = rating_count
        if release_date is not UNSET:
            field_dict["releaseDate"] = release_date
        if release_precision is not UNSET:
            field_dict["releasePrecision"] = release_precision
        if runtime_minutes is not UNSET:
            field_dict["runtimeMinutes"] = runtime_minutes
        if score is not UNSET:
            field_dict["score"] = score
        if score_by_category is not UNSET:
            field_dict["scoreByCategory"] = score_by_category
        if synopsis is not UNSET:
            field_dict["synopsis"] = synopsis
        if title is not UNSET:
            field_dict["title"] = title
        if wikidata_qid is not UNSET:
            field_dict["wikidataQid"] = wikidata_qid

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.entity_ref import EntityRef  # noqa: PLC0415
        from ..models.item_json_external_ids import ItemJSONExternalIds  # noqa: PLC0415
        from ..models.item_json_provenance import ItemJSONProvenance  # noqa: PLC0415
        from ..models.item_json_score_by_category import ItemJSONScoreByCategory  # noqa: PLC0415
        from ..models.item_meta import ItemMeta  # noqa: PLC0415
        from ..models.synopsis import Synopsis  # noqa: PLC0415

        d = dict(src_dict)
        id = d.pop("id")

        locale = d.pop("locale")

        slug = d.pop("slug")

        type_ = d.pop("type")

        def _parse_confidence(data: object) -> None | str | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(None | str | Unset, data)

        confidence = _parse_confidence(d.pop("confidence", UNSET))

        _entities = d.pop("entities", UNSET)
        entities: list[EntityRef] | Unset = UNSET
        if _entities is not UNSET:
            entities = []
            for entities_item_data in _entities:
                entities_item = EntityRef.from_dict(entities_item_data)

                entities.append(entities_item)

        _external_ids = d.pop("externalIds", UNSET)
        external_ids: ItemJSONExternalIds | Unset
        if isinstance(_external_ids, Unset):
            external_ids = UNSET
        else:
            external_ids = ItemJSONExternalIds.from_dict(_external_ids)

        def _parse_formula_version(data: object) -> int | None | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(int | None | Unset, data)

        formula_version = _parse_formula_version(d.pop("formulaVersion", UNSET))

        _meta = d.pop("meta", UNSET)
        meta: ItemMeta | Unset
        if isinstance(_meta, Unset):
            meta = UNSET
        else:
            meta = ItemMeta.from_dict(_meta)

        def _parse_page_count(data: object) -> int | None | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(int | None | Unset, data)

        page_count = _parse_page_count(d.pop("pageCount", UNSET))

        _provenance = d.pop("provenance", UNSET)
        provenance: ItemJSONProvenance | Unset
        if isinstance(_provenance, Unset):
            provenance = UNSET
        else:
            provenance = ItemJSONProvenance.from_dict(_provenance)

        rating_count = d.pop("ratingCount", UNSET)

        release_date = d.pop("releaseDate", UNSET)

        release_precision = d.pop("releasePrecision", UNSET)

        def _parse_runtime_minutes(data: object) -> int | None | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(int | None | Unset, data)

        runtime_minutes = _parse_runtime_minutes(d.pop("runtimeMinutes", UNSET))

        def _parse_score(data: object) -> float | None | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            return cast(float | None | Unset, data)

        score = _parse_score(d.pop("score", UNSET))

        _score_by_category = d.pop("scoreByCategory", UNSET)
        score_by_category: ItemJSONScoreByCategory | Unset
        if isinstance(_score_by_category, Unset):
            score_by_category = UNSET
        else:
            score_by_category = ItemJSONScoreByCategory.from_dict(_score_by_category)

        def _parse_synopsis(data: object) -> None | Synopsis | Unset:
            if data is None:
                return data
            if isinstance(data, Unset):
                return data
            try:
                if not isinstance(data, dict):
                    raise TypeError()
                synopsis_type_0 = Synopsis.from_dict(data)

                return synopsis_type_0
            except (TypeError, ValueError, AttributeError, KeyError):
                pass
            return cast(None | Synopsis | Unset, data)

        synopsis = _parse_synopsis(d.pop("synopsis", UNSET))

        title = d.pop("title", UNSET)

        wikidata_qid = d.pop("wikidataQid", UNSET)

        item_json = cls(
            id=id,
            locale=locale,
            slug=slug,
            type_=type_,
            confidence=confidence,
            entities=entities,
            external_ids=external_ids,
            formula_version=formula_version,
            meta=meta,
            page_count=page_count,
            provenance=provenance,
            rating_count=rating_count,
            release_date=release_date,
            release_precision=release_precision,
            runtime_minutes=runtime_minutes,
            score=score,
            score_by_category=score_by_category,
            synopsis=synopsis,
            title=title,
            wikidata_qid=wikidata_qid,
        )

        item_json.additional_properties = d
        return item_json

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
