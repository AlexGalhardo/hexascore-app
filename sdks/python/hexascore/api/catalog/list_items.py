from http import HTTPStatus
from typing import Any
from urllib.parse import quote

import httpx

from ...client import AuthenticatedClient, Client
from ...models.item_page_json import ItemPageJSON
from ...models.list_items_type import ListItemsType
from ...models.problem import Problem
from ...types import UNSET, Response, Unset


def _get_kwargs(
    type_: ListItemsType,
    *,
    genre: str | Unset = UNSET,
    platform: str | Unset = UNSET,
    year: str | Unset = UNSET,
    min_score: str | Unset = UNSET,
    max_score: str | Unset = UNSET,
    sort: str | Unset = UNSET,
    limit: str | Unset = UNSET,
    cursor: str | Unset = UNSET,
    locale: str | Unset = UNSET,
) -> dict[str, Any]:

    params: dict[str, Any] = {}

    params["genre"] = genre

    params["platform"] = platform

    params["year"] = year

    params["min_score"] = min_score

    params["max_score"] = max_score

    params["sort"] = sort

    params["limit"] = limit

    params["cursor"] = cursor

    params["locale"] = locale

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/{type_}".format(
            type_=quote(str(type_), safe=""),
        ),
        "params": params,
    }

    return _kwargs


def _parse_response(*, client: AuthenticatedClient | Client, response: httpx.Response) -> ItemPageJSON | Problem:
    if response.status_code == 200:
        response_200 = ItemPageJSON.from_dict(response.json())

        return response_200

    if response.status_code == 401:
        response_401 = Problem.from_dict(response.json())

        return response_401

    if response.status_code == 422:
        response_422 = Problem.from_dict(response.json())

        return response_422

    response_default = Problem.from_dict(response.json())

    return response_default


def _build_response(
    *, client: AuthenticatedClient | Client, response: httpx.Response
) -> Response[ItemPageJSON | Problem]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    type_: ListItemsType,
    *,
    client: AuthenticatedClient,
    genre: str | Unset = UNSET,
    platform: str | Unset = UNSET,
    year: str | Unset = UNSET,
    min_score: str | Unset = UNSET,
    max_score: str | Unset = UNSET,
    sort: str | Unset = UNSET,
    limit: str | Unset = UNSET,
    cursor: str | Unset = UNSET,
    locale: str | Unset = UNSET,
) -> Response[ItemPageJSON | Problem]:
    """listItems

    Args:
        type_ (ListItemsType):
        genre (str | Unset):
        platform (str | Unset):
        year (str | Unset):
        min_score (str | Unset):
        max_score (str | Unset):
        sort (str | Unset):
        limit (str | Unset):
        cursor (str | Unset):
        locale (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[ItemPageJSON | Problem]
    """

    kwargs = _get_kwargs(
        type_=type_,
        genre=genre,
        platform=platform,
        year=year,
        min_score=min_score,
        max_score=max_score,
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
    type_: ListItemsType,
    *,
    client: AuthenticatedClient,
    genre: str | Unset = UNSET,
    platform: str | Unset = UNSET,
    year: str | Unset = UNSET,
    min_score: str | Unset = UNSET,
    max_score: str | Unset = UNSET,
    sort: str | Unset = UNSET,
    limit: str | Unset = UNSET,
    cursor: str | Unset = UNSET,
    locale: str | Unset = UNSET,
) -> ItemPageJSON | Problem | None:
    """listItems

    Args:
        type_ (ListItemsType):
        genre (str | Unset):
        platform (str | Unset):
        year (str | Unset):
        min_score (str | Unset):
        max_score (str | Unset):
        sort (str | Unset):
        limit (str | Unset):
        cursor (str | Unset):
        locale (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        ItemPageJSON | Problem
    """

    return sync_detailed(
        type_=type_,
        client=client,
        genre=genre,
        platform=platform,
        year=year,
        min_score=min_score,
        max_score=max_score,
        sort=sort,
        limit=limit,
        cursor=cursor,
        locale=locale,
    ).parsed


async def asyncio_detailed(
    type_: ListItemsType,
    *,
    client: AuthenticatedClient,
    genre: str | Unset = UNSET,
    platform: str | Unset = UNSET,
    year: str | Unset = UNSET,
    min_score: str | Unset = UNSET,
    max_score: str | Unset = UNSET,
    sort: str | Unset = UNSET,
    limit: str | Unset = UNSET,
    cursor: str | Unset = UNSET,
    locale: str | Unset = UNSET,
) -> Response[ItemPageJSON | Problem]:
    """listItems

    Args:
        type_ (ListItemsType):
        genre (str | Unset):
        platform (str | Unset):
        year (str | Unset):
        min_score (str | Unset):
        max_score (str | Unset):
        sort (str | Unset):
        limit (str | Unset):
        cursor (str | Unset):
        locale (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[ItemPageJSON | Problem]
    """

    kwargs = _get_kwargs(
        type_=type_,
        genre=genre,
        platform=platform,
        year=year,
        min_score=min_score,
        max_score=max_score,
        sort=sort,
        limit=limit,
        cursor=cursor,
        locale=locale,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    type_: ListItemsType,
    *,
    client: AuthenticatedClient,
    genre: str | Unset = UNSET,
    platform: str | Unset = UNSET,
    year: str | Unset = UNSET,
    min_score: str | Unset = UNSET,
    max_score: str | Unset = UNSET,
    sort: str | Unset = UNSET,
    limit: str | Unset = UNSET,
    cursor: str | Unset = UNSET,
    locale: str | Unset = UNSET,
) -> ItemPageJSON | Problem | None:
    """listItems

    Args:
        type_ (ListItemsType):
        genre (str | Unset):
        platform (str | Unset):
        year (str | Unset):
        min_score (str | Unset):
        max_score (str | Unset):
        sort (str | Unset):
        limit (str | Unset):
        cursor (str | Unset):
        locale (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        ItemPageJSON | Problem
    """

    return (
        await asyncio_detailed(
            type_=type_,
            client=client,
            genre=genre,
            platform=platform,
            year=year,
            min_score=min_score,
            max_score=max_score,
            sort=sort,
            limit=limit,
            cursor=cursor,
            locale=locale,
        )
    ).parsed
