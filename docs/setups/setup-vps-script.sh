#!/usr/bin/env bash
# Self-host HexaScore on a fresh Ubuntu 24.04 LTS VPS (guide: docs/setups/setup-vps-ubuntu.md).
#
# Usage (as root, or with sudo):
#   sudo bash setup-vps-script.sh --ref v0.39.0 --domain hexascore.example.com --email ops@example.com
#
# Options (each one can also come from the environment):
#   --ref REF        git tag or branch to deploy             HEXASCORE_REF (required)
#   --domain HOST    public host name for Caddy and APP_URL  HEXASCORE_DOMAIN (prompted when missing)
#   --email ADDRESS  ACME account email for Caddy            HEXASCORE_EMAIL (prompted when missing)
#   --repo URL       git repository to clone                 HEXASCORE_REPO (default: the public GitHub repo)
#
# What it does, idempotently: apt packages, ufw (22, 80, 443), Docker Engine from Docker's apt repo
# at pinned versions, a system user `hexascore`, a checkout of REF in /opt/hexascore/src, a compose
# project (app, postgres, rustfs, caddy) in /opt/hexascore, generated secrets in /opt/hexascore/.env
# (mode 600, never printed), image build, `hexascore:predeploy` (migrations), start, and a wait on /up.
#
# Re-run it with a new --ref to update. An existing .env is never overwritten, so APP_KEY and the
# database password survive re-runs. Third-party secrets (Stripe, Resend, ...) are left empty: the
# script lists them and stops before starting the app until they are filled in.
set -euo pipefail

# Pinned versions. Docker packages: Docker's apt repo for noble; images: same tags as docker-compose.yml.
DOCKER_CE_VERSION="5:29.9.2-1~ubuntu.24.04~noble"
CONTAINERD_VERSION="2.4.2-3~ubuntu.24.04~noble"
BUILDX_VERSION="0.38.1-1~ubuntu.24.04~noble"
COMPOSE_VERSION="5.6.1-1~ubuntu.24.04~noble"
POSTGRES_IMAGE="postgres:18.6-alpine3.24"
RUSTFS_IMAGE="rustfs/rustfs:1.0.4"
CADDY_IMAGE="caddy:2.11.7-alpine"

BASE=/opt/hexascore
SRC="$BASE/src"
ENV_FILE="$BASE/.env"
APP_USER=hexascore
# Fixed subnet of the compose network, so TRUSTED_PROXIES can name Caddy's range.
NET_SUBNET="172.31.250.0/24"

REF="${HEXASCORE_REF:-}"
DOMAIN="${HEXASCORE_DOMAIN:-}"
EMAIL="${HEXASCORE_EMAIL:-}"
REPO="${HEXASCORE_REPO:-https://github.com/AlexGalhardo/hexascore-app.git}"

die() {
	echo "setup-vps: $*" >&2
	exit 1
}
log() { echo "==> $*"; }

while [ $# -gt 0 ]; do
	case "$1" in
	--ref) REF="${2:?--ref needs a value}"; shift 2 ;;
	--domain) DOMAIN="${2:?--domain needs a value}"; shift 2 ;;
	--email) EMAIL="${2:?--email needs a value}"; shift 2 ;;
	--repo) REPO="${2:?--repo needs a value}"; shift 2 ;;
	-h | --help) sed -n '2,20p' "$0"; exit 0 ;;
	*) die "unknown argument: $1 (see --help)" ;;
	esac
done

[ "$(id -u)" -eq 0 ] || die "run as root (sudo)"
# shellcheck source=/dev/null
. /etc/os-release
[ "${ID:-}" = ubuntu ] && [ "${VERSION_ID:-}" = "24.04" ] || die "this script targets Ubuntu 24.04 LTS"
[ -n "$REF" ] || die "--ref (or HEXASCORE_REF) is required, for example --ref v0.39.0"

# Reads one non-secret key from the env file; secrets are only tested for emptiness, never printed.
env_get() { [ -f "$ENV_FILE" ] && sed -n "s/^$1=//p" "$ENV_FILE" | tail -n 1 || true; }

DOMAIN="${DOMAIN:-$(env_get HEXASCORE_DOMAIN)}"
EMAIL="${EMAIL:-$(env_get ACME_EMAIL)}"
if [ -z "$DOMAIN" ] && [ -t 0 ]; then read -r -p "Public domain (DNS A/AAAA already pointing here): " DOMAIN; fi
if [ -z "$EMAIL" ] && [ -t 0 ]; then read -r -p "Email for Let's Encrypt (Caddy ACME account): " EMAIL; fi
[ -n "$DOMAIN" ] || die "--domain (or HEXASCORE_DOMAIN) is required"
[ -n "$EMAIL" ] || die "--email (or HEXASCORE_EMAIL) is required"
case "$DOMAIN" in *[!A-Za-z0-9.-]* | "") die "invalid domain: $DOMAIN" ;; esac

log "apt packages"
export DEBIAN_FRONTEND=noninteractive
apt-get update -q
apt-get install -y -q ca-certificates curl git openssl ufw unattended-upgrades

log "firewall (ufw: 22, 80, 443)"
ufw allow OpenSSH >/dev/null
ufw allow 80/tcp >/dev/null
ufw allow 443/tcp >/dev/null
ufw --force enable >/dev/null

log "Docker Engine from download.docker.com (pinned)"
install -m 0755 -d /etc/apt/keyrings
if [ ! -s /etc/apt/keyrings/docker.asc ]; then
	curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
	chmod a+r /etc/apt/keyrings/docker.asc
fi
cat >/etc/apt/sources.list.d/docker.sources <<EOF
Types: deb
URIs: https://download.docker.com/linux/ubuntu
Suites: ${VERSION_CODENAME}
Components: stable
Signed-By: /etc/apt/keyrings/docker.asc
EOF
apt-get update -q
apt-get install -y -q --allow-downgrades --allow-change-held-packages \
	"docker-ce=$DOCKER_CE_VERSION" "docker-ce-cli=$DOCKER_CE_VERSION" "containerd.io=$CONTAINERD_VERSION" \
	"docker-buildx-plugin=$BUILDX_VERSION" "docker-compose-plugin=$COMPOSE_VERSION"
# Keep unattended-upgrades and apt upgrade from moving Docker off the pinned versions.
apt-mark hold docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin >/dev/null
systemctl enable --now docker >/dev/null

log "user $APP_USER and $BASE"
id -u "$APP_USER" >/dev/null 2>&1 || useradd --system --create-home --shell /usr/sbin/nologin "$APP_USER"
install -d -m 0755 -o root -g root "$BASE"
install -d -m 0755 -o "$APP_USER" -g "$APP_USER" "$SRC"

log "checkout $REF"
as_app() { runuser -u "$APP_USER" -- "$@"; }
if [ ! -d "$SRC/.git" ]; then
	as_app git clone --quiet "$REPO" "$SRC"
fi
as_app git -C "$SRC" fetch --quiet --tags --force origin "$REF"
as_app git -C "$SRC" checkout --quiet --detach FETCH_HEAD
APP_VERSION="$REF"
# Docker tags allow [A-Za-z0-9_.-]; branch names may carry slashes.
IMAGE_TAG="$(printf '%s' "$REF" | tr -c 'A-Za-z0-9_.-' '-')"

log "environment file $ENV_FILE"
if [ ! -f "$ENV_FILE" ]; then
	umask 077
	db_password="$(openssl rand -hex 24)"
	{
		echo "# Generated by setup-vps-script.sh. Secrets live only here: root-owned, mode 600, never in git."
		echo "HEXASCORE_DOMAIN=$DOMAIN"
		echo "ACME_EMAIL=$EMAIL"
		echo "APP_NAME=HexaScore"
		echo "APP_ENV=production"
		echo "APP_DEBUG=false"
		echo "APP_KEY=base64:$(openssl rand -base64 32)"
		echo "APP_URL=https://$DOMAIN"
		echo "API_HOST=api.$DOMAIN"
		echo "LOG_FORMAT=json"
		echo "POSTGRES_PASSWORD=$db_password"
		echo "DATABASE_URL=postgres://hexascore:$db_password@postgres:5432/hexascore?sslmode=disable"
		echo "DB_MAX_OPEN=20"
		echo "CACHE_STORE=database"
		echo "RATE_LIMIT_STORE=cache"
		echo "SESSION_DRIVER=database"
		echo "SESSION_RECORD_CLIENT=true"
		echo "S3_BUCKET=hexascore"
		echo "S3_REGION=us-east-1"
		echo "S3_ENDPOINT=http://rustfs:9000"
		echo "S3_ACCESS_KEY_ID=hexascore-$(openssl rand -hex 8)"
		echo "S3_SECRET_ACCESS_KEY=$(openssl rand -hex 32)"
		echo "S3_USE_PATH_STYLE=true"
		echo "S3_INSECURE_HTTP=true"
		echo "BACKUP_KEY=$(openssl rand -base64 32)"
		echo "TRUSTED_PROXIES=$NET_SUBNET"
		echo "CLIENT_IP_HEADER=X-Real-IP"
		echo "TRUST_FORWARDED_PROTO=true"
		echo "SCHEDULE_TIMEZONE=America/Sao_Paulo"
		echo "SCHEDULE_IN_SERVE=true"
		echo "QUEUE_NAMES=default,ingest"
		echo "QUEUE_CONCURRENCY=3"
		echo "SHUTDOWN_TIMEOUT=30s"
		echo "GOMEMLIMIT=900MiB"
		echo "NODE_OPTIONS=--max-old-space-size=384"
		echo "MAIL_MAILER=resend"
		echo "MAIL_FROM=HexaScore <noreply@$DOMAIN>"
		echo "MAIL_SUPPORT_ADDRESS=support@$DOMAIN"
		echo "ADMIN_EMAILS="
		echo "# Third-party secrets: fill these in, then run the script again."
		echo "RESEND_API_KEY="
		echo "STRIPE_KEY="
		echo "STRIPE_WEBHOOK_SECRET="
		echo "STRIPE_PRICE_PRO="
		echo "STRIPE_PRICE_BUSINESS="
		echo "# Optional, off while empty (see .env.example):"
		echo "SOCIAL_GOOGLE_CLIENT_ID="
		echo "SOCIAL_GOOGLE_CLIENT_SECRET="
		echo "TURNSTILE_SITE_KEY="
		echo "TURNSTILE_SECRET_KEY="
		echo "TURNSTILE_HOSTNAME="
		echo "TELEGRAM_BOT_TOKEN="
		echo "TELEGRAM_CHAT_ID="
	} >"$ENV_FILE"
	unset db_password
	umask 022
	log "generated APP_KEY, BACKUP_KEY, the database password and the RustFS keys (not shown)"
else
	log "keeping the existing $ENV_FILE"
fi
chown root:root "$ENV_FILE"
chmod 600 "$ENV_FILE"

log "compose project in $BASE"
cat >"$BASE/Caddyfile" <<'EOF'
{
	email {$ACME_EMAIL}
}

{$HEXASCORE_DOMAIN} {
	encode zstd gzip
	reverse_proxy app:3000 {
		# The app reads the client IP from this header, trusted only from TRUSTED_PROXIES.
		header_up X-Real-IP {remote_host}
	}
}
EOF

cat >"$BASE/compose.yml" <<EOF
# Generated by setup-vps-script.sh; edit the script, not this file.
name: hexascore

x-logging: &logging
  driver: json-file
  options: { max-size: "10m", max-file: "5" }

services:
  postgres:
    image: $POSTGRES_IMAGE
    restart: unless-stopped
    environment:
      POSTGRES_USER: hexascore
      POSTGRES_PASSWORD: \${POSTGRES_PASSWORD:?missing in .env}
      POSTGRES_DB: hexascore
    volumes: ["postgres:/var/lib/postgresql"]
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U hexascore"]
      interval: 5s
      retries: 30
    logging: *logging

  rustfs:
    image: $RUSTFS_IMAGE
    restart: unless-stopped
    command: ["/data"]
    environment:
      RUSTFS_ACCESS_KEY: \${S3_ACCESS_KEY_ID:?missing in .env}
      RUSTFS_SECRET_KEY: \${S3_SECRET_ACCESS_KEY:?missing in .env}
    # Console only on loopback: reach it through an SSH tunnel.
    ports: ["127.0.0.1:9001:9001"]
    volumes: ["rustfs:/data"]
    healthcheck:
      test: ["CMD-SHELL", "wget -q --spider http://127.0.0.1:9000/health"]
      interval: 5s
      timeout: 3s
      retries: 12
    logging: *logging

  app:
    build:
      context: ./src
      args: { APP_VERSION: "\${APP_VERSION:-}" }
    image: hexascore-app:\${IMAGE_TAG:-local}
    restart: unless-stopped
    env_file: [.env]
    depends_on:
      postgres: { condition: service_healthy }
      rustfs: { condition: service_healthy }
    volumes: ["app-storage:/app/storage"]
    healthcheck:
      test: ["CMD-SHELL", "wget -q --spider http://127.0.0.1:3000/up"]
      interval: 5s
      retries: 30
    logging: *logging

  caddy:
    image: $CADDY_IMAGE
    restart: unless-stopped
    depends_on:
      app: { condition: service_healthy }
    environment:
      HEXASCORE_DOMAIN: \${HEXASCORE_DOMAIN:?missing in .env}
      ACME_EMAIL: \${ACME_EMAIL:?missing in .env}
    ports: ["80:80", "443:443"]
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy-data:/data
      - caddy-config:/config
    logging: *logging

networks:
  default:
    ipam:
      config: [{ subnet: "$NET_SUBNET" }]

volumes:
  postgres:
  rustfs:
  app-storage:
  caddy-data:
  caddy-config:
EOF

compose() { APP_VERSION="$APP_VERSION" IMAGE_TAG="$IMAGE_TAG" docker compose --project-directory "$BASE" -f "$BASE/compose.yml" "$@"; }

log "postgres and rustfs"
compose up -d --wait postgres rustfs

missing=()
for key in STRIPE_KEY STRIPE_WEBHOOK_SECRET STRIPE_PRICE_PRO STRIPE_PRICE_BUSINESS; do
	[ -n "$(env_get "$key")" ] || missing+=("$key")
done
if [ "$(env_get MAIL_MAILER)" = resend ] && [ -z "$(env_get RESEND_API_KEY)" ]; then
	missing+=("RESEND_API_KEY")
fi
if [ "${#missing[@]}" -gt 0 ]; then
	echo
	echo "The app refuses to boot in production without these values. Fill them in $ENV_FILE"
	echo "(sudo nano $ENV_FILE), then run this script again:"
	printf '  %s\n' "${missing[@]}"
	echo "Optional and off while empty: ADMIN_EMAILS (backup:offsite), SOCIAL_GOOGLE_*, TURNSTILE_*, TELEGRAM_*."
	exit 0
fi

log "build hexascore-app:$IMAGE_TAG"
compose build app

log "pre-deploy: hexascore:predeploy (migrations)"
compose run --rm --no-deps app hexascore:predeploy

log "start app and caddy"
compose up -d --wait app caddy

log "wait for /up"
for _ in $(seq 1 30); do
	if compose exec -T app wget -q --spider http://127.0.0.1:3000/up; then
		log "HexaScore $REF is up at https://$DOMAIN"
		echo "Next: create the bucket 'hexascore' in the RustFS console (ssh -L 9001:127.0.0.1:9001 <user>@<vps>,"
		echo "then http://127.0.0.1:9001) if you have not yet, and read docs/setups/setup-vps-ubuntu.md for backups."
		exit 0
	fi
	sleep 2
done
die "/up did not answer; check: docker compose --project-directory $BASE logs app"
