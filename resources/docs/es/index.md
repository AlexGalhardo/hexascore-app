# Inicio rápido

Puntuaciones de la comunidad para videojuegos, películas, series y libros, de 0 a 10 en seis categorías por ítem, mediante una API JSON de solo lectura.

## 1. Crea una clave de API

Abre tu [panel de la API](/profile/api) y crea una clave. Empieza con `hx_live_` y se muestra una sola vez, así que guárdala en una variable de entorno y nunca en código del lado del cliente.

El plan Free incluye 100 solicitudes en total, suficiente para desarrollar y probar una integración.

## 2. Haz tu primera solicitud

La URL base es `https://api.hexascore.app/v1`. Todos los endpoints son `GET`. Estos son los cinco videojuegos mejor valorados:

```sh
curl -H "Authorization: Bearer $HEXASCORE_API_KEY" \
  "https://api.hexascore.app/v1/games?sort=score&limit=5"
```

## 3. Lee la respuesta

Las listas devuelven `data` más `meta.next_cursor` para la página siguiente. Los nombres de los campos usan snake_case y las fechas siguen el formato ISO 8601. `score` y `categories` son `null` hasta que un ítem tiene 5 valoraciones.

## Siguientes pasos

- Obtén un ítem con `GET /v1/games/{id}`.
- Relaciona los ids que ya tienes con `GET /v1/lookup`.
- Maneja los [errores](/docs/errors) y los [límites de solicitudes](/docs/rate-limits) y luego explora la [referencia de la API](/docs/reference).
