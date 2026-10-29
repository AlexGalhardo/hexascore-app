# Idempotent local setup: tools, .env, APP_KEY, local stack, JS deps, migrations.
$ErrorActionPreference = 'Stop'

Set-Location (Join-Path $PSScriptRoot '..')

function Invoke-Checked([string]$Command, [string[]]$Arguments) {
	& $Command @Arguments
	if ($LASTEXITCODE -ne 0) {
		throw "setup: '$Command $($Arguments -join ' ')' failed with exit code $LASTEXITCODE"
	}
}

foreach ($tool in 'go', 'bun', 'docker', 'vukem') {
	if (-not (Get-Command $tool -ErrorAction SilentlyContinue)) {
		throw "setup: missing tool: $tool"
	}
}

if (-not (Test-Path '.env')) {
	Copy-Item '.env.example' '.env'
}

# vukem key:generate refuses to run without a fresh key, so only generate when APP_KEY is empty.
if (-not (Select-String -Path '.env' -Pattern '^APP_KEY=.' -Quiet)) {
	Invoke-Checked 'vukem' @('key:generate')
}

Invoke-Checked 'docker' @('compose', 'up', '-d', '--wait')
Invoke-Checked 'bun' @('install', '--frozen-lockfile')
Invoke-Checked 'vukem' @('migrate')
