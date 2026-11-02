# Autenticación

Cada solicitud se autentica con una clave de API secreta enviada como token Bearer.

## Token Bearer

Envía la clave en el encabezado `Authorization`. Las claves tienen el formato `hx_live_` seguido de 32 letras y dígitos. No se aceptan claves en la query string, de modo que nunca terminan en los registros de acceso ni en el historial del navegador.

```sh
curl -H "Authorization: Bearer hx_live_..." https://api.hexascore.app/v1/games
```

Una clave ausente, mal formada, desconocida o revocada devuelve `401` con el código [`unauthorized`](/docs/errors#unauthorized) y un encabezado `WWW-Authenticate: Bearer`.

## Administrar claves

- Crea y revoca claves en [/profile/api](/profile/api) (se requiere correo verificado; hasta 5 claves activas).
- La clave completa se muestra una sola vez. Solo guardamos un hash SHA-256, así que no podemos recuperarla por ti.
- La revocación es inmediata. Para rotar, crea una clave nueva, despliégala y luego revoca la anterior.
- Todas las claves de una cuenta comparten el mismo plan, la misma cuota y el mismo límite por segundo.

## Mantén las claves en secreto

Llama a la API desde tu servidor. Una clave incrustada en una página web o en una app móvil puede ser extraída y usada para agotar tu cuota. Si una clave se filtra, revócala de inmediato.
