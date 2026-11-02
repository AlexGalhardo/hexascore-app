# Webhooks

Recibe una solicitud HTTP firmada cuando cambien los ítems, en lugar de hacer polling.

## Configuración

Los webhooks están disponibles en los planes Hobby (2 endpoints) y PRO (10 endpoints). Crea un endpoint con `POST /v1/webhooks`:

| Campo | Significado |
| --- | --- |
| `url` | Una URL `https` que responda en pocos segundos. Se rechazan las direcciones privadas e internas. |
| `events` | Uno o más de `item.created`, `item.updated`, `item.deleted`. |
| `itemTypes` | Uno o más de `games`, `movies`, `series`, `books`. |

La respuesta contiene el `secret` de firma, que empieza con `whsec_`. Se muestra una sola vez: guárdalo en tu gestor de secretos. Usa `GET /v1/webhooks` para listar los endpoints, `DELETE /v1/webhooks/{id}` para eliminar uno y `POST /v1/webhooks/{id}/test` para enviar una entrega de prueba. Administrar webhooks no cuenta para tu cuota.

## Qué enviamos

Enviamos un lote de hasta 100 eventos en un único `POST` con cuerpo JSON. Los payloads solo llevan ids y nombres de campos; vuelve a obtener el ítem para leer sus datos.

```json
{
  "type": "item.batch",
  "deliveredAt": "2026-11-02T10:15:00Z",
  "events": [
    {
      "id": "chg_1048576",
      "type": "item.updated",
      "itemType": "games",
      "itemId": "0192f3a4-7b1c-7e21-9c55-3f1e2d4a6b71",
      "fields": ["score", "ratingCount"],
      "changedAt": "2026-11-02T10:14:12Z"
    },
    {
      "id": "chg_1048577",
      "type": "item.deleted",
      "itemType": "movies",
      "itemId": "0192f3a4-7b1c-7e21-9c55-3f1e2d4a5521",
      "mergedInto": "0192f3a4-7b1c-7e21-9c55-3f1e2d4a5520",
      "changedAt": "2026-11-02T10:14:30Z"
    }
  ]
}
```

Los eventos del mismo ítem dentro de una ventana de entrega se combinan. Los ids de evento son únicos, así que úsalos para ignorar una repetición.

## Verifica la firma

Las entregas siguen el estándar [Standard Webhooks](https://www.standardwebhooks.com). Cada solicitud tiene tres encabezados: `webhook-id`, `webhook-timestamp` (segundos Unix) y `webhook-signature` (uno o más valores `v1,<base64>` separados por espacios).

Para verificar una entrega:

- Lee el cuerpo sin procesar de la solicitud. No lo analices ni lo vuelvas a serializar.
- Rechaza la solicitud cuando `webhook-timestamp` esté a más de 5 minutos de tu reloj.
- Calcula el HMAC-SHA256 en base64 de `<webhook-id>.<webhook-timestamp>.<cuerpo sin procesar>`, usando como clave el secreto sin el prefijo `whsec_`, decodificado desde base64.
- Compáralo con cada valor `v1` en tiempo constante. Acepta la solicitud cuando uno coincida.

```ts
import { createHmac, timingSafeEqual } from "node:crypto";

export function verifyWebhook(secret: string, headers: Headers, rawBody: string): boolean {
	const id = headers.get("webhook-id");
	const timestamp = headers.get("webhook-timestamp");
	const signatures = headers.get("webhook-signature");
	if (!id || !timestamp || !signatures) return false;
	if (Math.abs(Date.now() / 1000 - Number(timestamp)) > 300) return false;

	const key = Buffer.from(secret.replace(/^whsec_/, ""), "base64");
	const expected = createHmac("sha256", key).update(`${id}.${timestamp}.${rawBody}`).digest();
	return signatures.split(" ").some((entry) => {
		const [version, value] = entry.split(",");
		if (version !== "v1" || !value) return false;
		const got = Buffer.from(value, "base64");
		return got.length === expected.length && timingSafeEqual(got, expected);
	});
}
```

```python
import base64
import binascii
import hashlib
import hmac
import time


def verify_webhook(secret: str, headers: dict[str, str], raw_body: bytes) -> bool:
    try:
        msg_id = headers["webhook-id"]
        timestamp = headers["webhook-timestamp"]
        signatures = headers["webhook-signature"]
        if abs(time.time() - int(timestamp)) > 300:
            return False
    except (KeyError, ValueError):
        return False

    key = base64.b64decode(secret.removeprefix("whsec_"))
    signed = f"{msg_id}.{timestamp}.".encode() + raw_body
    expected = hmac.new(key, signed, hashlib.sha256).digest()
    for entry in signatures.split(" "):
        version, _, value = entry.partition(",")
        try:
            got = base64.b64decode(value)
        except binascii.Error:
            continue
        if version == "v1" and hmac.compare_digest(got, expected):
            return True
    return False
```

## Reintentos y fallos

Responde con cualquier estado `2xx` en pocos segundos y haz el trabajo real después. Reintentamos las demás respuestas y los timeouts 8 veces, con intervalos crecientes a partir de 30 segundos. Un `410 Gone` detiene la entrega definitivamente. Cuando tres entregas seguidas agotan sus reintentos, el endpoint se desactiva y enviamos un correo al dueño de la cuenta. Las entregas se registran durante 30 días.

La entrega es al menos una vez y sin orden estricto. Los webhooks aún pueden perder eventos, así que llama a [`/v1/changes`](/docs/changes) con el `since` de tu último evento procesado para cubrir cualquier vacío.
