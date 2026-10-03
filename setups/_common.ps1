# Shared helpers for setups/setup-windows-*.ps1. Dot-sourced, not executed.
# ASCII only: Windows PowerShell 5.1 reads BOM-less scripts as ANSI.

$ErrorActionPreference = "Stop"

$ExpectedBun = "1.4.2"
$LocalDatabaseUrl = "postgres://hexascore:hexascore@localhost:5432/hexascore"
$LocalUpstashUrl = "http://localhost:8079"
$LocalUpstashToken = "local-dev-token"

$RootDir = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$EnvFile = Join-Path $RootDir "apps\web\.env.local"
$EnvExample = Join-Path $RootDir ".env.example"

function Write-Step([string]$Message) { Write-Host ""; Write-Host "==> $Message" -ForegroundColor Cyan }
function Write-Ok([string]$Message) { Write-Host "  ok " -ForegroundColor Green -NoNewline; Write-Host $Message }
function Write-Warn([string]$Message) { Write-Host "  warn " -ForegroundColor Yellow -NoNewline; Write-Host $Message }
function Stop-Setup([string]$Message) {
	Write-Host "  error " -ForegroundColor Red -NoNewline
	Write-Host $Message
	exit 1
}

# Runs a native command from the repo root and fails the setup on a non-zero exit code.
function Invoke-Native([string]$FailMessage, [scriptblock]$Command) {
	Push-Location $RootDir
	try {
		& $Command
		if ($LASTEXITCODE -ne 0) { Stop-Setup $FailMessage }
	} finally {
		Pop-Location
	}
}

function Test-Bun {
	Write-Step "Checking Bun"
	if (-not (Get-Command bun -ErrorAction SilentlyContinue)) {
		Stop-Setup "Bun not found. Install ${ExpectedBun}: powershell -c `"& {`$(irm bun.sh/install.ps1)} -Version $ExpectedBun`""
	}
	$version = (bun --version).Trim()
	if ($version -eq $ExpectedBun) { Write-Ok "bun $version" }
	else { Write-Warn "bun $version found, project pins $ExpectedBun (packageManager). Run: bun upgrade --version $ExpectedBun" }
}

function Install-Deps {
	Write-Step "Installing dependencies (bun install)"
	Invoke-Native "bun install failed (see output above)" { bun install --frozen-lockfile }
	Write-Ok "dependencies installed"
}

function New-Secret {
	$bytes = New-Object byte[] 32
	[System.Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($bytes)
	return -join ($bytes | ForEach-Object { $_.ToString("x2") })
}

function Read-EnvLines { return [System.IO.File]::ReadAllText($EnvFile) -split "`r?`n" }

function Get-EnvValue([string]$Key) {
	foreach ($line in Read-EnvLines) {
		if ($line.StartsWith("$Key=")) { return $line.Substring($Key.Length + 1).Trim('"') }
	}
	return ""
}

# Sets KEY="value" (replacing or appending). Writes UTF-8 without BOM + LF to a temp file, then swaps it in,
# so a failure never leaves a half-written .env.local.
function Set-EnvValue([string]$Key, [string]$Value) {
	$lines = [System.Collections.Generic.List[string]]::new()
	$found = $false
	foreach ($line in Read-EnvLines) {
		if ($line.StartsWith("$Key=")) { $lines.Add("$Key=`"$Value`""); $found = $true }
		else { $lines.Add($line) }
	}
	if (-not $found) {
		if ($lines.Count -gt 0 -and $lines[$lines.Count - 1] -eq "") { $lines.Insert($lines.Count - 1, "$Key=`"$Value`"") }
		else { $lines.Add("$Key=`"$Value`"") }
	}
	$tmp = "$EnvFile.tmp"
	[System.IO.File]::WriteAllText($tmp, ($lines -join "`n"), (New-Object System.Text.UTF8Encoding($false)))
	Move-Item -Force $tmp $EnvFile
}

# Mode "docker" (local Postgres + Upstash proxy) or "pglite" (no Docker).
function Initialize-EnvFile([string]$Mode) {
	Write-Step "Preparing apps/web/.env.local"
	if (-not (Test-Path $EnvFile)) {
		if (-not (Test-Path $EnvExample)) { Stop-Setup ".env.example not found at $EnvExample" }
		Copy-Item $EnvExample $EnvFile
		Write-Ok "created from .env.example"
	} else {
		Write-Ok "already exists, only filling empty/local values"
	}

	foreach ($key in @("BETTER_AUTH_SECRET", "CRON_SECRET")) {
		if (-not (Get-EnvValue $key)) { Set-EnvValue $key (New-Secret); Write-Ok "generated $key" }
	}

	$dbUrl = Get-EnvValue "DATABASE_URL"
	$upstashUrl = Get-EnvValue "UPSTASH_REDIS_REST_URL"
	if ($Mode -eq "docker") {
		if (-not $dbUrl) { Set-EnvValue "DATABASE_URL" $LocalDatabaseUrl; Write-Ok "DATABASE_URL -> local Docker Postgres" }
		elseif ($dbUrl -ne $LocalDatabaseUrl) { Write-Warn "DATABASE_URL points elsewhere; keeping it (migrations will run there)" }
		if (-not $upstashUrl) {
			Set-EnvValue "UPSTASH_REDIS_REST_URL" $LocalUpstashUrl
			Set-EnvValue "UPSTASH_REDIS_REST_TOKEN" $LocalUpstashToken
			Write-Ok "UPSTASH_REDIS_REST_* -> local proxy"
		}
	} else {
		# Only undo values the Docker setup wrote; anything else was set on purpose by the developer.
		if ($dbUrl -eq $LocalDatabaseUrl) { Set-EnvValue "DATABASE_URL" ""; Write-Ok "DATABASE_URL cleared (was the local Docker URL) -> PGlite" }
		elseif ($dbUrl) { Write-Warn "DATABASE_URL is set, so PGlite will NOT be used. Clear it in apps/web/.env.local for PGlite." }
		if ($upstashUrl -eq $LocalUpstashUrl) {
			Set-EnvValue "UPSTASH_REDIS_REST_URL" ""
			Set-EnvValue "UPSTASH_REDIS_REST_TOKEN" ""
			Write-Ok "UPSTASH_REDIS_REST_* cleared (local proxy needs Docker)"
		}
	}
}

function Invoke-MigrateAndSeed {
	Write-Step "Applying database migrations"
	Invoke-Native "migrations failed (see output above)" { bun run db:migrate }
	Write-Ok "migrations applied"

	Write-Step "Seeding"
	if (Test-Path (Join-Path $RootDir "apps\web\lib\db\seed.ts")) {
		Invoke-Native "seed failed (see output above)" { bun run db:seed }
		Write-Ok "seed applied"
	} else {
		Write-Warn "no seed script yet (apps/web/lib/db/seed.ts), skipping"
	}
}

function Write-NextSteps {
	Write-Step "Done"
	Write-Host "  Start the app:  bun run dev"
	Write-Host "  Web:            http://localhost:3000"
	Write-Host "  Env file:       apps/web/.env.local"
}
