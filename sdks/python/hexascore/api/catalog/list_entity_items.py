from http import HTTPStatus
from typing import Any
from urllib.parse import quote

import httpx

from ...client import AuthenticatedClient, Client
from ...models.item_page_json import ItemPageJSON
from ...models.problem import Problem
from ...types import UNSET, Response, Unset


def _get_kwargs(
    id: str,
    *,
    type_: str | Unset = UNSET,
    role: str | Unset = UNSET,
    sort: str | Unset = UNSET,
    limit: str | Unset = UNSET,
    cursor: str | Unset = UNSET,
    locale: str | Unset = UNSET,
) -> dict[str, Any]:

    params: dict[str, Any] = {}

    params["type"] = type_

    params["role"] = role

    params["sort"] = sort

    params["limit"] = limit

    params["cursor"] = cursor

    params["locale"] = locale

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/entities/{id}/items".format(
            id=quote(str(id), safe=""),
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
    id: str,
    *,
    client: AuthenticatedClient,
    type_: str | Unset = UNSET,
    role: str | Unset = UNSET,
    sort: str | Unset = UNSET,
    limit: str | Unset = UNSET,
    cursor: str | Unset = UNSET,
    locale: str | Unset = UNSET,
) -> Response[ItemPageJSON | Problem]:
    """listEntityItems

    Args:
        id (str):
        type_ (str | Unset):
        role (str | Unset):
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
        id=id,
        type_=type_,
        role=role,
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
    id: str,
    *,
    client: AuthenticatedClient,
    type_: str | Unset = UNSET,
    role: str | Unset = UNSET,
    sort: str | Unset = UNSET,
    limit: str | Unset = UNSET,
    cursor: str | Unset = UNSET,
    locale: str | Unset = UNSET,
) -> ItemPageJSON | Problem | None:
    """listEntityItems

    Args:
        id (str):
        type_ (str | Unset):
        role (str | Unset):
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
        id=id,
        client=client,
        type_=type_,
        role=role,
        sort=sort,
        limit=limit,
        cursor=cursor,
        locale=locale,
    ).parsed


async def asyncio_detailed(
    id: str,
    *,
    client: AuthenticatedClient,
    type_: str | Unset = UNSET,
    role: str | Unset = UNSET,
    sort: str | Unset = UNSET,
    limit: str | Unset = UNSET,
    cursor: str | Unset = UNSET,
    locale: str | Unset = UNSET,
) -> Response[ItemPageJSON | Problem]:
    """listEntityItems

    Args:
        id (str):
        type_ (str | Unset):
        role (str | Unset):
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
        id=id,
        type_=type_,
        role=role,
        sort=sort,
        limit=limit,
        cursor=cursor,
        locale=locale,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    id: str,
    *,
    client: AuthenticatedClient,
    type_: str | Unset = UNSET,
    role: str | Unset = UNSET,
    sort: str | Unset = UNSET,
    limit: str | Unset = UNSET,
    cursor: str | Unset = UNSET,
    locale: str | Unset = UNSET,
) -> ItemPageJSON | Problem | None:
    """listEntityItems

    Args:
        id (str):
        type_ (str | Unset):
        role (str | Unset):
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
            id=id,
            client=client,
            type_=type_,
            role=role,
            sort=sort,
            limit=limit,
            cursor=cursor,
            locale=locale,
        )
    ).parsed
