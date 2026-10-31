from http import HTTPStatus
from typing import Any

import httpx

from ...client import AuthenticatedClient, Client
from ...models.item_json import ItemJSON
from ...models.problem import Problem
from ...types import UNSET, Response, Unset


def _get_kwargs(
    *,
    source: str | Unset = UNSET,
    id: str | Unset = UNSET,
    locale: str | Unset = UNSET,
) -> dict[str, Any]:

    params: dict[str, Any] = {}

    params["source"] = source

    params["id"] = id

    params["locale"] = locale

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/lookup",
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
    *,
    client: AuthenticatedClient,
    source: str | Unset = UNSET,
    id: str | Unset = UNSET,
    locale: str | Unset = UNSET,
) -> Response[ItemJSON | Problem]:
    """lookupItem

    Args:
        source (str | Unset):
        id (str | Unset):
        locale (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[ItemJSON | Problem]
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
    client: AuthenticatedClient,
    source: str | Unset = UNSET,
    id: str | Unset = UNSET,
    locale: str | Unset = UNSET,
) -> ItemJSON | Problem | None:
    """lookupItem

    Args:
        source (str | Unset):
        id (str | Unset):
        locale (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        ItemJSON | Problem
    """

    return sync_detailed(
        client=client,
        source=source,
        id=id,
        locale=locale,
    ).parsed


async def asyncio_detailed(
    *,
    client: AuthenticatedClient,
    source: str | Unset = UNSET,
    id: str | Unset = UNSET,
    locale: str | Unset = UNSET,
) -> Response[ItemJSON | Problem]:
    """lookupItem

    Args:
        source (str | Unset):
        id (str | Unset):
        locale (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[ItemJSON | Problem]
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
    client: AuthenticatedClient,
    source: str | Unset = UNSET,
    id: str | Unset = UNSET,
    locale: str | Unset = UNSET,
) -> ItemJSON | Problem | None:
    """lookupItem

    Args:
        source (str | Unset):
        id (str | Unset):
        locale (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        ItemJSON | Problem
    """

    return (
        await asyncio_detailed(
            client=client,
            source=source,
            id=id,
            locale=locale,
        )
    ).parsed
