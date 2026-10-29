from http import HTTPStatus
from typing import Any

import httpx

from ...client import AuthenticatedClient, Client
from ...models.entity_page_json import EntityPageJSON
from ...models.problem import Problem
from ...types import UNSET, Response, Unset


def _get_kwargs(
    *,
    q: str | Unset = UNSET,
    kind: str | Unset = UNSET,
    locale: str | Unset = UNSET,
    limit: str | Unset = UNSET,
    page: str | Unset = UNSET,
) -> dict[str, Any]:

    params: dict[str, Any] = {}

    params["q"] = q

    params["kind"] = kind

    params["locale"] = locale

    params["limit"] = limit

    params["page"] = page

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: dict[str, Any] = {
        "method": "get",
        "url": "/api/v1/search/entities",
        "params": params,
    }

    return _kwargs


def _parse_response(*, client: AuthenticatedClient | Client, response: httpx.Response) -> EntityPageJSON | Problem:
    if response.status_code == 200:
        response_200 = EntityPageJSON.from_dict(response.json())

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
) -> Response[EntityPageJSON | Problem]:
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
    kind: str | Unset = UNSET,
    locale: str | Unset = UNSET,
    limit: str | Unset = UNSET,
    page: str | Unset = UNSET,
) -> Response[EntityPageJSON | Problem]:
    """searchEntities

    Args:
        q (str | Unset):
        kind (str | Unset):
        locale (str | Unset):
        limit (str | Unset):
        page (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[EntityPageJSON | Problem]
    """

    kwargs = _get_kwargs(
        q=q,
        kind=kind,
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
    kind: str | Unset = UNSET,
    locale: str | Unset = UNSET,
    limit: str | Unset = UNSET,
    page: str | Unset = UNSET,
) -> EntityPageJSON | Problem | None:
    """searchEntities

    Args:
        q (str | Unset):
        kind (str | Unset):
        locale (str | Unset):
        limit (str | Unset):
        page (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        EntityPageJSON | Problem
    """

    return sync_detailed(
        client=client,
        q=q,
        kind=kind,
        locale=locale,
        limit=limit,
        page=page,
    ).parsed


async def asyncio_detailed(
    *,
    client: AuthenticatedClient,
    q: str | Unset = UNSET,
    kind: str | Unset = UNSET,
    locale: str | Unset = UNSET,
    limit: str | Unset = UNSET,
    page: str | Unset = UNSET,
) -> Response[EntityPageJSON | Problem]:
    """searchEntities

    Args:
        q (str | Unset):
        kind (str | Unset):
        locale (str | Unset):
        limit (str | Unset):
        page (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[EntityPageJSON | Problem]
    """

    kwargs = _get_kwargs(
        q=q,
        kind=kind,
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
    kind: str | Unset = UNSET,
    locale: str | Unset = UNSET,
    limit: str | Unset = UNSET,
    page: str | Unset = UNSET,
) -> EntityPageJSON | Problem | None:
    """searchEntities

    Args:
        q (str | Unset):
        kind (str | Unset):
        locale (str | Unset):
        limit (str | Unset):
        page (str | Unset):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        EntityPageJSON | Problem
    """

    return (
        await asyncio_detailed(
            client=client,
            q=q,
            kind=kind,
            locale=locale,
            limit=limit,
            page=page,
        )
    ).parsed
