# run-all.ps1 (Day 8): builds and runs the whole stack with Docker, checks it end to end,
# writes notes\run-results.md. Needs only Docker Desktop (running). Go runs inside a container.
# Add -Keep to leave the stack running afterwards.
param([switch]$Keep)
$ErrorActionPreference = 'Continue'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$Out  = Join-Path $Root 'notes\run-results.md'
$Base = 'http://localhost:8080'
$script:Results = @()
$script:Log = New-Object System.Text.StringBuilder
[void]$script:Log.AppendLine("# Day 8 run results"); [void]$script:Log.AppendLine("")
[void]$script:Log.AppendLine("Run at: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')  Machine: $env:COMPUTERNAME"); [void]$script:Log.AppendLine("")

function Save { $script:Log.ToString() | Out-File -FilePath $Out -Encoding utf8 }
function Sh([string]$cmd, [string]$dir = $Root) {
  Push-Location $dir
  $bat = Join-Path $env:TEMP 'day08-run-step.cmd'
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
  [void]$script:Log.AppendLine("## $name  [$tag]"); [void]$script:Log.AppendLine('')
  [void]$script:Log.AppendLine('```'); [void]$script:Log.AppendLine($detail); [void]$script:Log.AppendLine('```'); [void]$script:Log.AppendLine('')
  Save
}
function Finish {
  $sum = New-Object System.Text.StringBuilder
  [void]$sum.AppendLine('| Step | Result |'); [void]$sum.AppendLine('|---|---|')
  foreach ($r in $script:Results) { [void]$sum.AppendLine("| " + $r[0] + " | " + $r[1] + " |") }
  $txt = $script:Log.ToString().Replace("Machine: $env:COMPUTERNAME`r`n", "Machine: $env:COMPUTERNAME`r`n`r`n## Summary`r`n`r`n" + $sum.ToString())
  $txt | Out-File -FilePath $Out -Encoding utf8
  Write-Host ''; Write-Host "Done. Results written to $Out"
  Read-Host 'Press Enter to close'
}
# HTTP helper: returns Code, Body (string), Json (object or $null), Headers
function Http([string]$method, [string]$path, [string]$body = $null, [string]$key = $null) {
  $h = @{}
  if ($key) { $h['Idempotency-Key'] = $key }
  try {
    if ($body) { $r = Invoke-WebRequest -UseBasicParsing -Method $method -Uri ($Base + $path) -Headers $h -Body $body -ContentType 'application/json' -TimeoutSec 15 }
    else       { $r = Invoke-WebRequest -UseBasicParsing -Method $method -Uri ($Base + $path) -Headers $h -TimeoutSec 15 }
    $code = [int]$r.StatusCode; $txt = $r.Content; $hd = $r.Headers
  } catch {
    $resp = $_.Exception.Response
    if ($resp -eq $null) { return @{ Code = 0; Body = $_.Exception.Message; Json = $null; Headers = @{} } }
    $code = [int]$resp.StatusCode
    $sr = New-Object System.IO.StreamReader($resp.GetResponseStream()); $txt = $sr.ReadToEnd(); $hd = $resp.Headers
  }
  $j = $null; try { $j = $txt | ConvertFrom-Json } catch {}
  return @{ Code = $code; Body = $txt; Json = $j; Headers = $hd }
}
function WaitStatus([string]$id, [string]$want, [int]$secs) {
  $end = (Get-Date).AddSeconds($secs); $last = ''
  while ((Get-Date) -lt $end) {
    $r = Http 'GET' "/claims/$id"
    if ($r.Json) { $last = $r.Json.status; if ($last -eq $want) { return $true } }
    Start-Sleep -Seconds 2
  }
  return $false
}
function Unsent { $o = Sh 'docker compose exec -T postgres psql -U claims -d claims -tAc "select count(*) from outbox where sent_at is null"'; return $o.Out.Trim() }

# ---------------------------------------------------------------- 0 docker
$dockerOk = (Sh 'docker info').Code -eq 0
Step 'Docker running' $dockerOk ((Sh 'docker version --format "client {{.Client.Version}} / server {{.Server.Version}}"').Out)
if (-not $dockerOk) { Write-Host 'Start Docker Desktop and run again.'; Finish; exit 1 }

# ---------------------------------------------------------------- 1 go.sum + unit tests (Go inside a container)
$GO = 'docker run --rm -v "%CD%:/src" -v day08-gomod:/go/pkg/mod -w /src golang:1.25'
$tidy = Sh "$GO go mod tidy"
Step 'go mod tidy (creates go.sum)' ($tidy.Code -eq 0 -and (Test-Path (Join-Path $Root 'go.sum'))) (Tail $tidy.Out 20)
$fmt = Sh "$GO gofmt -l ."
Step 'gofmt clean (no files listed)' ($fmt.Code -eq 0 -and $fmt.Out.Trim() -eq '') ('listed: ' + $fmt.Out)
$vet = Sh "$GO go vet ./..."
Step 'go vet ./... (compiles every package incl. adapters)' ($vet.Code -eq 0) (Tail $vet.Out 30)
$ut = Sh "$GO go test ./... -race -count=1"
Step 'go test ./... -race' ($ut.Code -eq 0) (Tail $ut.Out 30)
if ($vet.Code -ne 0) { Write-Host 'Does not compile, stopping.'; Finish; exit 1 }

# ---------------------------------------------------------------- 2 stack up
$null = Sh 'docker compose down -v'
$up = Sh 'docker compose up -d --build'
Step 'docker compose up --build' ($up.Code -eq 0) (Tail $up.Out 25)
$ready = $false
for ($i = 0; $i -lt 60; $i++) { $r = Http 'GET' '/readyz'; if ($r.Code -eq 200) { $ready = $true; break }; Start-Sleep -Seconds 3 }
$ps = Sh 'docker compose ps -a'
Step 'api ready (/readyz 200)' $ready (Tail $ps.Out 15)
if (-not $ready) {
  $lg = Sh 'docker compose logs --tail 40 api worker migrate kafka-init'
  Step 'logs after failed start' $false (Tail $lg.Out 80)
  $null = Sh 'docker compose down -v'; Finish; exit 1
}
$mg = Sh 'docker compose exec -T postgres psql -U claims -d claims -tAc "select count(*) from information_schema.tables where table_name in (''claims'',''outbox'')"'
Step 'migration created claims + outbox' ($mg.Out.Trim() -eq '2') $mg.Out
$tp = Sh 'docker compose exec -T kafka /opt/kafka/bin/kafka-topics.sh --bootstrap-server kafka:9092 --list'
Step 'topics exist' (($tp.Out -match 'claims.submitted') -and ($tp.Out -match 'claims.submitted.dlq')) $tp.Out

# ---------------------------------------------------------------- 3 API behaviour
$k1 = 'run-' + [guid]::NewGuid().ToString('N').Substring(0, 12)
$good = '{"patient_ref":"FAKE-AB12","amount":150000}'
$r1 = Http 'POST' '/claims' $good $k1
Step 'POST /claims -> 201 PENDING' ($r1.Code -eq 201 -and $r1.Json.status -eq 'PENDING') $r1.Body
$id1 = $r1.Json.id
$r2 = Http 'POST' '/claims' $good $k1
Step 'same Idempotency-Key + same body -> 200, same id' ($r2.Code -eq 200 -and $r2.Json.id -eq $id1) $r2.Body
$r3 = Http 'POST' '/claims' '{"patient_ref":"FAKE-AB12","amount":999}' $k1
Step 'same key + different body -> 409' ($r3.Code -eq 409) $r3.Body
$r4 = Http 'POST' '/claims' $good
Step 'no Idempotency-Key -> 400' ($r4.Code -eq 400) $r4.Body
$r5 = Http 'POST' '/claims' '{"patient_ref":"1100100123456","amount":1}' 'x1'
Step 'real-looking patient_ref refused -> 400' ($r5.Code -eq 400) $r5.Body
Step 'worker approves claim (async)' (WaitStatus $id1 'APPROVED' 40) ((Http 'GET' "/claims/$id1").Body)

$kBig = 'run-' + [guid]::NewGuid().ToString('N').Substring(0, 12)
$rb = Http 'POST' '/claims' '{"patient_ref":"FAKE-BIG1","amount":9000000}' $kBig
Step 'worker rejects over-limit claim' ($rb.Code -eq 201 -and (WaitStatus $rb.Json.id 'REJECTED' 40)) ((Http 'GET' "/claims/$($rb.Json.id)").Body)

$ls = Http 'GET' '/claims?status=APPROVED&limit=5'
Step 'GET /claims?status=APPROVED lists it' ($ls.Code -eq 200 -and ($ls.Body -match $id1)) (Tail $ls.Body 5)
$null = Http 'GET' "/claims/$id1"
$ck = Sh "docker compose exec -T valkey valkey-cli exists claim:$id1"
Step 'GET by id is cached in Valkey' ($ck.Out.Trim() -eq '1') ("exists claim:$id1 -> " + $ck.Out)
Step 'outbox drained (0 unsent)' ((Unsent) -eq '0') ('unsent=' + (Unsent))

# ---------------------------------------------------------------- 4 DLQ
$kP = 'run-' + [guid]::NewGuid().ToString('N').Substring(0, 12)
$rp = Http 'POST' '/claims' '{"patient_ref":"FAKE-POISON","amount":100}' $kP
$dlq = Sh 'docker compose exec -T kafka /opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server kafka:9092 --topic claims.submitted.dlq --from-beginning --max-messages 1 --timeout-ms 60000'
$stay = (Http 'GET' "/claims/$($rp.Json.id)").Json.status
Step 'poison claim: retried then parked in DLQ, stays PENDING' (($dlq.Out -match $rp.Json.id) -and $stay -eq 'PENDING') ((Tail $dlq.Out 6) + "`nstatus=" + $stay)

# ---------------------------------------------------------------- 5 broker outage (outbox)
$null = Sh 'docker compose stop kafka'
$kO = 'run-' + [guid]::NewGuid().ToString('N').Substring(0, 12)
$ro = Http 'POST' '/claims' '{"patient_ref":"FAKE-OUT1","amount":2000}' $kO
Start-Sleep -Seconds 3
$pend = Unsent
Step 'Kafka down: API still accepts (201), event waits in outbox' ($ro.Code -eq 201 -and [int]$pend -ge 1) ("code=$($ro.Code) unsent=$pend")
$null = Sh 'docker compose start kafka'
$okOut = WaitStatus $ro.Json.id 'APPROVED' 120
Step 'Kafka back: relay sends, claim gets APPROVED, nothing lost' ($okOut -and (Unsent) -eq '0') ('status ok=' + $okOut + ' unsent=' + (Unsent))

# ---------------------------------------------------------------- 6 graceful stop
$null = Sh 'docker compose stop worker api'
$ec1 = (Sh 'docker inspect -f "{{.State.ExitCode}}" day08-worker-1').Out.Trim()
$ec2 = (Sh 'docker inspect -f "{{.State.ExitCode}}" day08-api-1').Out.Trim()
Step 'worker and api exit 0 on SIGTERM' ($ec1 -eq '0' -and $ec2 -eq '0') ("worker=$ec1 api=$ec2")

# ---------------------------------------------------------------- 7 log sample + cleanup
$lg = Sh 'docker compose logs --no-color --tail 12 worker'
Step 'worker log sample (info only, for the record)' $true (Tail $lg.Out 14)
if (-not $Keep) { $dn = Sh 'docker compose down -v'; Step 'docker compose down -v' ($dn.Code -eq 0) (Tail $dn.Out 6) }
Finish
