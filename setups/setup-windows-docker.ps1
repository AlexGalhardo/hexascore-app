# Local setup with Docker: Postgres 18 + Redis + Upstash REST proxy (infra/docker-compose.yaml). Idempotent.
# Run: powershell -ExecutionPolicy Bypass -File setups\setup-windows-docker.ps1
. (Join-Path $PSScriptRoot "_common.ps1")

$ComposeFile = Join-Path $RootDir "infra\docker-compose.yaml"

Test-Bun

Write-Step "Checking Docker"
if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
	Stop-Setup "Docker not found. Install Docker Desktop, or use setups\setup-windows-pglite.ps1"
}
# try/catch: on Windows PowerShell 5.1, redirected native stderr turns into a terminating error under "Stop".
try { $serverVersion = docker version --format "{{.Server.Version}}" 2>$null } catch { $serverVersion = $null }
if ($LASTEXITCODE -ne 0 -or -not $serverVersion) {
	Stop-Setup "Docker daemon is not running. Start Docker Desktop and run this script again."
}
try { docker compose version *> $null } catch { $null = $_ }
if ($LASTEXITCODE -ne 0) { Stop-Setup "Docker Compose v2 plugin not found ('docker compose')." }
Write-Ok "docker $serverVersion"

Install-Deps
Initialize-EnvFile "docker"

Write-Step "Starting containers (Postgres, Redis, Upstash REST proxy)"
Invoke-Native "containers did not become healthy. Inspect: docker compose -f infra/docker-compose.yaml logs (is port 5432/6379/8079 already in use?)" {
	docker compose -f $ComposeFile up -d --wait --wait-timeout 120
}
Write-Ok "postgres, redis and Upstash REST proxy healthy"

Invoke-MigrateAndSeed
Write-NextSteps
Write-Host "  Postgres:       $LocalDatabaseUrl"
Write-Host "  Upstash REST:   $LocalUpstashUrl (token $LocalUpstashToken)"
Write-Host "  Stop:           docker compose -f infra/docker-compose.yaml down   (add -v to drop data)"
