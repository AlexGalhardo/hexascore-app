from http import HTTPStatus
from typing import Any

import httpx

from ...client import AuthenticatedClient, Client
from ...models.change_page_json import ChangePageJSON
from ...models.problem import Problem
from ...types import UNSET, Response, Unset


def _get_kwargs(
    *,
    since: str | Unset = UNSET,
    type_: str | Unset = UNSET,
    cursor: str | Unset = UNSET,
    limit: str | Unset = UNSET,
) -> dict[str, Any]:

    params: dict[str, Any] = {}

    params["since"] = since

    params["type"] = type_

    params["cursor"] = cursor

    params["limit"] = limit

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/api/v1/changes",
        "params": params,
    }

    return _kwargs


def _parse_response(*, client: AuthenticatedClient | Client, response: httpx.Response) -> ChangePageJSON | Problem:
    if response.status_code == 200:
        response_200 = ChangePageJSON.from_dict(response.json())

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
) -> Response[ChangePageJSON | Problem]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    *,
    client: AuthenticatedClient,
    since: str | Unset = UNSET,
    type_: str | Unset = UNSET,
    cursor: str | Unset = UNSET,
    limit: str | Unset = UNSET,
) -> Response[ChangePageJSON | Problem]:
    """listChanges

    Args:
        since (str | Unset):
        type_ (str | Unset):
        cursor (str | Unset):
        limit (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[ChangePageJSON | Problem]
    """

    kwargs = _get_kwargs(
        since=since,
        type_=type_,
        cursor=cursor,
        limit=limit,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    *,
    client: AuthenticatedClient,
    since: str | Unset = UNSET,
    type_: str | Unset = UNSET,
    cursor: str | Unset = UNSET,
    limit: str | Unset = UNSET,
) -> ChangePageJSON | Problem | None:
    """listChanges

    Args:
        since (str | Unset):
        type_ (str | Unset):
        cursor (str | Unset):
        limit (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        ChangePageJSON | Problem
    """

    return sync_detailed(
        client=client,
        since=since,
        type_=type_,
        cursor=cursor,
        limit=limit,
    ).parsed


async def asyncio_detailed(
    *,
    client: AuthenticatedClient,
    since: str | Unset = UNSET,
    type_: str | Unset = UNSET,
    cursor: str | Unset = UNSET,
    limit: str | Unset = UNSET,
) -> Response[ChangePageJSON | Problem]:
    """listChanges

    Args:
        since (str | Unset):
        type_ (str | Unset):
        cursor (str | Unset):
        limit (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[ChangePageJSON | Problem]
    """

    kwargs = _get_kwargs(
        since=since,
        type_=type_,
        cursor=cursor,
        limit=limit,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    *,
    client: AuthenticatedClient,
    since: str | Unset = UNSET,
    type_: str | Unset = UNSET,
    cursor: str | Unset = UNSET,
    limit: str | Unset = UNSET,
) -> ChangePageJSON | Problem | None:
    """listChanges

    Args:
        since (str | Unset):
        type_ (str | Unset):
        cursor (str | Unset):
        limit (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        ChangePageJSON | Problem
    """

    return (
        await asyncio_detailed(
            client=client,
            since=since,
            type_=type_,
            cursor=cursor,
            limit=limit,
        )
    ).parsed
