# Self-hosting on an Ubuntu 24.04 VPS

This guide runs HexaScore on one Ubuntu 24.04 LTS server with Docker Compose, as an alternative to Railway (the production setup in [SDD 11](../sdd/11-infra-deploy.md)). Four containers run behind Caddy: the app image built from the repository's `Dockerfile`, PostgreSQL 18.6, RustFS for S3-compatible object storage, and Caddy for TLS.

[setup-vps-script.sh](setup-vps-script.sh) automates every step below on a fresh server. The guide explains what it does, so you can run it by hand or adapt it.

## Sizing

The app process is capped by `GOMEMLIMIT=900MiB` for Go and `NODE_OPTIONS=--max-old-space-size=384` for the SSR child (`.env.example`). Add Postgres, RustFS and Caddy, plus the image build (bun and Go compile in the same machine), and a reasonable starting point is:

- 2 vCPU, 4 GB RAM (the build is the peak; with 2 GB add swap or build elsewhere);
- 40 GB SSD or more, depending on the catalogue, covers and the 30 daily plus 12 monthly backups kept in the bucket;
- a public IPv4 (and IPv6 if you have it) with DNS `A`/`AAAA` records for your domain pointing at it before Caddy starts.

This is an estimate from the configured memory caps, not a measured benchmark.

## Server basics

### Non-root user and SSH

Log in as root once, create your own sudo user with your SSH key, and use it from then on:

```sh
adduser deploy
usermod -aG sudo deploy
install -d -m 700 -o deploy -g deploy /home/deploy/.ssh
cp /root/.ssh/authorized_keys /home/deploy/.ssh/authorized_keys
chown deploy:deploy /home/deploy/.ssh/authorized_keys
```

Check that `ssh deploy@<server>` works in a second terminal before you change the SSH settings in the [security checklist](#security-checklist).

The application itself runs in containers; the script also creates a system user `hexascore` without a login shell that owns the source checkout in `/opt/hexascore/src`.

### Firewall

```sh
sudo ufw allow OpenSSH
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

Docker publishes container ports through its own iptables rules, which bypass ufw. That is why only Caddy publishes ports (80 and 443); Postgres and the app are reachable only on the Compose network, and the RustFS console binds to `127.0.0.1`.

### Docker Engine

Install Docker Engine from Docker's apt repository as described on <https://docs.docker.com/engine/install/ubuntu/>, with exact versions. The script pins:

| Package | Version |
| --- | --- |
| `docker-ce`, `docker-ce-cli` | `5:29.9.2-1~ubuntu.24.04~noble` |
| `containerd.io` | `2.4.2-3~ubuntu.24.04~noble` |
| `docker-buildx-plugin` | `0.38.1-1~ubuntu.24.04~noble` |
| `docker-compose-plugin` | `5.6.1-1~ubuntu.24.04~noble` |

```sh
sudo apt-get install -y docker-ce=5:29.9.2-1~ubuntu.24.04~noble docker-ce-cli=5:29.9.2-1~ubuntu.24.04~noble \
  containerd.io=2.4.2-3~ubuntu.24.04~noble docker-buildx-plugin=0.38.1-1~ubuntu.24.04~noble \
  docker-compose-plugin=5.6.1-1~ubuntu.24.04~noble
sudo apt-mark hold docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
```

The hold keeps upgrades from moving Docker; bump the versions on purpose.

## The application image

Two options:

- **Build from source** (what the script does): check out a release tag and build the repository's `Dockerfile`. It is a multi-stage build (bun 1.4.4 for the Vite and SSR assets, Go 1.27.3 for a static binary, `node:26.13.0-alpine3.24` runtime with `pg_dump` 18.6 and `cwebp`), runs as a non-root user and starts `/app/app`. Pass the tag as `APP_VERSION` so responses carry it in `X-Hexascore-Version`.

  ```sh
  git clone https://github.com/AlexGalhardo/hexascore-app.git /opt/hexascore/src
  git -C /opt/hexascore/src checkout v0.39.0
  docker build --build-arg APP_VERSION=v0.39.0 -t hexascore-app:v0.39.0 /opt/hexascore/src
  ```

- **Pull from GHCR**: `.github/workflows/release.yml` pushes the image of every `vX.Y.Z` tag to `ghcr.io/alexgalhardo/hexascore-app` with the tags `X.Y.Z` and `X.Y`. Use the exact `X.Y.Z` tag, for example `docker pull ghcr.io/alexgalhardo/hexascore-app:0.39.0`. Whether the package is public is a GitHub setting; if the pull asks for credentials, build from source instead.

## Database and object storage

PostgreSQL runs as the same image as development, `postgres:18.6-alpine3.24`, but without the `fsync=off` flags of the local `docker-compose.yml`, with a named volume mounted at `/var/lib/postgresql` and a generated password. A native PostgreSQL 18.6 from apt.postgresql.org works too (see [setup-local-without-docker-unix.md](setup-local-without-docker-unix.md)); point `DATABASE_URL` at it.

Object storage is `rustfs/rustfs:1.0.4` (the local Compose version) with generated keys, or any external S3-compatible bucket: set `S3_ENDPOINT`, `S3_REGION`, `S3_BUCKET`, `S3_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY` and `S3_USE_PATH_STYLE` from the provider, and drop the `rustfs` service. An external bucket is the better choice for backups, because backups on the same disk as the database do not survive the loss of the server.

With RustFS, create the bucket once in its console. The console listens on `127.0.0.1:9001` of the server; reach it through an SSH tunnel and sign in with `S3_ACCESS_KEY_ID` and `S3_SECRET_ACCESS_KEY` from the env file:

```sh
ssh -L 9001:127.0.0.1:9001 deploy@<server>
```

Then open <http://127.0.0.1:9001> and create the bucket named in `S3_BUCKET` (`hexascore` in the script).

## Environment variables

[.env.example](../../.env.example) lists and explains every variable. Production keeps them in `/opt/hexascore/.env`, owned by root with mode 600, passed to the app container with `env_file`. The script writes this file once and never overwrites it.

Required in production:

| Variable | Value | Notes |
| --- | --- | --- |
| `APP_ENV` | `production` | the image default; turns on CanonicalHost, HTTPS upgrade and the production boot checks |
| `APP_KEY` | `base64:` + 32 random bytes | `echo "base64:$(openssl rand -base64 32)"`; keep an offline copy, it encrypts sessions and secrets |
| `APP_URL` | `https://your.domain` | every other host is redirected here (except `/up`) |
| `DATABASE_URL` | `postgres://hexascore:<password>@postgres:5432/hexascore?sslmode=disable` | the password from `openssl rand -hex 24` |
| `MAIL_MAILER`, `RESEND_API_KEY` | `resend` and the Resend key | `log` only writes mail to the log; also set `MAIL_FROM` and `MAIL_SUPPORT_ADDRESS` to addresses of your domain |
| `STRIPE_KEY`, `STRIPE_WEBHOOK_SECRET`, `STRIPE_PRICE_PRO`, `STRIPE_PRICE_BUSINESS` | from your Stripe account | billing refuses to boot in production without them; the webhook endpoint is `https://your.domain/stripe/webhook` |
| `TRUSTED_PROXIES`, `CLIENT_IP_HEADER`, `TRUST_FORWARDED_PROTO` | Caddy's network, `X-Real-IP`, `true` | see [Reverse proxy](#reverse-proxy-with-caddy) |
| `RATE_LIMIT_STORE`, `CACHE_STORE`, `SESSION_DRIVER` | `cache`, `database`, `database` | the app refuses to boot in production without the cache rate-limit store |
| `S3_*` | your bucket | covers, ID exports and backups |
| `BACKUP_KEY` | `openssl rand -base64 32` | separate from `APP_KEY`; without an offline copy the dumps cannot be decrypted |

Recommended: `ADMIN_EMAILS` (receives the monthly off-site backup link), `LOG_FORMAT=json`, `DB_MAX_OPEN=20`. Optional and off while empty: Google sign-in, Turnstile, Telegram alerts, analytics, IndexNow. `API_HOST` defaults to `api.hexascore.app`; set it to your own API host name if you serve the paid API on a separate host.

Never commit `.env`, and never paste its values in chat or tickets.

## Pre-deploy and health check

Before a new version starts, run the pre-deploy entry point once against the database. It drops `INVALID` index leftovers of a failed concurrent build, then runs `migrate`, and exits non-zero on any error ([migrations runbook](../runbooks/migrations.md)):

```sh
docker compose run --rm --no-deps app hexascore:predeploy
```

The image's entrypoint is `/app/app`, so the argument is the console command (on Railway the same command is `/app/app hexascore:predeploy`). Migrations are forward-only in production.

`GET /up` is the health check. CanonicalHost lets it through for any host, so the container healthcheck uses `wget -q --spider http://127.0.0.1:3000/up`, and `docker compose up -d --wait` waits for it.

## Reverse proxy with Caddy

Caddy (`caddy:2.11.7-alpine`) obtains and renews Let's Encrypt certificates automatically for the domain in its site address, redirects HTTP to HTTPS and proxies to the app on the Compose network:

```text
{
	email {$ACME_EMAIL}
}

{$HEXASCORE_DOMAIN} {
	encode zstd gzip
	reverse_proxy app:3000 {
		header_up X-Real-IP {remote_host}
	}
}
```

Caddy also sets `X-Forwarded-For` and `X-Forwarded-Proto`. The app trusts forwarding headers only from `TRUSTED_PROXIES`, so set it to the range Caddy connects from. The script gives the Compose network a fixed subnet, `172.31.250.0/24`, and sets:

```text
TRUSTED_PROXIES=172.31.250.0/24
CLIENT_IP_HEADER=X-Real-IP
TRUST_FORWARDED_PROTO=true
```

This mirrors Railway, where the proxy sends `X-Real-IP` (SDD 11 §6). Without it, rate limits and the admin IP allowlist would see Caddy's address for every visitor. Do not publish the app's port 3000 on the host: anything that can reach it directly could forge the header if it came from the trusted range.

## Backups

The app schedules its own backups (SDD 11 §7.3) when the scheduler runs inside `serve` (`SCHEDULE_IN_SERVE=true`):

- `backup:run --kind daily` at 06:00 (`America/Sao_Paulo`): an encrypted `pg_dump` uploaded to `backups/postgres/daily/` in `S3_BUCKET`, verified and pruned to the newest 30 daily and 12 monthly dumps; on day 1 it also copies the monthly dump.
- `backup:offsite` on day 1 at 07:00 mails `ADMIN_EMAILS` a signed 7-day link to the monthly dump; download it to storage outside the server.

Run one by hand to check the setup:

```sh
docker compose exec app /app/app backup:run --kind daily
```

Restore with `/app/app backup:restore <key> --target-url <url>` as described in the [restore runbook](../runbooks/restore.md). Keep `BACKUP_KEY` and `APP_KEY` in a password manager: they are not in the dumps.

## Updating

With the script, run it again with the new tag; it keeps the env file, fetches the tag, rebuilds, runs the pre-deploy and restarts:

```sh
sudo bash /opt/hexascore/src/docs/setups/setup-vps-script.sh --ref v0.40.0
```

By hand, from `/opt/hexascore`:

```sh
sudo -u hexascore git -C src fetch --tags origin
sudo -u hexascore git -C src checkout --detach v0.40.0
sudo APP_VERSION=v0.40.0 IMAGE_TAG=v0.40.0 docker compose build app
sudo APP_VERSION=v0.40.0 IMAGE_TAG=v0.40.0 docker compose run --rm --no-deps app hexascore:predeploy
sudo APP_VERSION=v0.40.0 IMAGE_TAG=v0.40.0 docker compose up -d --wait app
```

To roll back, start the previous image tag (it is still on the server). The database stays on the new schema, which the expand/contract policy keeps compatible with the previous release ([rollback runbook](../runbooks/rollback.md)).

## Logs

The app logs JSON to stdout (`LOG_FORMAT=json`); Docker keeps it with rotation (`max-size 10m`, `max-file 5` in the generated compose file):

```sh
cd /opt/hexascore
sudo docker compose ps
sudo docker compose logs -f --tail 200 app
sudo docker compose logs --tail 200 caddy postgres rustfs
```

## Security checklist

- [ ] SSH with keys only: in `/etc/ssh/sshd_config.d/`, set `PasswordAuthentication no` and `PermitRootLogin no`, then `sudo systemctl reload ssh` (after checking key login works).
- [ ] ufw active with only 22, 80 and 443 open; no container port published except Caddy's.
- [ ] `unattended-upgrades` installed and enabled (`sudo dpkg-reconfigure -plow unattended-upgrades`) for Ubuntu security updates.
- [ ] Secrets only in `/opt/hexascore/.env`, owned by root, mode 600; never in the compose file, the image, shell history or git.
- [ ] Never commit `.env`; the repository's `.dockerignore` keeps it out of the image build context.
- [ ] `APP_KEY` and `BACKUP_KEY` backed up offline.
- [ ] Backups go to a bucket off the server, or the monthly off-site link is downloaded every month.
- [ ] Image tags and apt versions pinned; update them on purpose and read the release notes first.
- [ ] Production Stripe and Resend keys are restricted keys with only the permissions the app needs.

## What the script does

[setup-vps-script.sh](setup-vps-script.sh) needs root on Ubuntu 24.04 and a `--ref`; it prompts for the domain and email when they are not passed:

```sh
sudo bash setup-vps-script.sh --ref v0.39.0 --domain hexascore.example.com --email ops@example.com
```

1. Installs `ca-certificates`, `curl`, `git`, `openssl`, `ufw`, `unattended-upgrades`; opens 22, 80, 443 and enables ufw.
2. Installs Docker Engine from Docker's apt repository at the versions above and holds them.
3. Creates the system user `hexascore`, clones the repository into `/opt/hexascore/src` and checks out the tag or branch.
4. On the first run only, writes `/opt/hexascore/.env` (root, mode 600) with `APP_KEY`, `BACKUP_KEY`, the database password and the RustFS keys from `openssl rand`; nothing secret is printed. Domain and email are saved there too, so changing them later means editing that file.
5. Writes `/opt/hexascore/compose.yml` (app, postgres, rustfs, caddy) and `/opt/hexascore/Caddyfile`.
6. Starts Postgres and RustFS. If the Stripe values or `RESEND_API_KEY` (with `MAIL_MAILER=resend`) are empty, it lists them and stops; fill them in and run it again.
7. Builds the image, runs `hexascore:predeploy`, starts the app and Caddy, and waits for `/up`.

The script does not change SSH settings or create the RustFS bucket; do those by hand as described above.
