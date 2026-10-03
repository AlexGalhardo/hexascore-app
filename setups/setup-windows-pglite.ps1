# Local setup without Docker: empty DATABASE_URL -> in-process PGlite at apps/web/.pglite. Idempotent.
# Run: powershell -ExecutionPolicy Bypass -File setups\setup-windows-pglite.ps1
. (Join-Path $PSScriptRoot "_common.ps1")

Test-Bun
Install-Deps
Initialize-EnvFile "pglite"
Invoke-MigrateAndSeed
Write-NextSteps
Write-Host "  Database:       PGlite at apps/web/.pglite (delete the folder to reset)"
Write-Host "  Redis:          disabled (optional integration; set UPSTASH_REDIS_REST_* to enable)"
