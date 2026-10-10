from http import HTTPStatus
from typing import Any

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.item_detail import ItemDetail
from ...models.lookup_item_locale import LookupItemLocale
from ...models.lookup_item_source import LookupItemSource
from ...models.problem import Problem
from ...types import UNSET, Response, Unset


def _get_kwargs(
    *,
    source: LookupItemSource,
    id: str,
    locale: LookupItemLocale | Unset = LookupItemLocale.EN,
) -> dict[str, Any]:

    params: dict[str, Any] = {}

    json_source = source.value
    params["source"] = json_source

    params["id"] = id

    json_locale: str | Unset = UNSET
    if not isinstance(locale, Unset):
        json_locale = locale.value

    params["locale"] = json_locale

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/lookup",
        "params": params,
    }

    return _kwargs


def _parse_response(*, client: AuthenticatedClient | Client, response: httpx.Response) -> ItemDetail | Problem | None:
    if response.status_code == 200:
        response_200 = ItemDetail.from_dict(response.json())

        return response_200

    if response.status_code == 400:
        response_400 = Problem.from_dict(response.json())

        return response_400

    if response.status_code == 401:
        response_401 = Problem.from_dict(response.json())

        return response_401

    if response.status_code == 402:
        response_402 = Problem.from_dict(response.json())

        return response_402

    if response.status_code == 403:
        response_403 = Problem.from_dict(response.json())

        return response_403

    if response.status_code == 404:
        response_404 = Problem.from_dict(response.json())

        return response_404

    if response.status_code == 429:
        response_429 = Problem.from_dict(response.json())

        return response_429

    if response.status_code == 500:
        response_500 = Problem.from_dict(response.json())

        return response_500

    if response.status_code == 503:
        response_503 = Problem.from_dict(response.json())

        return response_503

    if client.raise_on_unexpected_status:
        raise errors.UnexpectedStatus(response.status_code, response.content)
    else:
        return None


def _build_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> Response[ItemDetail | Problem]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    *,
    client: AuthenticatedClient | Client,
    source: LookupItemSource,
    id: str,
    locale: LookupItemLocale | Unset = LookupItemLocale.EN,
) -> Response[ItemDetail | Problem]:
    """Look up by external id

     Finds the item for an IMDb, Steam, IGDB, TMDB, ISBN-13, Open Library or Wikidata id. Ids must match
    exactly (e.g. `tt0111161`, `Q42`).

    Args:
        source (LookupItemSource):
        id (str):
        locale (LookupItemLocale | Unset):  Default: LookupItemLocale.EN.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[ItemDetail | Problem]
    """

    kwargs = _get_kwargs(
        source=source,
        id=id,
        locale=locale,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    *,
    client: AuthenticatedClient | Client,
    source: LookupItemSource,
    id: str,
    locale: LookupItemLocale | Unset = LookupItemLocale.EN,
) -> ItemDetail | Problem | None:
    """Look up by external id

     Finds the item for an IMDb, Steam, IGDB, TMDB, ISBN-13, Open Library or Wikidata id. Ids must match
    exactly (e.g. `tt0111161`, `Q42`).

    Args:
        source (LookupItemSource):
        id (str):
        locale (LookupItemLocale | Unset):  Default: LookupItemLocale.EN.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        ItemDetail | Problem
    """

    return sync_detailed(
        client=client,
        source=source,
        id=id,
        locale=locale,
    ).parsed


async def asyncio_detailed(
    *,
    client: AuthenticatedClient | Client,
    source: LookupItemSource,
    id: str,
    locale: LookupItemLocale | Unset = LookupItemLocale.EN,
) -> Response[ItemDetail | Problem]:
    """Look up by external id

     Finds the item for an IMDb, Steam, IGDB, TMDB, ISBN-13, Open Library or Wikidata id. Ids must match
    exactly (e.g. `tt0111161`, `Q42`).

    Args:
        source (LookupItemSource):
        id (str):
        locale (LookupItemLocale | Unset):  Default: LookupItemLocale.EN.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[ItemDetail | Problem]
    """

    kwargs = _get_kwargs(
        source=source,
        id=id,
        locale=locale,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    *,
    client: AuthenticatedClient | Client,
    source: LookupItemSource,
    id: str,
    locale: LookupItemLocale | Unset = LookupItemLocale.EN,
) -> ItemDetail | Problem | None:
    """Look up by external id

     Finds the item for an IMDb, Steam, IGDB, TMDB, ISBN-13, Open Library or Wikidata id. Ids must match
    exactly (e.g. `tt0111161`, `Q42`).

    Args:
        source (LookupItemSource):
        id (str):
        locale (LookupItemLocale | Unset):  Default: LookupItemLocale.EN.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        ItemDetail | Problem
    """

    return (
        await asyncio_detailed(
            client=client,
            source=source,
            id=id,
            locale=locale,
        )
    ).parsed
