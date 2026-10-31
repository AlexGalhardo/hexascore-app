from http import HTTPStatus
from typing import Any
from urllib.parse import quote

import httpx

from ...client import AuthenticatedClient, Client
from ...models.activity_json import ActivityJSON
from ...models.get_item_activity_type import GetItemActivityType
from ...models.problem import Problem
from ...types import UNSET, Response, Unset


def _get_kwargs(
    type_: GetItemActivityType,
    id: str,
    *,
    from_: str | Unset = UNSET,
    to: str | Unset = UNSET,
) -> dict[str, Any]:

    params: dict[str, Any] = {}

    params["from"] = from_

    params["to"] = to

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/{type_}/{id}/activity".format(
            type_=quote(str(type_), safe=""),
            id=quote(str(id), safe=""),
        ),
        "params": params,
    }

    return _kwargs


def _parse_response(*, client: AuthenticatedClient | Client, response: httpx.Response) -> ActivityJSON | Problem:
    if response.status_code == 200:
        response_200 = ActivityJSON.from_dict(response.json())

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
) -> Response[ActivityJSON | Problem]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    type_: GetItemActivityType,
    id: str,
    *,
    client: AuthenticatedClient,
    from_: str | Unset = UNSET,
    to: str | Unset = UNSET,
) -> Response[ActivityJSON | Problem]:
    """getItemActivity

    Args:
        type_ (GetItemActivityType):
        id (str):
        from_ (str | Unset):
        to (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[ActivityJSON | Problem]
    """

    kwargs = _get_kwargs(
        type_=type_,
        id=id,
        from_=from_,
        to=to,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    type_: GetItemActivityType,
    id: str,
    *,
    client: AuthenticatedClient,
    from_: str | Unset = UNSET,
    to: str | Unset = UNSET,
) -> ActivityJSON | Problem | None:
    """getItemActivity

    Args:
        type_ (GetItemActivityType):
        id (str):
        from_ (str | Unset):
        to (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        ActivityJSON | Problem
    """

    return sync_detailed(
        type_=type_,
        id=id,
        client=client,
        from_=from_,
        to=to,
    ).parsed


async def asyncio_detailed(
    type_: GetItemActivityType,
    id: str,
    *,
    client: AuthenticatedClient,
    from_: str | Unset = UNSET,
    to: str | Unset = UNSET,
) -> Response[ActivityJSON | Problem]:
    """getItemActivity

    Args:
        type_ (GetItemActivityType):
        id (str):
        from_ (str | Unset):
        to (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[ActivityJSON | Problem]
    """

    kwargs = _get_kwargs(
        type_=type_,
        id=id,
        from_=from_,
        to=to,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    type_: GetItemActivityType,
    id: str,
    *,
    client: AuthenticatedClient,
    from_: str | Unset = UNSET,
    to: str | Unset = UNSET,
) -> ActivityJSON | Problem | None:
    """getItemActivity

    Args:
        type_ (GetItemActivityType):
        id (str):
        from_ (str | Unset):
        to (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        ActivityJSON | Problem
    """

    return (
        await asyncio_detailed(
            type_=type_,
            id=id,
            client=client,
            from_=from_,
            to=to,
        )
    ).parsed
