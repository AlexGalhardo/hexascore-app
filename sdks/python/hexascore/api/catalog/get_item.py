from http import HTTPStatus
from typing import Any
from urllib.parse import quote

import httpx

from ...client import AuthenticatedClient, Client
from ...models.item_json import ItemJSON
from ...models.problem import Problem
from ...types import UNSET, Response, Unset


def _get_kwargs(
    type_: str,
    id: str,
    *,
    locale: str | Unset = UNSET,
) -> dict[str, Any]:

    params: dict[str, Any] = {}

    params["locale"] = locale

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/api/v1/{type_}/{id}".format(
            type_=quote(str(type_), safe=""),
            id=quote(str(id), safe=""),
        ),
        "params": params,
    }

    return _kwargs


def _parse_response(*, client: AuthenticatedClient | Client, response: httpx.Response) -> ItemJSON | Problem:
    if response.status_code == 200:
        response_200 = ItemJSON.from_dict(response.json())

        return response_200

    if response.status_code == 401:
        response_401 = Problem.from_dict(response.json())

        return response_401

    if response.status_code == 422:
        response_422 = Problem.from_dict(response.json())

        return response_422

    response_default = Problem.from_dict(response.json())

    return response_default


def _build_response(*, client: AuthenticatedClient | Client, response: httpx.Response) -> Response[ItemJSON | Problem]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    type_: str,
    id: str,
    *,
    client: AuthenticatedClient,
    locale: str | Unset = UNSET,
) -> Response[ItemJSON | Problem]:
    """getItem

    Args:
        type_ (str):
        id (str):
        locale (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[ItemJSON | Problem]
    """

    kwargs = _get_kwargs(
        type_=type_,
        id=id,
        locale=locale,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    type_: str,
    id: str,
    *,
    client: AuthenticatedClient,
    locale: str | Unset = UNSET,
) -> ItemJSON | Problem | None:
    """getItem

    Args:
        type_ (str):
        id (str):
        locale (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        ItemJSON | Problem
    """

    return sync_detailed(
        type_=type_,
        id=id,
        client=client,
        locale=locale,
    ).parsed


async def asyncio_detailed(
    type_: str,
    id: str,
    *,
    client: AuthenticatedClient,
    locale: str | Unset = UNSET,
) -> Response[ItemJSON | Problem]:
    """getItem

    Args:
        type_ (str):
        id (str):
        locale (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[ItemJSON | Problem]
    """

    kwargs = _get_kwargs(
        type_=type_,
        id=id,
        locale=locale,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    type_: str,
    id: str,
    *,
    client: AuthenticatedClient,
    locale: str | Unset = UNSET,
) -> ItemJSON | Problem | None:
    """getItem

    Args:
        type_ (str):
        id (str):
        locale (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        ItemJSON | Problem
    """

    return (
        await asyncio_detailed(
            type_=type_,
            id=id,
            client=client,
            locale=locale,
        )
    ).parsed
