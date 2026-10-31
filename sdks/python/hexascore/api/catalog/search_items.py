from http import HTTPStatus
from typing import Any

import httpx

from ...client import AuthenticatedClient, Client
from ...models.item_page_json import ItemPageJSON
from ...models.problem import Problem
from ...types import UNSET, Response, Unset


def _get_kwargs(
    *,
    q: str | Unset = UNSET,
    type_: str | Unset = UNSET,
    locale: str | Unset = UNSET,
    limit: str | Unset = UNSET,
    page: str | Unset = UNSET,
) -> dict[str, Any]:

    params: dict[str, Any] = {}

    params["q"] = q

    params["type"] = type_

    params["locale"] = locale

    params["limit"] = limit

    params["page"] = page

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/search",
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
    *,
    client: AuthenticatedClient,
    q: str | Unset = UNSET,
    type_: str | Unset = UNSET,
    locale: str | Unset = UNSET,
    limit: str | Unset = UNSET,
    page: str | Unset = UNSET,
) -> Response[ItemPageJSON | Problem]:
    """searchItems

    Args:
        q (str | Unset):
        type_ (str | Unset):
        locale (str | Unset):
        limit (str | Unset):
        page (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[ItemPageJSON | Problem]
    """

    kwargs = _get_kwargs(
        q=q,
        type_=type_,
        locale=locale,
        limit=limit,
        page=page,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    *,
    client: AuthenticatedClient,
    q: str | Unset = UNSET,
    type_: str | Unset = UNSET,
    locale: str | Unset = UNSET,
    limit: str | Unset = UNSET,
    page: str | Unset = UNSET,
) -> ItemPageJSON | Problem | None:
    """searchItems

    Args:
        q (str | Unset):
        type_ (str | Unset):
        locale (str | Unset):
        limit (str | Unset):
        page (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        ItemPageJSON | Problem
    """

    return sync_detailed(
        client=client,
        q=q,
        type_=type_,
        locale=locale,
        limit=limit,
        page=page,
    ).parsed


async def asyncio_detailed(
    *,
    client: AuthenticatedClient,
    q: str | Unset = UNSET,
    type_: str | Unset = UNSET,
    locale: str | Unset = UNSET,
    limit: str | Unset = UNSET,
    page: str | Unset = UNSET,
) -> Response[ItemPageJSON | Problem]:
    """searchItems

    Args:
        q (str | Unset):
        type_ (str | Unset):
        locale (str | Unset):
        limit (str | Unset):
        page (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[ItemPageJSON | Problem]
    """

    kwargs = _get_kwargs(
        q=q,
        type_=type_,
        locale=locale,
        limit=limit,
        page=page,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    *,
    client: AuthenticatedClient,
    q: str | Unset = UNSET,
    type_: str | Unset = UNSET,
    locale: str | Unset = UNSET,
    limit: str | Unset = UNSET,
    page: str | Unset = UNSET,
) -> ItemPageJSON | Problem | None:
    """searchItems

    Args:
        q (str | Unset):
        type_ (str | Unset):
        locale (str | Unset):
        limit (str | Unset):
        page (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        ItemPageJSON | Problem
    """

    return (
        await asyncio_detailed(
            client=client,
            q=q,
            type_=type_,
            locale=locale,
            limit=limit,
            page=page,
        )
    ).parsed
