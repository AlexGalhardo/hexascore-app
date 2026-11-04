param(
    [Parameter(Mandatory = $true)]
    [ValidateSet('Export', 'Restore', 'Sync')]
    [string] $Action
)

$ErrorActionPreference = 'Stop'
$projectId = '51b50f39-b99d-4720-b177-e5b7f4f530d3'
$root = Split-Path -Parent $PSScriptRoot
$seedDir = Join-Path $root '.local-seeds'
$schemaFile = Join-Path $seedDir 'production-schema.dump'
$dataFile = Join-Path $seedDir 'production-catalog.dump'
$pruneFile = Join-Path $root 'scripts/production-catalog-prune.sql'
$previewDb = 'hexascore_prod_preview'
$tables = @(
    'data_sources', 'items', 'item_translations', 'item_alt_titles',
    'entities', 'entity_translations', 'item_entities', 'entity_links',
    'item_links', 'entity_stats', 'editions', 'stores', 'offers',
    'item_genre_sources', 'item_activity_daily', 'item_popularity_external',
    'trending_snapshots'
)

function Invoke-Checked([string] $exe, [string[]] $arguments) {
    & $exe @arguments
    if ($LASTEXITCODE -ne 0) { throw "$exe failed with exit code $LASTEXITCODE" }
}

function Export-Catalog {
    if (-not (Get-Command railway -ErrorAction SilentlyContinue)) {
        throw 'Railway CLI is required. Install it and run railway login in your own terminal.'
    }
    $pgDump = Get-Command pg_dump -ErrorAction SilentlyContinue
    if (-not $pgDump) {
        throw 'PostgreSQL 18 client (pg_dump) is required on PATH. See production-postgres-local.md.'
    }
    if ((& $pgDump.Source --version) -notmatch '\b18\.') {
        throw 'pg_dump major version 18 is required for Railway PostgreSQL 18.'
    }
    New-Item -ItemType Directory -Path $seedDir -Force | Out-Null

    # Keep credentials in this process only; never write Railway variables to a file.
    $json = & railway variable list --service Postgres --environment production --project $projectId --json
    if ($LASTEXITCODE -ne 0) { throw 'Could not read Railway variables. Run railway login locally, then retry.' }
    $variables = ($json | Out-String | ConvertFrom-Json)
    if ($null -ne $variables.PSObject.Properties['variables']) { $variables = $variables.variables }
    $url = $variables.DATABASE_URL
    if (-not $url) { throw 'DATABASE_URL is unavailable or sealed in Railway.' }
    $databaseUri = [Uri] $url
    $credentials = $databaseUri.UserInfo.Split(':', 2)
    if ($credentials.Count -ne 2) { throw 'Railway DATABASE_URL has no user/password pair.' }

    $stdout = Join-Path $seedDir 'tunnel.stdout'
    $stderr = Join-Path $seedDir 'tunnel.stderr'
    $tunnel = $null
    $port = 15432
    if (Test-NetConnection -ComputerName 127.0.0.1 -Port $port -InformationLevel Quiet -WarningAction SilentlyContinue) {
        throw 'Local port 15432 is already in use. Close the other process before exporting.'
    }
    try {
        $tunnel = Start-Process -FilePath (Get-Command railway).Source -ArgumentList @(
            'connect', 'Postgres', '--tunnel-only', '--port', "$port",
            '--project', $projectId, '--environment', 'production'
        ) -WindowStyle Hidden -RedirectStandardOutput $stdout -RedirectStandardError $stderr -PassThru

        $ready = $false
        for ($attempt = 0; $attempt -lt 80; $attempt++) {
            if ($tunnel.HasExited) { throw 'Railway tunnel stopped before it became ready.' }
            $client = [System.Net.Sockets.TcpClient]::new()
            try {
                $client.Connect('127.0.0.1', $port)
                $ready = $true
                break
            } catch {
                Start-Sleep -Milliseconds 250
            } finally {
                $client.Dispose()
            }
        }
        if (-not $ready) { throw 'Railway tunnel did not open on 127.0.0.1:15432.' }

        $env:PGPASSWORD = [Uri]::UnescapeDataString($credentials[1])
        $env:PGSSLMODE = 'require'
        $env:PGOPTIONS = '-c default_transaction_read_only=on'
        $common = @('-h', '127.0.0.1', '-p', "$port", '-U',
            [Uri]::UnescapeDataString($credentials[0]), '-d', $databaseUri.AbsolutePath.TrimStart('/'),
            '--format=custom', '--no-owner', '--no-acl')
        $newSchema = "$schemaFile.new"
        $newData = "$dataFile.new"
        Invoke-Checked $pgDump.Source ($common + @('--schema-only', "--file=$newSchema"))
        Invoke-Checked $pgDump.Source ($common + @('--data-only', "--file=$newData") +
            ($tables | ForEach-Object { "--table=public.$_" }))
        Move-Item -LiteralPath $newSchema -Destination $schemaFile -Force
        Move-Item -LiteralPath $newData -Destination $dataFile -Force
        [pscustomobject]@{
            capturedAtUtc = (Get-Date).ToUniversalTime().ToString('o')
            project = 'hexascore'; environment = 'production'; service = 'Postgres'
            tables = $tables
        } | ConvertTo-Json -Depth 3 | Set-Content -LiteralPath (Join-Path $seedDir 'manifest.json') -Encoding utf8
        Write-Host "Saved the production catalogue seed in $seedDir"
    } finally {
        Remove-Item Env:\PGPASSWORD, Env:\PGSSLMODE, Env:\PGOPTIONS -ErrorAction SilentlyContinue
        if ($tunnel -and -not $tunnel.HasExited) { Stop-Process -Id $tunnel.Id -Force }
        Remove-Item -LiteralPath $stdout, $stderr -Force -ErrorAction SilentlyContinue
    }
}

function Restore-Catalog {
    if (-not (Test-Path -LiteralPath $schemaFile) -or -not (Test-Path -LiteralPath $dataFile)) {
        throw "No seed found in $seedDir. Run -Action Export first."
    }
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) { throw 'Docker is required.' }
    Push-Location $root
    try {
        Invoke-Checked docker @('compose', 'up', '-d', '--wait', 'postgres')
        $container = (& docker compose ps -q postgres | Out-String).Trim()
        if ($LASTEXITCODE -ne 0 -or -not $container) { throw 'Local postgres container is unavailable.' }
        # Only the isolated local preview database is replaced.
        Invoke-Checked docker @('exec', $container, 'dropdb', '-U', 'hexascore', '--if-exists', '--force', $previewDb)
        Invoke-Checked docker @('exec', $container, 'createdb', '-U', 'hexascore', $previewDb)
        Invoke-Checked docker @('cp', $schemaFile, "${container}:/tmp/production-schema.dump")
        Invoke-Checked docker @('cp', $dataFile, "${container}:/tmp/production-catalog.dump")
        Invoke-Checked docker @('exec', $container, 'pg_restore', '-U', 'hexascore', '-d', $previewDb,
            '--no-owner', '--no-acl', '--clean', '--if-exists', '--exit-on-error', '--single-transaction', '--section=pre-data',
            '/tmp/production-schema.dump')
        Invoke-Checked docker @('exec', $container, 'pg_restore', '-U', 'hexascore', '-d', $previewDb,
            '--no-owner', '--no-acl', '--exit-on-error', '--single-transaction',
            '/tmp/production-catalog.dump')
        Invoke-Checked docker @('exec', $container, 'pg_restore', '-U', 'hexascore', '-d', $previewDb,
            '--no-owner', '--no-acl', '--exit-on-error', '--single-transaction', '--section=post-data',
            '/tmp/production-schema.dump')
        # Local preview keeps 1000 items per type; the full export stays untouched in .local-seeds.
        Invoke-Checked docker @('cp', $pruneFile, "${container}:/tmp/production-prune.sql")
        Invoke-Checked docker @('exec', $container, 'psql', '-U', 'hexascore', '-d', $previewDb,
            '-v', 'ON_ERROR_STOP=1', '-f', '/tmp/production-prune.sql')
        Invoke-Checked docker @('exec', $container, 'rm', '-f', '/tmp/production-schema.dump', '/tmp/production-catalog.dump', '/tmp/production-prune.sql')
        Invoke-Checked docker @('exec', $container, 'psql', '-U', 'hexascore', '-d', $previewDb,
            '-Atc', 'SELECT count(*) FROM items')
        Write-Host "Restored production catalogue data into local database $previewDb"
    } finally {
        Pop-Location
    }
}

switch ($Action) {
    'Export' { Export-Catalog }
    'Restore' { Restore-Catalog }
    'Sync' { Export-Catalog; Restore-Catalog }
}
