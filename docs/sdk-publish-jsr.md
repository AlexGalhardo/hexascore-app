# Publicar o SDK TypeScript no JSR

Passo a passo para publicar `@hexascore/sdk` em <https://jsr.io>, a partir da mesma pasta `sdks/ts` que vai para o npm ([`sdk-publish-npm.md`](sdk-publish-npm.md)).

| | |
| --- | --- |
| Pacote | `@hexascore/sdk` |
| Pasta | `sdks/ts` |
| Configuração | `sdks/ts/jsr.json` |
| O que é publicado | o **código-fonte TypeScript** (`src/**/*.ts`, sem os testes), `README.md`, `LICENSE`, `jsr.json` |
| Versão | campo `version` de `sdks/ts/jsr.json` (hoje `0.1.0`) |
| Estado | ainda não publicado; o escopo `@hexascore` estava livre em 2026-10-10 |

## O que é o JSR e no que difere do npm

| | npm | JSR |
| --- | --- | --- |
| O que você envia | JavaScript já compilado (`dist/`) + `.d.ts` | o TypeScript original; o JSR compila e gera os tipos |
| Passo de build | obrigatório (`bun run build`) | nenhum |
| Login | usuário e senha + 2FA | conta do GitHub |
| Runtimes | Node, Bun e o que usar `node_modules` | Deno nativo; Node, Bun e outros pelo CLI `jsr` |
| Documentação | o README | gerada sozinha a partir dos comentários JSDoc |
| Módulos | ESM ou CommonJS | só ESM |

Publicar nos dois é o normal: o npm alcança todo mundo, o JSR dá a melhor experiência em Deno e documentação automática.

## Regras do JSR que este pacote já atende

- **Imports relativos com extensão.** O JSR exige `./arquivo.ts`, não `./arquivo`. O gerador está configurado com `importFileExtension: ".ts"` (`sdks/ts/openapi-ts.config.ts`), e o build do npm reescreve para `.js` (`rewriteRelativeImportExtensions` no `tsconfig.json`).
- **Sem "slow types".** Toda função e constante exportada precisa de tipo explícito, para o JSR gerar os `.d.ts` sem rodar o compilador inteiro. O ensaio do passo 2 confere isso.
- **Só ESM**, sem `require`.

## 1. Preparar a conta (uma vez)

1. Entre em <https://jsr.io> com **Sign in with GitHub**. Não há senha nem cadastro separado.
2. Crie o escopo: <https://jsr.io/new>, campo *Scope* = `hexascore`. O escopo é gratuito e pertence à sua conta; depois dá para convidar outras pessoas em *Scope → Members*.
3. Na mesma tela, crie o pacote: *Package name* = `sdk`. Resultado: `@hexascore/sdk`.
4. (Opcional, para publicar por automação) Em `jsr.io/@hexascore/sdk` → *Settings → GitHub repository*, vincule `AlexGalhardo/hexascore-app`.

## 2. Ensaiar a publicação

Na raiz do repositório:

```sh
bun run openapi:check     # o contrato commitado está em dia com o código
bun run sdk:generate      # regenera sdks/ts/src/gen
```

Em `sdks/ts`:

```sh
bun run typecheck
bun test
bunx jsr@0.13.5 publish --dry-run
```

O `--dry-run` não envia nada e não pede login. Ele:

- confere os "slow types" (`Checking for slow types in the public API...`);
- lista os arquivos que seriam enviados: devem ser `LICENSE`, `README.md`, `jsr.json` e `src/**/*.ts`, sem `*.test.ts`;
- termina com `Success Dry run complete`.

Se houver alterações não commitadas, o CLI para com `Aborting due to uncommitted changes`. Commite antes, ou use `--allow-dirty` só para o ensaio.

## 3. Escolher a versão

Edite `version` em `sdks/ts/jsr.json`. Use a mesma de `sdks/ts/package.json` e `sdks/python/pyproject.toml`. Regra de SemVer em [`sdk-publish-npm.md`](sdk-publish-npm.md#3-escolher-a-versão).

No JSR uma versão publicada é **imutável**: não dá para reenviar nem apagar, só retirar (*yank*).

## 4. Publicar

```sh
cd sdks/ts
bunx jsr@0.13.5 publish
```

1. O CLI mostra um link e abre o navegador em `jsr.io/auth?code=...`.
2. Confira o pacote e a versão na tela e clique em **Approve**. A autorização vale só para essa publicação.
3. O terminal confirma a publicação de `@hexascore/sdk@0.1.0`.

Conferir: <https://jsr.io/@hexascore/sdk>. A aba *Docs* já mostra a documentação gerada dos comentários JSDoc, e a aba *Score* mostra a nota do pacote.

## 5. Testar a instalação

Deno:

```sh
deno add jsr:@hexascore/sdk
```

```ts
import { createHexascoreClient, listGames } from "@hexascore/sdk";
```

Node, Bun, pnpm e yarn (o CLI configura o registro compatível com npm do JSR e instala):

```sh
npx jsr add @hexascore/sdk
bunx jsr add @hexascore/sdk
pnpm dlx jsr add @hexascore/sdk
```

## 6. Melhorar a nota do pacote (JSR Score)

A nota aparece na página e pesa na busca. O que conta:

| Critério | Estado |
| --- | --- |
| Tem README | sim |
| Sem slow types | sim (conferido no dry-run) |
| Descrição do pacote | preencha em *Settings → Description* no site |
| Runtimes compatíveis marcados | marque em *Settings → Runtime compatibility* (Deno, Node, Bun, browsers, Workers) |
| Símbolos exportados documentados com JSDoc | parcial: o código gerado traz as descrições do OpenAPI |
| Proveniência (publicado por CI com OIDC) | não, enquanto a publicação for manual |

## Publicar por automação (opcional)

Com o GitHub Actions desativado neste repositório, a publicação é manual. Quando for automatizar, o JSR dispensa token: com o repositório vinculado (passo 1.4), um job com `permissions: id-token: write` roda `npx jsr publish` em `sdks/ts` e o JSR autentica pelo OIDC do GitHub, gravando a proveniência.

## Erros comuns

As mensagens exatas variam com a versão do CLI; a tabela descreve o sintoma.

| Sintoma | Causa | Solução |
| --- | --- | --- |
| `Aborting due to uncommitted changes` | árvore de trabalho suja | commite, ou `--allow-dirty` no ensaio |
| Erro de "slow types" apontando uma função ou constante exportada | falta o tipo explícito na API pública | declare o tipo de retorno (ou o tipo da constante) |
| Erro sobre import relativo sem extensão | import `./arquivo` em vez de `./arquivo.ts` | rode `bun run sdk:generate` (o gerador já emite a extensão) |
| Escopo ou pacote inexistente | `@hexascore` ou `sdk` ainda não criados no site | passos 1.2 e 1.3 |
| Versão já existe | essa versão já foi publicada | suba a versão em `jsr.json` |
| Sem autorização para publicar | conta do GitHub sem acesso ao escopo | entre com a conta dona do escopo ou peça convite |

## Corrigir uma publicação errada

Não existe despublicar. Em `jsr.io/@hexascore/sdk/versions`, use **Yank** na versão com problema: ela continua disponível para quem já a fixou, mas deixa de ser resolvida por padrão. Depois publique a versão corrigida.

## Checklist

- [ ] escopo `@hexascore` e pacote `sdk` criados no jsr.io
- [ ] `bun run openapi:check` e `bun run sdk:generate` sem diferenças não commitadas
- [ ] versão igual em `jsr.json`, `package.json` e `pyproject.toml`
- [ ] `bun run typecheck` e `bun test` passando
- [ ] `bunx jsr@0.13.5 publish --dry-run` termina em `Success Dry run complete`
- [ ] `bunx jsr@0.13.5 publish` aprovado no navegador
- [ ] descrição e runtimes preenchidos no site
- [ ] `CHANGELOG.md`, `PLAN.md`
