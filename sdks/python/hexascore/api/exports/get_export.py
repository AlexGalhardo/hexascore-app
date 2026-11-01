from http import HTTPStatus
from typing import Any, cast
from urllib.parse import quote

import httpx

from ...client import AuthenticatedClient, Client
from ...models.get_export_type import GetExportType
from ...models.problem import Problem
from ...types import Response


def _get_kwargs(
    type_: GetExportType,
    day: str,
) -> dict[str, Any]:

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/exports/{type_}/{day}".format(
            type_=quote(str(type_), safe=""),
            day=quote(str(day), safe=""),
        ),
    }

    return _kwargs


def _parse_response(*, client: AuthenticatedClient | Client, response: httpx.Response) -> Any | Problem:
    if response.status_code == 204:
        response_204 = cast(Any, None)
        return response_204

    if response.status_code == 401:
        response_401 = Problem.from_dict(response.json())

        return response_401

    if response.status_code == 422:
        response_422 = Problem.from_dict(response.json())

        return response_422

    response_default = Problem.from_dict(response.json())

    return response_default


def _build_response(*, client: AuthenticatedClient | Client, response: httpx.Response) -> Response[Any | Problem]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    type_: GetExportType,
    day: str,
    *,
    client: AuthenticatedClient,
) -> Response[Any | Problem]:
    """getExport

    Args:
        type_ (GetExportType):
        day (str):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Any | Problem]
    """

    kwargs = _get_kwargs(
        type_=type_,
        day=day,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    type_: GetExportType,
    day: str,
    *,
    client: AuthenticatedClient,
) -> Any | Problem | None:
    """getExport

    Args:
        type_ (GetExportType):
        day (str):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Any | Problem
    """

    return sync_detailed(
        type_=type_,
        day=day,
        client=client,
    ).parsed


async def asyncio_detailed(
    type_: GetExportType,
    day: str,
    *,
    client: AuthenticatedClient,
) -> Response[Any | Problem]:
    """getExport

    Args:
        type_ (GetExportType):
        day (str):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Any | Problem]
    """

    kwargs = _get_kwargs(
        type_=type_,
        day=day,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    type_: GetExportType,
    day: str,
    *,
    client: AuthenticatedClient,
) -> Any | Problem | None:
    """getExport

    Args:
        type_ (GetExportType):
        day (str):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Any | Problem
    """

    return (
        await asyncio_detailed(
            type_=type_,
            day=day,
            client=client,
        )
    ).parsed
