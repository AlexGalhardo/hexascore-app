from __future__ import annotations

from collections.abc import Mapping
from typing import Any, TypeVar

from attrs import define as _attrs_define

from ..types import UNSET, Unset

T = TypeVar("T", bound="ItemDetailExternalIds")


@_attrs_define
class ItemDetailExternalIds:
    """Identifiers of the item in other databases. Only known ids are present.

    Attributes:
        wikidata (str):
        imdb (str | Unset):
        igdb (str | Unset):
        steam (str | Unset):
        tmdb (str | Unset):
        isbn13 (str | Unset):
        openlibrary (str | Unset):
        hltb (str | Unset):
        goodreads (str | Unset):
        metacritic (str | Unset):
        rottentomatoes (str | Unset):
        epic (str | Unset):
        gog (str | Unset):
        skoob (str | Unset):
    """

    wikidata: str
    imdb: str | Unset = UNSET
    igdb: str | Unset = UNSET
    steam: str | Unset = UNSET
    tmdb: str | Unset = UNSET
    isbn13: str | Unset = UNSET
    openlibrary: str | Unset = UNSET
    hltb: str | Unset = UNSET
    goodreads: str | Unset = UNSET
    metacritic: str | Unset = UNSET
    rottentomatoes: str | Unset = UNSET
    epic: str | Unset = UNSET
    gog: str | Unset = UNSET
    skoob: str | Unset = UNSET

    def to_dict(self) -> dict[str, Any]:
        wikidata = self.wikidata

        imdb = self.imdb

        igdb = self.igdb

        steam = self.steam

        tmdb = self.tmdb

        isbn13 = self.isbn13

        openlibrary = self.openlibrary

        hltb = self.hltb

        goodreads = self.goodreads

        metacritic = self.metacritic

        rottentomatoes = self.rottentomatoes

        epic = self.epic

        gog = self.gog

        skoob = self.skoob

        field_dict: dict[str, Any] = {}

        field_dict.update(
            {
                "wikidata": wikidata,
            }
        )
        if imdb is not UNSET:
            field_dict["imdb"] = imdb
        if igdb is not UNSET:
            field_dict["igdb"] = igdb
        if steam is not UNSET:
            field_dict["steam"] = steam
        if tmdb is not UNSET:
            field_dict["tmdb"] = tmdb
        if isbn13 is not UNSET:
            field_dict["isbn13"] = isbn13
        if openlibrary is not UNSET:
            field_dict["openlibrary"] = openlibrary
        if hltb is not UNSET:
            field_dict["hltb"] = hltb
        if goodreads is not UNSET:
            field_dict["goodreads"] = goodreads
        if metacritic is not UNSET:
            field_dict["metacritic"] = metacritic
        if rottentomatoes is not UNSET:
            field_dict["rottentomatoes"] = rottentomatoes
        if epic is not UNSET:
            field_dict["epic"] = epic
        if gog is not UNSET:
            field_dict["gog"] = gog
        if skoob is not UNSET:
            field_dict["skoob"] = skoob

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        d = dict(src_dict)
        wikidata = d.pop("wikidata")

        imdb = d.pop("imdb", UNSET)

        igdb = d.pop("igdb", UNSET)

        steam = d.pop("steam", UNSET)

        tmdb = d.pop("tmdb", UNSET)

        isbn13 = d.pop("isbn13", UNSET)

        openlibrary = d.pop("openlibrary", UNSET)

        hltb = d.pop("hltb", UNSET)

        goodreads = d.pop("goodreads", UNSET)

        metacritic = d.pop("metacritic", UNSET)

        rottentomatoes = d.pop("rottentomatoes", UNSET)

        epic = d.pop("epic", UNSET)

        gog = d.pop("gog", UNSET)

        skoob = d.pop("skoob", UNSET)

        item_detail_external_ids = cls(
            wikidata=wikidata,
            imdb=imdb,
            igdb=igdb,
            steam=steam,
            tmdb=tmdb,
            isbn13=isbn13,
            openlibrary=openlibrary,
            hltb=hltb,
            goodreads=goodreads,
            metacritic=metacritic,
            rottentomatoes=rottentomatoes,
            epic=epic,
            gog=gog,
            skoob=skoob,
        )

        return item_detail_external_ids
