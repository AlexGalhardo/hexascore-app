from http import HTTPStatus
from typing import Any

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.item_list import ItemList
from ...models.list_series_locale import ListSeriesLocale
from ...models.list_series_sort import ListSeriesSort
from ...models.problem import Problem
from ...types import UNSET, Response, Unset


def _get_kwargs(
    *,
    genre: str | Unset = UNSET,
    year: int | Unset = UNSET,
    min_score: int | Unset = UNSET,
    max_score: int | Unset = UNSET,
    min_screenplay: int | Unset = UNSET,
    min_acting: int | Unset = UNSET,
    min_characters: int | Unset = UNSET,
    min_pacing: int | Unset = UNSET,
    min_production: int | Unset = UNSET,
    min_entertainment: int | Unset = UNSET,
    sort: ListSeriesSort | Unset = ListSeriesSort.POPULARITY_DESC,
    limit: int | Unset = 10,
    cursor: str | Unset = UNSET,
    locale: ListSeriesLocale | Unset = ListSeriesLocale.EN,
) -> dict[str, Any]:

    params: dict[str, Any] = {}

    params["genre"] = genre

    params["year"] = year

    params["min_score"] = min_score

    params["max_score"] = max_score

    params["min_screenplay"] = min_screenplay

    params["min_acting"] = min_acting

    params["min_characters"] = min_characters

    params["min_pacing"] = min_pacing

    params["min_production"] = min_production

    params["min_entertainment"] = min_entertainment

    json_sort: str | Unset = UNSET
    if not isinstance(sort, Unset):
        json_sort = sort.value

    params["sort"] = json_sort

    params["limit"] = limit

    params["cursor"] = cursor

    json_locale: str | Unset = UNSET
    if not isinstance(locale, Unset):
        json_locale = locale.value

    params["locale"] = json_locale

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/series",
        "params": params,
    }

    return _kwargs


def _parse_response(*, client: AuthenticatedClient | Client, response: httpx.Response) -> ItemList | Problem | None:
    if response.status_code == 200:
        response_200 = ItemList.from_dict(response.json())

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


def _build_response(*, client: AuthenticatedClient | Client, response: httpx.Response) -> Response[ItemList | Problem]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    *,
    client: AuthenticatedClient | Client,
    genre: str | Unset = UNSET,
    year: int | Unset = UNSET,
    min_score: int | Unset = UNSET,
    max_score: int | Unset = UNSET,
    min_screenplay: int | Unset = UNSET,
    min_acting: int | Unset = UNSET,
    min_characters: int | Unset = UNSET,
    min_pacing: int | Unset = UNSET,
    min_production: int | Unset = UNSET,
    min_entertainment: int | Unset = UNSET,
    sort: ListSeriesSort | Unset = ListSeriesSort.POPULARITY_DESC,
    limit: int | Unset = 10,
    cursor: str | Unset = UNSET,
    locale: ListSeriesLocale | Unset = ListSeriesLocale.EN,
) -> Response[ItemList | Problem]:
    """List TV series

     Pages through TV series with filters. Categories for this type: `screenplay`, `acting`,
    `characters`, `pacing`, `production`, `entertainment` — filter with `min_<category>`. Pass
    `next_cursor` back as `cursor` until it is `null`.

    Args:
        genre (str | Unset):
        year (int | Unset):
        min_score (int | Unset):
        max_score (int | Unset):
        min_screenplay (int | Unset):
        min_acting (int | Unset):
        min_characters (int | Unset):
        min_pacing (int | Unset):
        min_production (int | Unset):
        min_entertainment (int | Unset):
        sort (ListSeriesSort | Unset):  Default: ListSeriesSort.POPULARITY_DESC.
        limit (int | Unset):  Default: 10.
        cursor (str | Unset):
        locale (ListSeriesLocale | Unset):  Default: ListSeriesLocale.EN.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[ItemList | Problem]
    """

    kwargs = _get_kwargs(
        genre=genre,
        year=year,
        min_score=min_score,
        max_score=max_score,
        min_screenplay=min_screenplay,
        min_acting=min_acting,
        min_characters=min_characters,
        min_pacing=min_pacing,
        min_production=min_production,
        min_entertainment=min_entertainment,
        sort=sort,
        limit=limit,
        cursor=cursor,
        locale=locale,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    *,
    client: AuthenticatedClient | Client,
    genre: str | Unset = UNSET,
    year: int | Unset = UNSET,
    min_score: int | Unset = UNSET,
    max_score: int | Unset = UNSET,
    min_screenplay: int | Unset = UNSET,
    min_acting: int | Unset = UNSET,
    min_characters: int | Unset = UNSET,
    min_pacing: int | Unset = UNSET,
    min_production: int | Unset = UNSET,
    min_entertainment: int | Unset = UNSET,
    sort: ListSeriesSort | Unset = ListSeriesSort.POPULARITY_DESC,
    limit: int | Unset = 10,
    cursor: str | Unset = UNSET,
    locale: ListSeriesLocale | Unset = ListSeriesLocale.EN,
) -> ItemList | Problem | None:
    """List TV series

     Pages through TV series with filters. Categories for this type: `screenplay`, `acting`,
    `characters`, `pacing`, `production`, `entertainment` — filter with `min_<category>`. Pass
    `next_cursor` back as `cursor` until it is `null`.

    Args:
        genre (str | Unset):
        year (int | Unset):
        min_score (int | Unset):
        max_score (int | Unset):
        min_screenplay (int | Unset):
        min_acting (int | Unset):
        min_characters (int | Unset):
        min_pacing (int | Unset):
        min_production (int | Unset):
        min_entertainment (int | Unset):
        sort (ListSeriesSort | Unset):  Default: ListSeriesSort.POPULARITY_DESC.
        limit (int | Unset):  Default: 10.
        cursor (str | Unset):
        locale (ListSeriesLocale | Unset):  Default: ListSeriesLocale.EN.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        ItemList | Problem
    """

    return sync_detailed(
        client=client,
        genre=genre,
        year=year,
        min_score=min_score,
        max_score=max_score,
        min_screenplay=min_screenplay,
        min_acting=min_acting,
        min_characters=min_characters,
        min_pacing=min_pacing,
        min_production=min_production,
        min_entertainment=min_entertainment,
        sort=sort,
        limit=limit,
        cursor=cursor,
        locale=locale,
    ).parsed


async def asyncio_detailed(
    *,
    client: AuthenticatedClient | Client,
    genre: str | Unset = UNSET,
    year: int | Unset = UNSET,
    min_score: int | Unset = UNSET,
    max_score: int | Unset = UNSET,
    min_screenplay: int | Unset = UNSET,
    min_acting: int | Unset = UNSET,
    min_characters: int | Unset = UNSET,
    min_pacing: int | Unset = UNSET,
    min_production: int | Unset = UNSET,
    min_entertainment: int | Unset = UNSET,
    sort: ListSeriesSort | Unset = ListSeriesSort.POPULARITY_DESC,
    limit: int | Unset = 10,
    cursor: str | Unset = UNSET,
    locale: ListSeriesLocale | Unset = ListSeriesLocale.EN,
) -> Response[ItemList | Problem]:
    """List TV series

     Pages through TV series with filters. Categories for this type: `screenplay`, `acting`,
    `characters`, `pacing`, `production`, `entertainment` — filter with `min_<category>`. Pass
    `next_cursor` back as `cursor` until it is `null`.

    Args:
        genre (str | Unset):
        year (int | Unset):
        min_score (int | Unset):
        max_score (int | Unset):
        min_screenplay (int | Unset):
        min_acting (int | Unset):
        min_characters (int | Unset):
        min_pacing (int | Unset):
        min_production (int | Unset):
        min_entertainment (int | Unset):
        sort (ListSeriesSort | Unset):  Default: ListSeriesSort.POPULARITY_DESC.
        limit (int | Unset):  Default: 10.
        cursor (str | Unset):
        locale (ListSeriesLocale | Unset):  Default: ListSeriesLocale.EN.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[ItemList | Problem]
    """

    kwargs = _get_kwargs(
        genre=genre,
        year=year,
        min_score=min_score,
        max_score=max_score,
        min_screenplay=min_screenplay,
        min_acting=min_acting,
        min_characters=min_characters,
        min_pacing=min_pacing,
        min_production=min_production,
        min_entertainment=min_entertainment,
        sort=sort,
        limit=limit,
        cursor=cursor,
        locale=locale,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    *,
    client: AuthenticatedClient | Client,
    genre: str | Unset = UNSET,
    year: int | Unset = UNSET,
    min_score: int | Unset = UNSET,
    max_score: int | Unset = UNSET,
    min_screenplay: int | Unset = UNSET,
    min_acting: int | Unset = UNSET,
    min_characters: int | Unset = UNSET,
    min_pacing: int | Unset = UNSET,
    min_production: int | Unset = UNSET,
    min_entertainment: int | Unset = UNSET,
    sort: ListSeriesSort | Unset = ListSeriesSort.POPULARITY_DESC,
    limit: int | Unset = 10,
    cursor: str | Unset = UNSET,
    locale: ListSeriesLocale | Unset = ListSeriesLocale.EN,
) -> ItemList | Problem | None:
    """List TV series

     Pages through TV series with filters. Categories for this type: `screenplay`, `acting`,
    `characters`, `pacing`, `production`, `entertainment` — filter with `min_<category>`. Pass
    `next_cursor` back as `cursor` until it is `null`.

    Args:
        genre (str | Unset):
        year (int | Unset):
        min_score (int | Unset):
        max_score (int | Unset):
        min_screenplay (int | Unset):
        min_acting (int | Unset):
        min_characters (int | Unset):
        min_pacing (int | Unset):
        min_production (int | Unset):
        min_entertainment (int | Unset):
        sort (ListSeriesSort | Unset):  Default: ListSeriesSort.POPULARITY_DESC.
        limit (int | Unset):  Default: 10.
        cursor (str | Unset):
        locale (ListSeriesLocale | Unset):  Default: ListSeriesLocale.EN.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        ItemList | Problem
    """

    return (
        await asyncio_detailed(
            client=client,
            genre=genre,
            year=year,
            min_score=min_score,
            max_score=max_score,
            min_screenplay=min_screenplay,
            min_acting=min_acting,
            min_characters=min_characters,
            min_pacing=min_pacing,
            min_production=min_production,
            min_entertainment=min_entertainment,
            sort=sort,
            limit=limit,
            cursor=cursor,
            locale=locale,
        )
    ).parsed
