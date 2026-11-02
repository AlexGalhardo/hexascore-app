# Paginação

Os endpoints de lista usam cursores opacos: estáveis diante de inserções e rápidos em qualquer profundidade.

## Como funciona

`GET /v1/games`, `/v1/movies`, `/v1/series` e `/v1/books` retornam `{ "data": [...], "meta": { "next_cursor": "..." } }`. Para obter a próxima página, repita a mesma requisição com `cursor=<meta.next_cursor>`. Quando `meta.next_cursor` não estiver presente, você chegou ao fim.

- Mantenha todos os outros parâmetros idênticos. Um cursor de outro `sort` retorna `400`.
- Os cursores são opacos: não os interprete nem os construa, o formato pode mudar.
- A ordenação é estável: os empates são desfeitos pelo id do item, e itens sem nota ou sem data sempre vêm por último.

## Tamanho da página

`limit` tem padrão 10. O máximo depende do seu plano: 10 no Free, 100 no Pro e no Business. Um valor maior retorna `400 invalid-query`; a API nunca trunca em silêncio. Cada página conta como uma requisição.

## Exemplo: percorrer todas as páginas

```js
let cursor = null;
do {
  const url = new URL("https://api.hexascore.app/v1/games");
  url.searchParams.set("limit", "100");
  if (cursor) url.searchParams.set("cursor", cursor);
  const res = await fetch(url, { headers: { Authorization: `Bearer ${process.env.HEXASCORE_API_KEY}` } });
  const page = await res.json();
  for (const game of page.data) console.log(game.title);
  cursor = page.meta.next_cursor ?? null;
} while (cursor);
```

## Busca

`GET /v1/search` não é paginado: retorna as melhores correspondências (até 50) para uma consulta, da melhor para a pior.
