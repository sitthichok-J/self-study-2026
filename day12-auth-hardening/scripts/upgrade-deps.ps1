# upgrade-deps.ps1 (Day 12): bump the vulnerable dependencies found by Trivy, inside Docker (golang:1.26).
# Backs up go.mod/go.sum to notes\ first. Result goes to notes\upgrade-deps.md
param([switch]$NoPause)
$ErrorActionPreference = 'Continue'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$Out  = Join-Path $Root 'notes\upgrade-deps.md'
Push-Location $Root
if (-not (Test-Path 'notes\go.mod.before')) { Copy-Item go.mod 'notes\go.mod.before' }
if (-not (Test-Path 'notes\go.sum.before')) { Copy-Item go.sum 'notes\go.sum.before' }
$inner = 'go version && go get github.com/jackc/pgx/v5@v5.9.2 github.com/redis/go-redis/v9@v9.7.3 golang.org/x/net@v0.60.0 github.com/klauspost/compress@v1.18.7 && go mod tidy && go vet ./... && go test ./... && go list -m all'
$bat = Join-Path $env:TEMP 'day12-upgrade-step.cmd'
$cmd = 'docker run --rm -v "%CD%:/src" -v day12-go-mod:/go/pkg/mod -v day12-go-build:/root/.cache/go-build -w /src golang:1.26 sh -c "' + $inner + '"'
Set-Content -Path $bat -Value ("@echo off`r`n" + $cmd + " 2>&1") -Encoding ASCII
$o = & cmd /c $bat | Out-String
$code = $LASTEXITCODE
Pop-Location
$l = $o.TrimEnd() -split "`r?`n"
if ($l.Count -gt 120) { $l = @('... earlier lines omitted') + ($l | Select-Object -Last 120) }
$md = @('# Day 12 dependency upgrade', '', "Run at: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')", '', "exit code: $code (0 = go get, tidy, vet and test all passed)", '', '```') + $l + @('```')
$md -join "`n" | Out-File -FilePath $Out -Encoding utf8
Write-Host "exit $code - see notes\upgrade-deps.md"
if (-not $NoPause) { Read-Host 'Press Enter to close' | Out-Null }
exit $code
