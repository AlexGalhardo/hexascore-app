# Idempotent local setup: tools, .env, APP_KEY, local stack, JS deps, migrations.
# -NoServices skips the Docker stack and migrations (CI Windows runners have
# no Linux containers); everything else still runs.
param([switch]$NoServices)
$ErrorActionPreference = 'Stop'

Set-Location (Join-Path $PSScriptRoot '..')

function Invoke-Checked([string]$Command, [string[]]$Arguments) {
	& $Command @Arguments
	if ($LASTEXITCODE -ne 0) {
		throw "setup: '$Command $($Arguments -join ' ')' failed with exit code $LASTEXITCODE"
	}
}

$tools = @('go', 'bun', 'vukem')
if (-not $NoServices) { $tools += 'docker' }
foreach ($tool in $tools) {
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

Invoke-Checked 'git' @('config', 'core.hooksPath', '.githooks')
# Pinned to the version in .github/workflows/ci.yml (GOLANGCI_LINT_VERSION). Install docs: https://golangci-lint.run/docs/welcome/install/
$golangciLintVersion = '2.15.0'

function Get-GolangciLintVersion {
	if (-not (Get-Command golangci-lint -ErrorAction SilentlyContinue)) {
		return $null
	}
	$out = & golangci-lint version
	if ($out -match 'version (\d+\.\d+\.\d+)') {
		return $Matches[1]
	}
	return $null
}

if ((Get-GolangciLintVersion) -ne $golangciLintVersion) {
	Invoke-Checked 'go' @('install', "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v$golangciLintVersion")
}
if ((Get-GolangciLintVersion) -ne $golangciLintVersion) {
	throw "setup: golangci-lint is not $golangciLintVersion; make sure $(go env GOPATH)\bin is on PATH"
}

if (-not $NoServices) {
	Invoke-Checked 'docker' @('compose', 'up', '-d', '--wait')
}
Invoke-Checked 'bun' @('install', '--frozen-lockfile')
if (-not $NoServices) {
	Invoke-Checked 'vukem' @('migrate')
}
