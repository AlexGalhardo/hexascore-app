# Scripts de setup local

Sobem o app localmente com um comando. Todos são idempotentes: rode de novo quando quiser (depois de um pull, de uma migration nova, de um `.env.local` quebrado).

| Script | SO | Banco | Redis (API Upstash) | Requer |
| --- | --- | --- | --- | --- |
| `setup-unix-docker.sh` | Linux / macOS / Git Bash | Postgres 18 no Docker (`localhost:5432`) | proxy local `http://localhost:8079` | Bun 1.4.2, Docker |
| `setup-windows-docker.ps1` | Windows | Postgres 18 no Docker (`localhost:5432`) | proxy local `http://localhost:8079` | Bun 1.4.2, Docker Desktop |
| `setup-unix-pglite.sh` | Linux / macOS / Git Bash | PGlite em processo (`apps/web/.pglite`) | desligado | Bun 1.4.2 |
| `setup-windows-pglite.ps1` | Windows | PGlite em processo (`apps/web/.pglite`) | desligado | Bun 1.4.2 |

```sh
./setups/setup-unix-docker.sh
```

```powershell
powershell -ExecutionPolicy Bypass -File setups\setup-windows-docker.ps1
```

## O que fazem

1. Checam o Bun (avisam se não for o `1.4.2` fixado) e, nas variantes Docker, se o daemon e o `docker compose` estão disponíveis.
2. `bun install --frozen-lockfile`.
3. Criam `apps/web/.env.local` a partir do `.env.example` se não existir e preenchem só valores vazios: `BETTER_AUTH_SECRET` e `CRON_SECRET` aleatórios (64 hex); as variantes Docker também apontam `DATABASE_URL` e `UPSTASH_REDIS_REST_URL`/`_TOKEN` para os containers locais. Valores que você definiu nunca são sobrescritos; as variantes PGlite só limpam os valores locais escritos pelas variantes Docker. O arquivo é regravado via arquivo temporário, então uma falha nunca o deixa pela metade.
4. Variantes Docker: `docker compose -f infra/docker-compose.yaml up -d --wait` (Postgres, Redis e o proxy REST Upstash, todos com healthcheck).
5. `bun run db:migrate` e depois `bun run db:seed` se `apps/web/lib/db/seed.ts` existir (senão pula).
6. Mostram os próximos passos: `bun run dev` → http://localhost:3000.

`_common.sh` / `_common.ps1` são os helpers compartilhados; não rode diretamente.

## Serviços locais (variantes Docker)

| Serviço | Endereço | Credenciais |
| --- | --- | --- |
| Postgres 18 | `postgres://hexascore:hexascore@localhost:5432/hexascore` | user/senha/db `hexascore` |
| Redis 8 | `redis://localhost:6379` | nenhuma |
| Proxy REST Upstash (`serverless-redis-http`) | `http://localhost:8079` | token `local-dev-token` |

Parar: `docker compose -f infra/docker-compose.yaml down` (com `-v` apaga também os volumes de dados). App em container: `docker compose -f infra/docker-compose.yaml --profile app up -d --build` (usa `infra/Dockerfile` e aplica as migrations ao subir).

## Problemas comuns

- **Porta em uso** (5432, 6379, 8079): pare o outro Postgres/Redis ou troque a porta do host em `infra/docker-compose.yaml` e no `DATABASE_URL`.
- **`[env] Invalid environment variables`**: a mensagem lista cada chave inválida; corrija em `apps/web/.env.local` (veja `.env.example`).
- **Resetar o PGlite**: apague `apps/web/.pglite` e rode o script de novo.
