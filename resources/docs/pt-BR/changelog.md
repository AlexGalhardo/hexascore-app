# Changelog

A versão faz parte do caminho. A v1 só muda de forma retrocompatível.

## Política de versionamento

Dentro de `/v1` fazemos apenas mudanças aditivas: novos endpoints, novos parâmetros opcionais, novos campos de resposta e novos textos de `detail` nos erros. Ignore campos desconhecidos e não dependa da ordem dos campos.

Mudanças incompatíveis (remover ou renomear um campo, mudar um tipo, tornar um parâmetro obrigatório) são lançadas em um novo caminho (`/v2`). Os valores de `code` dos erros, as chaves de categoria e os nomes de ordenação fazem parte do contrato.

## Descontinuação e encerramento

- Uma versão descontinuada continua funcionando por pelo menos 6 meses depois do lançamento da sucessora.
- As respostas dela trazem os cabeçalhos `Deprecation` (RFC 9745) e `Sunset` (RFC 8594) com a data em que deixa de funcionar, além de um `Link` para esta página.
- Os donos de conta com tráfego recente na versão antiga recebem e-mail na descontinuação e 30 dias antes do encerramento.

## Mudanças

### v1: primeira versão

- Listas, detalhes e gêneros para `games`, `movies`, `series` e `books`.
- `/search`, `/lookup` por id externo e `/me/usage`.
