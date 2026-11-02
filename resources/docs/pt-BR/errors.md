# Erros

Os erros usam os detalhes de problema da RFC 9457, com um código estável legível por máquina.

## Formato

Todo erro tem `Content-Type: application/problem+json`. Trate a resposta pelo `code` (estável) e não por `title` ou `detail` (texto para humanos que pode mudar). `type` aponta para a seção correspondente abaixo, e `request_id` é igual ao cabeçalho `X-Request-Id`: informe-o ao falar com o suporte.

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

HTTP 400. Um parâmetro de consulta é desconhecido ou está fora do intervalo, ou `limit` é maior que o tamanho de página do seu plano (10 no Free). `errors` lista cada problema com o seu `path`. Corrija os parâmetros listados. Não conta na sua cota.

### invalid-cursor

HTTP 400. O `cursor` foi adulterado ou pertence a outra consulta. Recomece pela primeira página.

### changes-window-exceeded

HTTP 400. A janela `since` de `/v1/changes` é maior do que o endpoint permite. Peça um `since` mais recente.

### unauthorized

HTTP 401. Não há cabeçalho `Authorization: Bearer`, a chave é malformada, ou a chave não existe ou foi revogada. Confira o cabeçalho e a chave em [/profile/api](/profile/api). Nunca é repetida automaticamente.

### quota-exhausted

HTTP 402. As 100 requisições vitalícias do Free, ou a cota mensal dos planos pagos. O corpo inclui `plan` e `limit`. Faça upgrade do plano em [/pricing](/pricing). As cotas pagas são renovadas no dia 1º de cada mês (UTC).

### plan-required

HTTP 403. O endpoint exige um plano superior. Faça upgrade em [/pricing](/pricing).

### account-banned

HTTP 403. A conta dona da chave está suspensa. Fale conosco por [/contact](/contact) informando o `request_id`.

### not-found

HTTP 404. Id de item desconhecido, id que pertence a outro tipo (por exemplo, o id de um jogo em `/movies`), caminho de tipo desconhecido, ou nenhuma correspondência em `/lookup`. As buscas exigem o id exato.

### rate-limited

HTTP 429. Mais requisições por segundo do que o seu plano permite. O cabeçalho `Retry-After` diz quantos segundos esperar. Não conta na sua cota. Espaçe as rajadas ou faça upgrade para um limite maior.

### internal-error

HTTP 500. Algo falhou do nosso lado. A requisição é estornada na sua cota. Tente de novo com backoff exponencial; se persistir, envie-nos o `request_id`.

### community-submissions-closed

HTTP 403. As sugestões da comunidade estão desativadas no momento. Tente novamente mais tarde.

### suggestion-not-eligible

HTTP 403. Sua conta não pode enviar sugestões (e-mail não verificado, conta muito nova, bloqueada para envios, ou a sugestão não é sua).

### suggestion-duplicate

HTTP 409. Já existe uma sugestão aberta para o mesmo item e os mesmos campos. Edite-a ou retire-a em vez de criar outra.

### stale-before

HTTP 409. O campo mudou desde que você o carregou. Recarregue e aplique a edição de novo.

### invalid-transition

HTTP 409. A sugestão não pode passar do status atual para o solicitado.

### field-locked

HTTP 423. Um moderador bloqueou este campo; edições da comunidade nele não são aceitas.

### suggestion-quota

HTTP 429. Você atingiu a cota diária de sugestões. Tente de novo depois da renovação.

### suggestion-cooldown

HTTP 429. Aguarde o fim do intervalo de espera; o cabeçalho `Retry-After` informa os segundos.

### validation

HTTP 422. O envio falhou na validação.

### service-unavailable

HTTP 503. Uma dependência está temporariamente fora do ar. Vem com `Retry-After`. Tente de novo após o intervalo indicado.

## Novas tentativas

Repita apenas `429`, `500` e `503`, respeitando `Retry-After` quando presente e, caso contrário, usando backoff exponencial. Todo endpoint é um `GET` somente leitura, então repetir a chamada é sempre seguro.
