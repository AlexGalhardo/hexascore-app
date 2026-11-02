# Autenticação

Toda requisição é autenticada com uma chave de API secreta enviada como token Bearer.

## Token Bearer

Envie a chave no cabeçalho `Authorization`. As chaves têm o formato `hx_live_` seguido de 32 letras e dígitos. Chaves na query string não são aceitas, para que nunca apareçam nos logs de acesso nem no histórico do navegador.

```sh
curl -H "Authorization: Bearer hx_live_..." https://api.hexascore.app/v1/games
```

Uma chave ausente, malformada, desconhecida ou revogada retorna `401` com o código [`unauthorized`](/docs/errors#unauthorized) e um cabeçalho `WWW-Authenticate: Bearer`.

## Gerenciando chaves

- Crie e revogue chaves em [/profile/api](/profile/api) (é preciso ter o e-mail verificado; até 5 chaves ativas).
- A chave completa é exibida uma única vez. Guardamos apenas um hash SHA-256, então não conseguimos recuperá-la para você.
- A revogação é imediata. Para trocar a chave, crie uma nova, publique-a e só então revogue a antiga.
- Todas as chaves de uma conta compartilham o mesmo plano, a mesma cota e o mesmo limite por segundo.

## Mantenha as chaves em segredo

Chame a API a partir do seu servidor. Uma chave embutida em uma página web ou em um app móvel pode ser extraída e usada para esgotar a sua cota. Se uma chave vazar, revogue-a imediatamente.
