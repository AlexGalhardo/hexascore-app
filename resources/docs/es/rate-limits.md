# Límites de solicitudes

Cada plan tiene una cuota de solicitudes, un límite por segundo y un tamaño máximo de página.

## Planes

| Plan | Cuota | Solicitudes / segundo | Tamaño máx. de página |
| --- | --- | --- | --- |
| Free | 100 en total | 1 (ráfaga de 5) | 10 |
| Hobby | 10.000 / mes | 1 (ráfaga de 5) | 20 |
| PRO | 50.000 / mes | 3 (ráfaga de 15) | 50 |

Precios y mejoras de plan: [/pricing](/pricing).

## Qué cuenta

- Toda solicitud autenticada que llega a un endpoint cuenta como una, incluidas las que devuelven 404.
- No cuentan: `401`, `400` por parámetros inválidos, `429` ni ningún `5xx` (se reembolsa).
- `GET /v1/me/usage` también cuenta, así que lee los encabezados de respuesta de tus llamadas habituales siempre que puedas.
- Las cuotas de pago se restablecen a las 00:00 UTC del día 1 de cada mes. La cuota Free nunca se restablece.

## Encabezados de respuesta

Las respuestas exitosas te indican en qué punto estás (campos de encabezado RateLimit del IETF).

| Encabezado | Significado |
| --- | --- |
| `RateLimit-Limit` | Solicitudes del período actual (la cuota de tu plan). |
| `RateLimit-Remaining` | Solicitudes restantes antes del 402. |
| `RateLimit-Reset` | Segundos hasta que se restablece la cuota; 0 para la cuota vitalicia de Free. |
| `Retry-After` | En 429 y 503: segundos de espera antes de reintentar. |

## Cuando alcanzas un límite

El plan Free acepta una ráfaga de 5 solicitudes a la vez y luego mantiene 1 solicitud por segundo; los demás planes permiten ráfagas de hasta su límite por segundo.

Demasiadas solicitudes por segundo devuelven [`429 rate-limited`](/docs/errors#rate-limited): espera `Retry-After` segundos. Una cuota agotada devuelve [`402 quota-exhausted`](/docs/errors#quota-exhausted): mejora tu plan; reintentar no servirá hasta el restablecimiento.
