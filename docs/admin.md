# Admin (`/admin`)

Painel de moderação e operação do HexaScore. Acesso só para contas com `role = admin` (e não banidas): qualquer outra pessoa recebe **404** em `/admin` e em todas as subpáginas e actions (`app/[locale]/admin/guard.ts`, `isAdmin()`).

## Seções

| Rota | O que faz |
| --- | --- |
| `/admin` | Visão geral e atalhos |
| `/admin/contact` | Mensagens do formulário de contato |
| `/admin/reports` | Denúncias de usuários e listas: resolver/descartar, ocultar listas |
| `/admin/items` | Busca de itens, edição, links de afiliado (Steam/GOG, host em allowlist) e merge de itens duplicados (mesmo tipo, transação com lock) |
| `/admin/ingestion` | Status da ingestão (último fetch/normalização por fonte) |
| (perfis) | Banir usuário (`auth.api.banUser`, revoga sessões) |

Prefixo de idioma vale como no resto do site: `/pt-br/admin`, `/es/admin`.

## Contas admin

### Dono (produção e sandbox)

`aleexgvieira@gmail.com` (nickname `lekyz`) tem `role = admin` em **produção** (`hexascore.app`) e no **sandbox** (`dev.hexascore.app`). Login com a senha da própria conta. Depois de ganhar a role, saia e entre de novo para a sessão refletir a permissão.

### Admin padrão do seed (só local e sandbox)

| Campo | Valor |
| --- | --- |
| E-mail | `admin@hexascore.test` |
| Senha | `hexascoreBR@123` (sobrescrevível por `SEED_ADMIN_PASSWORD`; e-mail por `SEED_ADMIN_EMAIL`) |
| Nickname | `admin` |
| Id | `seed-admin` |

- Criado por **toda** execução do seed (`bun run db:seed`, `--existing-items`), exceto `--purge`; `bun lib/db/seed.ts --admin-only` cria só ele. Re-executar redefine a senha para a configurada (`lib/db/seed-admin.ts`).
- **Nunca existe em produção**: o seed recusa `VERCEL_ENV=production` e banco remoto sem `SEED_TARGET=sandbox` (`lib/db/seed-guard.ts`).
- Domínio `.test` (RFC 2606) não recebe e-mail: ninguém consegue pedir "esqueci a senha" para essa conta. A senha está no repositório de propósito (conta de teste); por isso a conta só pode viver em local/sandbox.
- `--purge` remove só usuários `seed-user-*`; o admin padrão permanece.

## Conceder ou revogar admin

- Local (PGlite/Docker): `bun run admin:grant <email>` em `apps/web` (pare o `bun run dev` antes com PGlite).
- Produção/sandbox: atualizar `user.role` para `admin` (ou `user`) no banco certo. Sempre confira o host do `DATABASE_URL` antes (produção `ep-blue-hall-…`, sandbox `ep-falling-sky-…`); nunca rode o seed contra produção.
