# Publicar o SDK TypeScript no npm

Passo a passo para publicar `@hexascore/sdk` (pasta `sdks/ts`) em <https://www.npmjs.com>. O mesmo código-fonte vai para o JSR: ver [`sdk-publish-jsr.md`](sdk-publish-jsr.md). O SDK Python está em [`sdk-publish-pypi.md`](sdk-publish-pypi.md).

| | |
| --- | --- |
| Pacote | `@hexascore/sdk` |
| Pasta | `sdks/ts` |
| O que é publicado | `dist/` (JavaScript ESM + `.d.ts`), `README.md`, `LICENSE`, `package.json` |
| Versão | campo `version` de `sdks/ts/package.json` (hoje `0.1.0`) |
| Estado | ainda não publicado; o nome `@hexascore/sdk` estava livre em 2026-10-10 |

## 1. Preparar a conta (uma vez)

1. Crie a conta em <https://www.npmjs.com/signup> e confirme o e-mail.
2. Ative a autenticação em dois fatores: *Account → Two-Factor Authentication*. Escolha proteger **login e publicação** (`auth-and-writes`). Sem 2FA o npm não deixa publicar pacotes novos.
3. Crie a organização que dá direito ao escopo `@hexascore`: <https://www.npmjs.com/org/create>, nome `hexascore`, plano gratuito (*Unlimited public packages*). O escopo de um pacote é sempre o nome de um usuário ou de uma organização; sem a organização `hexascore`, `npm publish` responde `404 Not Found - Scope not found`.
4. Faça login na máquina:

   ```sh
   npm login
   npm whoami        # deve mostrar seu usuário
   ```

## 2. Conferir o pacote antes de publicar

Na raiz do repositório:

```sh
bun run openapi:check     # o contrato commitado está em dia com o código
bun run sdk:generate      # regenera sdks/ts/src/gen a partir de apps/web/public/openapi.json
```

Em `sdks/ts`:

```sh
bun run typecheck         # tsc --noEmit
bun test                  # 3 testes com fetch simulado
bun run build             # gera dist/
npm pack --dry-run        # lista exatamente o que iria para o registro
```

O `npm pack --dry-run` deve listar só `dist/**`, `README.md`, `LICENSE` e `package.json`, com `name: @hexascore/sdk` e a versão certa. Nenhum arquivo `*.test.*` e nada de `src/`.

Para testar o pacote de verdade em outro projeto antes de publicar:

```sh
npm pack                                   # cria hexascore-sdk-0.1.0.tgz
cd /caminho/de/um/projeto-de-teste
npm install /caminho/para/hexascore-sdk-0.1.0.tgz
```

## 3. Escolher a versão

O SDK tem versão própria, independente da versão do site. Regra (SemVer):

| Mudança | Exemplo | Bump |
| --- | --- | --- |
| Correção sem mudar a API pública | corrigir README, tipo errado | patch: `0.1.0` → `0.1.1` |
| Endpoint ou campo novo, compatível | novo `listAwards` | minor: `0.1.0` → `0.2.0` |
| Remoção ou mudança incompatível | campo renomeado | major (antes da `1.0.0`: minor) |

Mantenha a **mesma versão** nos três lugares: `sdks/ts/package.json`, `sdks/ts/jsr.json` e `sdks/python/pyproject.toml`. Para o npm:

```sh
cd sdks/ts
npm version 0.1.1 --no-git-tag-version    # só edita o package.json
```

Depois ajuste `jsr.json` à mão e commite.

## 4. Publicar

```sh
cd sdks/ts
npm publish
```

- `prepublishOnly` roda `typecheck`, os testes e o `build` antes de enviar; se algo falhar, nada é publicado.
- `publishConfig.access` já é `public` no `package.json`. Pacote com escopo é privado por padrão; sem isso o npm responderia `402 Payment Required`.
- O npm pede o código do 2FA (`--otp=123456` passa direto).

Conferir:

```sh
npm view @hexascore/sdk version dist-tags
```

e a página <https://www.npmjs.com/package/@hexascore/sdk>.

## 5. Depois de publicar

1. Tire do `sdks/ts/README.md` qualquer aviso de "ainda não publicado" (hoje não há).
2. Marque o commit: `git tag sdk-v0.1.0 && git push origin sdk-v0.1.0`.
3. Registre no `CHANGELOG.md` e marque a pendência no `PLAN.md`.

## Versões de teste (pré-release)

Para publicar sem virar a versão padrão de quem instala:

```sh
npm version 0.2.0-beta.0 --no-git-tag-version
npm publish --tag beta
npm install @hexascore/sdk@beta      # só quem pedir a tag recebe
```

## Publicar por automação (opcional)

Com o GitHub Actions desativado neste repositório, a publicação é manual. Quando for automatizar:

- **Trusted publishing (recomendado):** em npmjs.com → pacote → *Settings → Trusted Publisher*, aponte para o repositório `AlexGalhardo/hexascore-app` e o arquivo do workflow. O job precisa de `permissions: id-token: write` e roda `npm publish` sem token; o npm gera a atestação de proveniência sozinho.
- **Token:** *Access Tokens → Generate New Token → Granular Access Token*, só com permissão de escrita no pacote `@hexascore/sdk`, com validade. Guarde como secret `NPM_TOKEN` e use `NODE_AUTH_TOKEN` no job. Nunca commite o token nem o `.npmrc` com ele.

## Erros comuns

| Erro | Causa | Solução |
| --- | --- | --- |
| `404 Not Found - Scope not found` | organização `hexascore` não existe ou você não é membro | crie a organização (passo 1.3) |
| `402 Payment Required` | pacote com escopo publicado como privado | `npm publish --access public` |
| `403 Forbidden - You cannot publish over the previously published versions` | a versão já existe | suba a versão; versão publicada nunca é sobrescrita |
| `EOTP` / `one-time password` | faltou o código do 2FA | `npm publish --otp=<código>` |
| `ENEEDAUTH` | sem login | `npm login` |
| Quem instala recebe `ERR_MODULE_NOT_FOUND` | `dist/` antigo ou ausente | `bun run build` e confira `npm pack --dry-run` |

## Corrigir uma publicação errada

- **Despublicar:** só nas primeiras 72 horas e se nenhum outro pacote depender dela: `npm unpublish @hexascore/sdk@0.1.1`. O número da versão fica queimado para sempre.
- **Depois disso:** publique uma versão corrigida e avise na antiga: `npm deprecate @hexascore/sdk@0.1.1 "use 0.1.2"`.

## Checklist

- [ ] `bun run openapi:check` e `bun run sdk:generate` sem diferenças não commitadas
- [ ] versão igual em `package.json`, `jsr.json` e `pyproject.toml`
- [ ] `bun run typecheck`, `bun test`, `bun run build` passando
- [ ] `npm pack --dry-run` lista só o esperado
- [ ] `npm publish` e `npm view @hexascore/sdk version`
- [ ] tag `sdk-vX.Y.Z`, `CHANGELOG.md`, `PLAN.md`
