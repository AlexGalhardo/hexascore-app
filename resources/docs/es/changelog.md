# Changelog

La versión forma parte de la ruta. La v1 solo cambia de forma retrocompatible.

## Política de versionado

Dentro de `/v1` solo hacemos cambios aditivos: nuevos endpoints, nuevos parámetros opcionales, nuevos campos de respuesta y nuevos textos de `detail` en los errores. Ignora los campos desconocidos y no dependas del orden de los campos.

Los cambios incompatibles (eliminar o renombrar un campo, cambiar un tipo, hacer obligatorio un parámetro) se publican en una nueva ruta (`/v2`). Los valores de `code` de los errores, las claves de categoría y los nombres de ordenamiento forman parte del contrato.

## Obsolescencia y cierre

- Una versión obsoleta sigue funcionando al menos 6 meses después del lanzamiento de su sucesora.
- Sus respuestas incluyen los encabezados `Deprecation` (RFC 9745) y `Sunset` (RFC 8594) con la fecha en que deja de funcionar, además de un `Link` a esta página.
- A los dueños de cuenta con tráfico reciente en la versión anterior se les envía un correo al declararla obsoleta y 30 días antes del cierre.

## Cambios

### v1: primera versión

- Listas, detalles y géneros para `games`, `movies`, `series` y `books`.
- `/search`, `/lookup` por id externo y `/me/usage`.
