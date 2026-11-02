# Início rápido

Notas da comunidade para jogos, filmes, séries e livros, de 0 a 10 em seis categorias por item, por meio de uma API JSON somente leitura.

## 1. Crie uma chave de API

Abra o seu [painel da API](/profile/api) e crie uma chave. Ela começa com `hx_live_` e é exibida uma única vez, então guarde-a em uma variável de ambiente e nunca em código do lado do cliente.

O plano Free inclui 100 requisições no total, o suficiente para desenvolver e testar uma integração.

## 2. Faça sua primeira requisição

A URL base é `https://api.hexascore.app/v1`. Todo endpoint é um `GET`. Veja os cinco jogos mais bem avaliados:

```sh
curl -H "Authorization: Bearer $HEXASCORE_API_KEY" \
  "https://api.hexascore.app/v1/games?sort=score&limit=5"
```

## 3. Leia a resposta

As listas retornam `data` mais `meta.next_cursor` para a próxima página. Os nomes dos campos usam snake_case e as datas seguem o formato ISO 8601. `score` e `categories` são `null` até que um item tenha 5 avaliações.

## Próximos passos

- Busque um item com `GET /v1/games/{id}`.
- Mapeie os ids que você já tem com `GET /v1/lookup`.
- Trate os [erros](/docs/errors) e os [limites de requisições](/docs/rate-limits) e depois explore a [referência da API](/docs/reference).
