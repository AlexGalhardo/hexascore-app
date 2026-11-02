# Errores

Los errores usan los detalles de problema de la RFC 9457, con un código estable legible por máquina.

## Formato

Todo error tiene `Content-Type: application/problem+json`. Decide según `code` (estable) y no según `title` o `detail` (texto para personas que puede cambiar). `type` enlaza con la sección correspondiente más abajo, y `request_id` es igual al encabezado `X-Request-Id`: inclúyelo cuando contactes con soporte.

```json
{
  "type": "https://hexascore.app/problems/quota-exhausted",
  "title": "Quota exhausted",
  "status": 402,
  "detail": "Monthly quota is exhausted. Upgrade at https://hexascore.app/pricing.",
  "instance": "/api/v1/games",
  "request_id": "01JB7Q4V0Q8M3K2Y7T9H5N6C1D",
  "code": "quota-exhausted",
  "plan": "pro",
  "limit": 50000
}
```

## Códigos

### invalid-query

HTTP 400. Un parámetro de consulta es desconocido o está fuera de rango, o `limit` supera el tamaño de página de tu plan (10 en Free). `errors` enumera cada problema con su `path`. Corrige los parámetros indicados. No cuenta para tu cuota.

### invalid-cursor

HTTP 400. El `cursor` fue alterado o pertenece a otra consulta. Vuelve a empezar desde la primera página.

### changes-window-exceeded

HTTP 400. La ventana `since` de `/v1/changes` es mayor de lo que permite el endpoint. Pide un `since` más reciente.

### unauthorized

HTTP 401. Falta el encabezado `Authorization: Bearer`, la clave está mal formada, o la clave no existe o fue revocada. Revisa el encabezado y la clave en [/profile/api](/profile/api). Nunca se reintenta automáticamente.

### quota-exhausted

HTTP 402. Las 100 solicitudes vitalicias de Free, o la cuota mensual de los planes de pago. El cuerpo incluye `plan` y `limit`. Mejora tu plan en [/pricing](/pricing). Las cuotas de pago se restablecen el día 1 de cada mes (UTC).

### plan-required

HTTP 403. El endpoint requiere un plan superior. Mejora tu plan en [/pricing](/pricing).

### account-banned

HTTP 403. La cuenta propietaria de la clave está suspendida. Contáctanos a través de [/contact](/contact) con el `request_id`.

### not-found

HTTP 404. Id de ítem desconocido, un id que pertenece a otro tipo (por ejemplo, el id de un videojuego en `/movies`), una ruta de tipo desconocida, o ninguna coincidencia en `/lookup`. Las búsquedas exigen el id exacto.

### rate-limited

HTTP 429. Más solicitudes por segundo de las que permite tu plan. El encabezado `Retry-After` indica cuántos segundos esperar. No cuenta para tu cuota. Espacia las ráfagas o mejora tu plan para obtener un límite mayor.

### internal-error

HTTP 500. Algo falló de nuestro lado. La solicitud se reembolsa a tu cuota. Reintenta con backoff exponencial; si persiste, envíanos el `request_id`.

### community-submissions-closed

HTTP 403. Las sugerencias de la comunidad están desactivadas por ahora. Inténtalo de nuevo más tarde.

### suggestion-not-eligible

HTTP 403. Tu cuenta no puede enviar sugerencias (correo sin verificar, cuenta demasiado nueva, bloqueada para envíos, o la sugerencia no es tuya).

### suggestion-duplicate

HTTP 409. Ya existe una sugerencia abierta para el mismo ítem y los mismos campos. Edítala o retírala en lugar de crear otra.

### stale-before

HTTP 409. El campo cambió desde que lo cargaste. Recarga y vuelve a aplicar tu edición.

### invalid-transition

HTTP 409. La sugerencia no puede pasar de su estado actual al solicitado.

### field-locked

HTTP 423. Un moderador bloqueó este campo; no se aceptan ediciones de la comunidad en él.

### suggestion-quota

HTTP 429. Alcanzaste tu cuota diaria de sugerencias. Inténtalo de nuevo después del restablecimiento.

### suggestion-cooldown

HTTP 429. Espera a que termine el tiempo de espera; el encabezado `Retry-After` indica los segundos.

### validation

HTTP 422. El envío no superó la validación.

### service-unavailable

HTTP 503. Una dependencia está caída temporalmente. Incluye `Retry-After`. Reintenta después del plazo indicado.

## Reintentos

Reintenta solo `429`, `500` y `503`, respetando `Retry-After` cuando esté presente y, si no, con backoff exponencial. Todo endpoint es un `GET` de solo lectura, así que reintentar siempre es seguro.
