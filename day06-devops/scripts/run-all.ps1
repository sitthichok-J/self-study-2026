# run-all.ps1: runs the Day 6 checks on a real machine and writes notes\run-results.md
# Needs: Docker Desktop (running), Go, kind, kubectl. Installs nothing. Cleans up what it creates.
# Usage: double-click scripts\run-all.bat  (or: powershell -ExecutionPolicy Bypass -File scripts\run-all.ps1)
$ErrorActionPreference = 'Continue'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)   # ...\day06-devops
$Day5 = Join-Path (Split-Path -Parent $Root) 'day05-go-service'
$Out  = Join-Path $Root 'notes\run-results.md'
$env:PATH = (Join-Path $Root 'tools') + ';' + $env:PATH
$script:Results = @()
$script:Log = New-Object System.Text.StringBuilder
[void]$script:Log.AppendLine("# Day 6 run results")
[void]$script:Log.AppendLine("")
[void]$script:Log.AppendLine("Run at: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')  Machine: $env:COMPUTERNAME")
[void]$script:Log.AppendLine("")

function Save { $script:Log.ToString() | Out-File -FilePath $Out -Encoding utf8 }

# run a command line through cmd so stderr does not become PowerShell errors; returns @{Code;Out}
function Sh([string]$cmd, [string]$dir = $Root) {
  Push-Location $dir
  $bat = Join-Path $env:TEMP 'claims-run-step.cmd'
  Set-Content -Path $bat -Value ("@echo off`r`n" + $cmd + " 2>&1") -Encoding ASCII
  $o = & cmd /c $bat | Out-String
  $c = $LASTEXITCODE
  Pop-Location
  return @{ Code = $c; Out = $o.TrimEnd() }
}
function Tail([string]$s, [int]$n = 25) {
  $l = $s -split "`r?`n"
  if ($l.Count -gt $n) { return ("... (" + ($l.Count - $n) + " earlier lines omitted)`n" + (($l | Select-Object -Last $n) -join "`n")) }
  return $s
}
function Step([string]$name, [bool]$ok, [string]$detail) {
  $tag = if ($ok) { 'PASS' } else { 'FAIL' }
  Write-Host ("[{0}] {1}" -f $tag, $name)
  $script:Results += ,@($name, $tag)
  [void]$script:Log.AppendLine("## $name  [$tag]")
  [void]$script:Log.AppendLine('')
  [void]$script:Log.AppendLine('```')
  [void]$script:Log.AppendLine($detail)
  [void]$script:Log.AppendLine('```')
  [void]$script:Log.AppendLine('')
  Save
}
function Http([string]$method, [string]$url, [string]$body) {
  try {
    if ($body) { $r = Invoke-WebRequest -UseBasicParsing -Method $method -Uri $url -Body $body -ContentType 'application/json' -TimeoutSec 10 }
    else { $r = Invoke-WebRequest -UseBasicParsing -Method $method -Uri $url -TimeoutSec 10 }
    return @{ Code = [int]$r.StatusCode; Body = $r.Content }
  } catch {
    $code = 0; if ($_.Exception.Response) { $code = [int]$_.Exception.Response.StatusCode }
    return @{ Code = $code; Body = $_.Exception.Message }
  }
}

# ---------------------------------------------------------------- 0 tools
$v = @()
foreach ($c in @('docker version --format "docker {{.Client.Version}} / server {{.Server.Version}}"', 'go version', 'kind version', 'kubectl version --client')) {
  $r = Sh $c; $v += ("> " + $c + "`n" + (Tail $r.Out 4) + "  (exit " + $r.Code + ")")
}
$dockerOk = (Sh 'docker info').Code -eq 0
Step 'Tools present' $dockerOk ($v -join "`n`n")
if (-not $dockerOk) { Write-Host 'Docker is not running. Start Docker Desktop and run again.'; Save; Read-Host 'Press Enter to close'; exit 1 }

# ---------------------------------------------------------------- 1 Go: pgx, build, tests
$r1 = Sh 'go get github.com/jackc/pgx/v5' $Day5
$r2 = Sh 'go mod tidy' $Day5
$r3 = Sh 'go build -tags pgx ./...' $Day5
$r4 = Sh 'go vet -tags pgx ./...' $Day5
$r5 = Sh 'go test ./...' $Day5
Step 'Go: go get pgx + go mod tidy' (($r1.Code -eq 0) -and ($r2.Code -eq 0)) ((Tail $r1.Out 8) + "`n" + (Tail $r2.Out 8))
Step 'Go: build with pgx driver (-tags pgx)' ($r3.Code -eq 0) (Tail $r3.Out 20)
Step 'Go: vet (-tags pgx)' ($r4.Code -eq 0) (Tail $r4.Out 20)
Step 'Go: unit tests (go test ./...)' ($r5.Code -eq 0) (Tail $r5.Out 30)

# ---------------------------------------------------------------- 2 docker build + image sizes
$b1 = Sh "docker build -f day06-devops\Dockerfile --target build -t claims-api:build day05-go-service" (Split-Path -Parent $Root)
$b2 = Sh "docker build -f day06-devops\Dockerfile -t claims-api:dev day05-go-service" (Split-Path -Parent $Root)
$b2.Out | Out-File -FilePath (Join-Path $Root 'notes\docker-build.log') -Encoding utf8
$sz = Sh 'docker images claims-api --format "{{.Repository}}:{{.Tag}}  {{.Size}}"'
Step 'Docker: build stage image' ($b1.Code -eq 0) (Tail $b1.Out 12)
Step 'Docker: final image (distroless)' ($b2.Code -eq 0) ((Tail $b2.Out 40) + "`n`nImage sizes:`n" + $sz.Out)
$user = Sh 'docker image inspect claims-api:dev --format "user={{.Config.User}}"'
Step 'Docker: final image runs as non-root' ($user.Out -match 'nonroot') $user.Out

# ---------------------------------------------------------------- 3 compose
$up = Sh 'docker compose up --build -d'
$ready = $false; $hz = $null
for ($i = 0; $i -lt 45; $i++) { $hz = Http 'GET' 'http://localhost:8080/healthz' ''; if ($hz.Code -eq 200) { $ready = $true; break }; Start-Sleep -Seconds 2 }
$ps = Sh 'docker compose ps'
Step 'Compose: up --build (db healthy, api answers /healthz)' ($up.Code -eq 0 -and $ready) ((Tail $up.Out 10) + "`n`n" + $ps.Out + "`n`n/healthz -> " + $hz.Code + " " + $hz.Body)
if ($ready) {
  $p = Http 'POST' 'http://localhost:8080/claims' '{"patient_ref":"P001","hospital":"Test Hospital","amount_satang":150000}'
  $g = Http 'GET' 'http://localhost:8080/claims' ''
  Step 'Compose: create + list claim through API and Postgres' (($p.Code -eq 201 -or $p.Code -eq 200) -and $g.Body -match 'P001') ("POST /claims -> " + $p.Code + " " + $p.Body + "`nGET /claims -> " + $g.Code + " " + $g.Body)
  $null = Sh 'docker compose restart api'
  Start-Sleep -Seconds 6
  $g2 = Http 'GET' 'http://localhost:8080/claims' ''
  Step 'Compose: data survives api restart (stored in Postgres)' ($g2.Body -match 'P001') ("GET /claims after restart -> " + $g2.Code + " " + $g2.Body)
  $null = Sh 'docker compose stop api'
  $lg = Sh 'docker compose logs api --no-log-prefix --tail 15'
  Step 'Compose: graceful shutdown log on stop' (($lg.Out -match 'shutting down') -or ($lg.Out -match 'shutdown')) $lg.Out
}
$null = Sh 'docker compose down -v'

# ---------------------------------------------------------------- 4 kind + kubernetes
$null = Sh 'kind delete cluster --name claims'
$kc = Sh 'kind create cluster --name claims --image kindest/node:v1.34.0 --retain'
$kc.Out | Out-File -FilePath (Join-Path $Root 'notes\kind-create.log') -Encoding utf8
$null = Sh 'kind export logs notes\kind-logs --name claims'
if ($kc.Code -ne 0) { $null = Sh 'kind delete cluster --name claims' }
Step 'kind: create cluster' ($kc.Code -eq 0) (Tail $kc.Out 12)
if ($kc.Code -eq 0) {
  $ld = Sh 'kind load docker-image claims-api:dev --name claims'
  Step 'kind: load image' ($ld.Code -eq 0) (Tail $ld.Out 6)
  $ap = Sh 'kubectl apply -k day06-devops\k8s' (Split-Path -Parent $Root)
  Step 'kubectl apply -k k8s' ($ap.Code -eq 0) $ap.Out
  $w = Sh 'kubectl -n claims wait --for=condition=ready pod --all --timeout=240s'
  $pods = Sh 'kubectl -n claims get pods'
  $ds = if ($w.Code -ne 0) { (Sh 'kubectl -n claims describe pod -l app=claims-api').Out } else { '' }
  Step 'k8s: all pods Ready' ($w.Code -eq 0) ($w.Out + "`n`n" + $pods.Out + "`n`n" + (Tail $ds 25))

  # self-healing
  $old = (Sh 'kubectl -n claims get pods -l app=claims-api --no-headers').Out.Trim()
  $oldName = ($old -split '\s+')[0]
  $null = Sh 'kubectl -n claims delete pod -l app=claims-api --wait=false'
  $sw = [Diagnostics.Stopwatch]::StartNew(); $healed = $false
  while ($sw.Elapsed.TotalSeconds -lt 120) {
    Start-Sleep -Seconds 2
    $cur = (Sh 'kubectl -n claims get pods -l app=claims-api --no-headers').Out
    foreach ($ln in ($cur -split "`r?`n")) { if ($ln -match '1/1\s+Running' -and $ln -notmatch [regex]::Escape($oldName)) { $healed = $true } }
    if ($healed) { break }
  }
  $pods2 = Sh 'kubectl -n claims get pods'
  Step 'k8s: delete api pod -> new pod created and Ready' $healed ("old pod: $oldName`nseconds until a new Ready pod: " + [math]::Round($sw.Elapsed.TotalSeconds, 1) + "`n`n" + $pods2.Out)

  # service via port-forward
  $pf = Start-Process -FilePath 'kubectl' -ArgumentList '-n claims port-forward svc/claims-api 18080:80' -PassThru -WindowStyle Hidden
  Start-Sleep -Seconds 6
  $hk = Http 'GET' 'http://localhost:18080/healthz' ''
  Step 'k8s: Service reachable via port-forward (/healthz)' ($hk.Code -eq 200) ("-> " + $hk.Code + " " + $hk.Body)
  try { Stop-Process -Id $pf.Id -Force -ErrorAction SilentlyContinue } catch {}

  # readiness fail
  $pj = Join-Path $env:TEMP 'claims-patch.json'
  Set-Content -Path $pj -Value '[{"op":"replace","path":"/spec/template/spec/containers/0/readinessProbe/httpGet/path","value":"/nope"}]' -Encoding ASCII
  $pt = Sh ('kubectl -n claims patch deployment claims-api --type=json --patch-file "' + $pj + '"')
  Start-Sleep -Seconds 40
  $pods3 = Sh 'kubectl -n claims get pods'
  $ep = Sh 'kubectl -n claims get endpoints claims-api'
  $oldStillServes = ($ep.Out -match '\d+\.\d+\.\d+\.\d+')
  Step 'k8s: readiness failing on new pods (old pod keeps serving, maxUnavailable 0)' ($pt.Code -eq 0) ("patch: " + $pt.Out + "`n`n" + $pods3.Out + "`n`n" + $ep.Out + "`n`nendpoints still has an address: $oldStillServes")
  $un = Sh 'kubectl -n claims rollout undo deployment/claims-api'
  $st = Sh 'kubectl -n claims rollout status deployment/claims-api --timeout=150s'
  Step 'k8s: rollout undo recovers' ($st.Code -eq 0) ($un.Out + "`n" + $st.Out)

  # graceful shutdown in k8s
  $lf = Join-Path $env:TEMP 'claims-api-follow.log'
  if (Test-Path $lf) { Remove-Item $lf -Force }
  $fl = Start-Process -FilePath 'kubectl' -ArgumentList '-n claims logs -f deploy/claims-api' -PassThru -WindowStyle Hidden -RedirectStandardOutput $lf
  Start-Sleep -Seconds 3
  $null = Sh 'kubectl -n claims rollout restart deployment/claims-api'
  $st2 = Sh 'kubectl -n claims rollout status deployment/claims-api --timeout=150s'
  Start-Sleep -Seconds 12
  try { Stop-Process -Id $fl.Id -Force -ErrorAction SilentlyContinue } catch {}
  $fol = if (Test-Path $lf) { Get-Content $lf -Raw } else { '' }
  Step 'k8s: rollout restart logs graceful shutdown of old pod' (($fol -match 'shutting down') -or ($fol -match 'shutdown')) ((Tail $st2.Out 6) + "`n`nlast lines of followed log:`n" + (Tail $fol 15))

  $null = Sh 'kind delete cluster --name claims'
}

# ---------------------------------------------------------------- 5 CI files: static lint only
$al = Sh 'docker run --rm -v "%CD%:/repo" -w /repo rhysd/actionlint:latest -color day06-devops/ci/github-actions-ci.yml' (Split-Path -Parent $Root)
Step 'CI: actionlint on GitHub Actions file (static check, pipeline NOT executed)' ($al.Code -eq 0) ($al.Out + "`n(exit " + $al.Code + ")")

# ---------------------------------------------------------------- summary
$sum = New-Object System.Text.StringBuilder
[void]$sum.AppendLine('| Step | Result |'); [void]$sum.AppendLine('|---|---|')
foreach ($r in $script:Results) { [void]$sum.AppendLine("| " + $r[0] + " | " + $r[1] + " |") }
$txt = $script:Log.ToString()
$txt = $txt.Replace("Machine: $env:COMPUTERNAME`r`n", "Machine: $env:COMPUTERNAME`r`n`r`n## Summary`r`n`r`n" + $sum.ToString())
$txt | Out-File -FilePath $Out -Encoding utf8
Write-Host ''
Write-Host "Done. Results written to $Out"
Read-Host 'Press Enter to close'
