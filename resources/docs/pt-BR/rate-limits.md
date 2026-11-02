# Limites de requisições

Cada plano tem uma cota de requisições, um limite por segundo e um tamanho máximo de página.

## Planos

| Plano | Cota | Requisições / segundo | Tamanho máx. da página |
| --- | --- | --- | --- |
| Free | 100 no total | 1 (rajada de 5) | 10 |
| Pro | 50.000 / mês | 10 | 100 |
| Business | 500.000 / mês | 50 | 100 |

Preços e upgrades: [/pricing](/pricing).

## O que conta

- Toda requisição autenticada que chega a um endpoint conta como uma, inclusive as que retornam 404.
- Não contam: `401`, `400` por parâmetros inválidos, `429` e qualquer `5xx` (estornado).
- `GET /v1/me/usage` também conta, então leia os cabeçalhos de resposta das suas chamadas normais sempre que puder.
- As cotas dos planos pagos são renovadas às 00:00 UTC do dia 1º de cada mês. A cota do Free nunca é renovada.

## Cabeçalhos de resposta

As respostas bem-sucedidas informam em que ponto você está (campos de cabeçalho RateLimit do IETF).

| Cabeçalho | Significado |
| --- | --- |
| `RateLimit-Limit` | Requisições no período atual (a cota do seu plano). |
| `RateLimit-Remaining` | Requisições restantes antes do 402. |
| `RateLimit-Reset` | Segundos até a cota ser renovada; 0 para a cota vitalícia do Free. |
| `Retry-After` | Em 429 e 503: segundos de espera antes de tentar de novo. |

## Quando você atinge um limite

O plano Free aceita uma rajada de 5 requisições de uma vez e depois mantém 1 requisição por segundo; os outros planos permitem rajadas de até o seu limite por segundo.

Requisições por segundo em excesso retornam [`429 rate-limited`](/docs/errors#rate-limited): aguarde `Retry-After` segundos. Cota esgotada retorna [`402 quota-exhausted`](/docs/errors#quota-exhausted): faça upgrade do plano; tentar de novo não ajuda até a renovação.
