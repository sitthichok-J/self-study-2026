# run-all.ps1 (Day 10): ทดสอบ Go + เทสต์ static, ยก stack (backend + Prometheus + Grafana + Jaeger) ด้วย Docker
# แล้วตรวจ auth / IDOR / metrics / trace / log จริง เขียนผลลง notes\run-results.md
# ต้องมี Docker Desktop (เปิดอยู่) อย่างเดียว ใส่ -Keep เพื่อปล่อย stack ไว้หลังรันเสร็จ  ใส่ -NoPause ถ้าไม่ต้องการรอกด Enter
param([switch]$Keep, [switch]$NoPause)
$ErrorActionPreference = 'Continue'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$Out  = Join-Path $Root 'notes\run-results.md'
$Api  = 'http://localhost:8080'
$Prom = 'http://localhost:9091'
$Jg   = 'http://localhost:16686'
$Gf   = 'http://localhost:3001'
$script:Results = @()
$script:Log = New-Object System.Text.StringBuilder
[void]$script:Log.AppendLine("# Day 10 run results"); [void]$script:Log.AppendLine("")
[void]$script:Log.AppendLine("Run at: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')  Machine: $env:COMPUTERNAME"); [void]$script:Log.AppendLine("")

function Save { $script:Log.ToString() | Out-File -FilePath $Out -Encoding utf8 }
function Sh([string]$cmd, [string]$dir = $Root) {
  Push-Location $dir
  $bat = Join-Path $env:TEMP 'day10-run-step.cmd'
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
# ไม่ให้ secret หลุดลงไฟล์ผล: ซ่อนค่าที่รู้ว่าเป็นความลับก่อนเขียนเสมอ
function Scrub([string]$s) {
  foreach ($v in @($script:Secret, $script:GfPw, $script:TokAlice, $script:TokBob, $script:TokOps)) {
    if ($v -and $v.Length -ge 8) { $s = $s.Replace($v, '[redacted]') }
  }
  return $s
}
function Step([string]$name, [bool]$ok, [string]$detail) {
  $tag = if ($ok) { 'PASS' } else { 'FAIL' }
  Write-Host ("[{0}] {1}" -f $tag, $name)
  $script:Results += ,@($name, $tag)
  [void]$script:Log.AppendLine("## $name  [$tag]"); [void]$script:Log.AppendLine('')
  [void]$script:Log.AppendLine('```'); [void]$script:Log.AppendLine((Scrub $detail)); [void]$script:Log.AppendLine('```'); [void]$script:Log.AppendLine('')
  Save
}
function Finish {
  $sum = New-Object System.Text.StringBuilder
  [void]$sum.AppendLine('| Step | Result |'); [void]$sum.AppendLine('|---|---|')
  foreach ($r in $script:Results) { [void]$sum.AppendLine("| " + $r[0] + " | " + $r[1] + " |") }
  $pass = @($script:Results | Where-Object { $_[1] -eq 'PASS' }).Count
  $head = "Machine: $env:COMPUTERNAME`r`n`r`n## Summary`r`n`r`n$pass / $($script:Results.Count) steps passed`r`n`r`n" + $sum.ToString()
  $txt = $script:Log.ToString().Replace("Machine: $env:COMPUTERNAME`r`n", $head)
  $txt | Out-File -FilePath $Out -Encoding utf8
  Write-Host ''; Write-Host "Done. $pass / $($script:Results.Count) passed. Results written to $Out"
  if (-not $NoPause) { Read-Host 'Press Enter to close' }
}
# HTTP helper (PowerShell 5.1: body ของ error อยู่ใน ErrorDetails ไม่ใช่ stream)
function Http([string]$method, [string]$url, [string]$body = $null, [hashtable]$headers = @{}) {
  try {
    if ($body) { $r = Invoke-WebRequest -UseBasicParsing -Method $method -Uri $url -Headers $headers -Body $body -ContentType 'application/json' -TimeoutSec 20 }
    else       { $r = Invoke-WebRequest -UseBasicParsing -Method $method -Uri $url -Headers $headers -TimeoutSec 20 }
    $code = [int]$r.StatusCode; $txt = $r.Content; $hd = $r.Headers
  } catch {
    $resp = $_.Exception.Response
    if ($resp -eq $null) { return @{ Code = 0; Body = $_.Exception.Message; Json = $null; Headers = @{} } }
    $code = [int]$resp.StatusCode
    $txt = ''
    if ($_.ErrorDetails -and $_.ErrorDetails.Message) { $txt = $_.ErrorDetails.Message }
    else {
      try { $st = $resp.GetResponseStream(); if ($st.CanSeek) { $st.Position = 0 }; $sr = New-Object System.IO.StreamReader($st); $txt = $sr.ReadToEnd() } catch {}
    }
    $hd = $resp.Headers
  }
  $j = $null; try { $j = $txt | ConvertFrom-Json } catch {}
  return @{ Code = $code; Body = $txt; Json = $j; Headers = $hd }
}
function Hd($r, [string]$name) { try { $v = $r.Headers[$name]; if ($v -eq $null) { return '' }; return ([string]$v).Trim() } catch { return '' } }
function Bearer([string]$tok) { return @{ 'Authorization' = "Bearer $tok" } }
function WaitStatus([string]$id, [string]$tok, [string]$want, [int]$secs) {
  $end = (Get-Date).AddSeconds($secs)
  while ((Get-Date) -lt $end) {
    $r = Http 'GET' "$Api/claims/$id" $null (Bearer $tok)
    if ($r.Json -and $r.Json.status -eq $want) { return $r }
    Start-Sleep -Seconds 2
  }
  return $null
}
function PromQ([string]$q) {
  $r = Http 'GET' ("$Prom/api/v1/query?query=" + [uri]::EscapeDataString($q))
  if ($r.Json -and $r.Json.status -eq 'success') { return $r.Json.data.result }
  return @()
}
function PromSum([string]$q) {
  $t = 0.0; foreach ($x in (PromQ $q)) { $t += [double]$x.value[1] }; return $t
}
function WaitProm([string]$q, [double]$min, [int]$secs) {
  $end = (Get-Date).AddSeconds($secs)
  while ((Get-Date) -lt $end) { $v = PromSum $q; if ($v -ge $min) { return $v }; Start-Sleep -Seconds 3 }
  return (PromSum $q)
}

# ---------------------------------------------------------------- 0 ข้อกำหนดเบื้องต้น
$dockerOk = (Sh 'docker info').Code -eq 0
Step 'Docker running' $dockerOk ((Sh 'docker version --format "client {{.Client.Version}} / server {{.Server.Version}}"').Out)
if (-not $dockerOk) { Write-Host 'Start Docker Desktop and run again.'; Finish; exit 1 }

# secret สุ่มใหม่ทุกเครื่อง เก็บใน .env (git-ignore) ไม่พิมพ์ลงจอ/ไฟล์ผล
$envFile = Join-Path $Root '.env'
function RandHex([int]$n) { $b = New-Object byte[] $n; $g = New-Object Security.Cryptography.RNGCryptoServiceProvider; $g.GetBytes($b); return (-join ($b | ForEach-Object { $_.ToString('x2') })) }
if (-not (Test-Path $envFile)) {
  Set-Content -Path $envFile -Value ("JWT_SECRET=" + (RandHex 32) + "`r`nGRAFANA_ADMIN_PASSWORD=" + (RandHex 12)) -Encoding ASCII
}
$script:Secret = ''; $script:GfPw = ''
foreach ($l in (Get-Content $envFile)) {
  if ($l -match '^JWT_SECRET=(.+)$') { $script:Secret = $Matches[1].Trim() }
  if ($l -match '^GRAFANA_ADMIN_PASSWORD=(.+)$') { $script:GfPw = $Matches[1].Trim() }
}
Step '.env มี JWT_SECRET (>=32 ตัวอักษร) และรหัส Grafana, ไฟล์ถูก git-ignore' ($script:Secret.Length -ge 32 -and $script:GfPw.Length -ge 8 -and ((Get-Content (Join-Path $Root '.gitignore')) -contains '.env')) ('JWT_SECRET length=' + $script:Secret.Length + ' (ค่าไม่แสดง)')

# ---------------------------------------------------------------- 1 เทสต์ใน container (ไม่ต้องลง Go/Python)
$GO = 'docker run --rm -v "%CD%:/src" -v day10-go-mod:/go/pkg/mod -v day10-go-build:/root/.cache/go-build -w /src golang:1.26'
$gt = Sh "$GO go test ./..."
Step 'go test ./... (unit + baseline snapshot)' ($gt.Code -eq 0) (Tail $gt.Out 30)
$gv = Sh "$GO go vet ./..."
Step 'go vet ./...' ($gv.Code -eq 0) (Tail $gv.Out 15)
$py = Sh 'docker run --rm -v "%CD%:/src" -w /src python:3.12-alpine sh -c "pip install -q pyyaml && python -m unittest discover -s tests -v"'
Step 'เทสต์ static (compose, Prometheus, dashboard, migration, secrets)' ($py.Code -eq 0 -and $py.Out -match 'OK') (Tail $py.Out 12)
if ($gt.Code -ne 0) { Write-Host 'Go tests failed, stopping.'; Finish; exit 1 }

# ---------------------------------------------------------------- 2 ยก stack
$null = Sh 'docker compose -p day08 down -v'
$null = Sh 'docker compose down -v'
$up = Sh 'docker compose up -d --build'
Step 'docker compose up --build (api, worker, postgres, valkey, kafka, prometheus, grafana, jaeger)' ($up.Code -eq 0) (Tail $up.Out 25)
$ready = $false
for ($i = 0; $i -lt 80; $i++) { $r = Http 'GET' "$Api/readyz"; if ($r.Code -eq 200) { $ready = $true; break }; Start-Sleep -Seconds 3 }
$ps = Sh 'docker compose ps -a'
Step 'api ready (GET /readyz 200) และ migration 000002 ผ่าน' $ready (Tail $ps.Out 18)
if (-not $ready) {
  $lg = Sh 'docker compose logs --tail 40 api worker migrate'
  Step 'logs after failed start' $false (Tail $lg.Out 80)
  $null = Sh 'docker compose down -v'; Finish; exit 1
}
$mg = Sh 'docker compose exec -T postgres psql -U claims -d claims -tAc "select column_name, is_nullable from information_schema.columns where table_name=''claims'' and column_name=''owner_id''"'
Step 'ตาราง claims มีคอลัมน์ owner_id (NOT NULL)' ($mg.Out -match 'owner_id\|NO') $mg.Out

# ---------------------------------------------------------------- 3 token
$null = Sh 'docker build --build-arg CMD=mint -t day10-mint .'
$env:JWT_SECRET = $script:Secret
function Mint([string]$sub, [string]$role = 'user') { $m = Sh "docker run --rm -e JWT_SECRET day10-mint -sub $sub -role $role"; return $m.Out.Trim() }
$script:TokAlice = Mint 'alice'; $script:TokBob = Mint 'bob'; $script:TokOps = Mint 'ops' 'staff'
$okTok = ($script:TokAlice -match '^[\w-]+\.[\w-]+\.[\w-]+$') -and ($script:TokBob -match '^[\w-]+\.[\w-]+\.[\w-]+$') -and ($script:TokOps -match '^[\w-]+\.[\w-]+\.[\w-]+$')
Step 'cmd/mint ออก token ได้ (alice, bob, ops=staff)' $okTok 'token ไม่แสดงในไฟล์ผล'
$short = Sh 'docker run --rm -e JWT_SECRET=short day10-mint -sub x'
Step 'secret สั้นกว่า 32 byte ถูกปฏิเสธ' ($short.Code -ne 0) $short.Out
$env:JWT_SECRET = ''
if (-not $okTok) { $null = Sh 'docker compose down -v'; Finish; exit 1 }

# ---------------------------------------------------------------- 4 authentication
$k1 = 'run-' + [guid]::NewGuid().ToString('N').Substring(0, 12)
$good = '{"patient_ref":"FAKE-AB12","amount":150000}'
$a401 = Http 'GET' "$Api/claims"
Step 'ไม่มี token -> 401 พร้อม WWW-Authenticate' ($a401.Code -eq 401 -and (Hd $a401 'WWW-Authenticate') -match 'Bearer') ("code=$($a401.Code) body=$($a401.Body)")
$parts = $script:TokAlice.Split('.')
$tamper = $parts[0] + '.' + $parts[1] + 'x.' + $parts[2]
$a2 = Http 'GET' "$Api/claims" $null (Bearer $tamper)
Step 'payload ถูกแก้ -> 401' ($a2.Code -eq 401) $a2.Body
$none = 'eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.' + $parts[1] + '.'
$a3 = Http 'GET' "$Api/claims" $null (Bearer $none)
Step 'alg=none -> 401' ($a3.Code -eq 401) $a3.Body
$a4 = Http 'GET' "$Api/claims" $null @{ 'Authorization' = 'Basic YWxpY2U6cHc=' }
Step 'scheme ไม่ใช่ Bearer -> 401' ($a4.Code -eq 401) $a4.Body
$a5 = Http 'GET' "$Api/healthz"
Step '/healthz ไม่ต้องมี token' ($a5.Code -eq 200) $a5.Body

# ---------------------------------------------------------------- 5 authorization (ลองโจมตีตัวเอง)
$idk = Bearer $script:TokAlice; $idk['Idempotency-Key'] = $k1
$r1 = Http 'POST' "$Api/claims" $good $idk
Step 'alice POST /claims -> 201 PENDING, owner_id = alice' ($r1.Code -eq 201 -and $r1.Json.status -eq 'PENDING' -and $r1.Json.owner_id -eq 'alice') $r1.Body
$id1 = $r1.Json.id
$r2 = Http 'POST' "$Api/claims" $good $idk
Step 'alice ส่ง key เดิมซ้ำ -> 200 id เดิม' ($r2.Code -eq 200 -and $r2.Json.id -eq $id1 -and (Hd $r2 'Idempotent-Replayed') -eq 'true') $r2.Body
$bk = Bearer $script:TokBob; $bk['Idempotency-Key'] = $k1
$r3 = Http 'POST' "$Api/claims" '{"patient_ref":"FAKE-ZZ99","amount":777}' $bk
Step 'bob ใช้ key เดียวกับ alice (body ต่าง) -> 201 เคลมใหม่ของ bob ไม่ใช่ 409' ($r3.Code -eq 201 -and $r3.Json.id -ne $id1 -and $r3.Json.owner_id -eq 'bob') $r3.Body
$idBob = $r3.Json.id
$x1 = Http 'GET' "$Api/claims/$id1" $null (Bearer $script:TokBob)
$unk = Http 'GET' "$Api/claims/00000000-0000-4000-8000-0000000000ff" $null (Bearer $script:TokBob)
Step 'IDOR: bob เปิดเคลมของ alice -> 404 เหมือนเคลมที่ไม่มีอยู่ และไม่มีข้อมูลหลุด' ($x1.Code -eq 404 -and $x1.Body -eq $unk.Body -and $x1.Body -notmatch 'FAKE-') ("code=$($x1.Code) body=$($x1.Body)")
$x2 = Http 'GET' "$Api/claims/$id1" $null (Bearer $script:TokAlice)
Step 'alice เปิดเคลมของตัวเอง -> 200' ($x2.Code -eq 200 -and $x2.Json.id -eq $id1) $x2.Body
$x3 = Http 'GET' "$Api/claims/$id1" $null (Bearer $script:TokOps)
Step 'ops (staff) เปิดเคลมของ alice -> 200' ($x3.Code -eq 200) $x3.Body
$l1 = Http 'GET' "$Api/claims" $null (Bearer $script:TokAlice)
$ownA = @($l1.Json.items | Where-Object { $_.owner_id -ne 'alice' }).Count
Step 'alice list เห็นเฉพาะของตัวเอง' ($l1.Code -eq 200 -and $l1.Json.items.Count -ge 1 -and $ownA -eq 0) ("items=$($l1.Json.items.Count) ของคนอื่น=$ownA")
$l2 = Http 'GET' "$Api/claims" $null (Bearer $script:TokOps)
$owners = @($l2.Json.items | ForEach-Object { $_.owner_id } | Sort-Object -Unique) -join ','
Step 'staff list เห็นของทุกคน' ($l2.Code -eq 200 -and $owners -match 'alice' -and $owners -match 'bob') ("owners: $owners")
$bad = Http 'POST' "$Api/claims" '{"patient_ref":"FAKE-AB12","amount":5,"owner_id":"bob"}' (@{ 'Authorization' = "Bearer $($script:TokAlice)"; 'Idempotency-Key' = ('own-' + $k1) })
Step 'ส่ง owner_id มาใน body -> 400 (เจ้าของมาจาก token เท่านั้น)' ($bad.Code -eq 400) $bad.Body

# ---------------------------------------------------------------- 6 header / port
$hh = Http 'GET' "$Api/healthz"
$hOk = ((Hd $hh 'X-Content-Type-Options') -eq 'nosniff') -and ((Hd $hh 'Cache-Control') -eq 'no-store') -and ((Hd $hh 'Content-Security-Policy') -match "default-src 'none'") -and ((Hd $hh 'X-Frame-Options') -eq 'DENY') -and ((Hd $hh 'Referrer-Policy') -eq 'no-referrer')
Step 'security headers ครบ' $hOk ("nosniff=$(Hd $hh 'X-Content-Type-Options') cache=$(Hd $hh 'Cache-Control') csp=$(Hd $hh 'Content-Security-Policy')")
$m1 = Http 'GET' "$Api/metrics" $null (Bearer $script:TokOps)
Step '/metrics ไม่มีบน port ของ API (404 แม้เป็น staff)' ($m1.Code -eq 404) ("code=$($m1.Code)")
$m2 = Http 'GET' 'http://localhost:9090/metrics'
Step 'port metrics (9090) ไม่ถูก publish ออกเครื่อง' ($m2.Code -eq 0) ($m2.Body)
$rid = Http 'GET' "$Api/healthz" $null @{ 'X-Request-Id' = 'run-req-123' }
Step 'X-Request-Id ที่ถูกต้องถูกส่งกลับ' ((Hd $rid 'X-Request-Id') -eq 'run-req-123') (Hd $rid 'X-Request-Id')
$rid2 = Http 'GET' "$Api/healthz" $null @{ 'X-Request-Id' = 'bad id with spaces' }
Step 'X-Request-Id แปลกๆ ถูกแทนด้วยค่าใหม่ 32 ตัว' ((Hd $rid2 'X-Request-Id') -match '^[0-9a-f]{32}$') (Hd $rid2 'X-Request-Id')
$pc = Sh 'docker compose ps --format "{{.Name}} {{.Ports}}"'
Step 'port ที่ publish ผูก 127.0.0.1 ทั้งหมด' (-not ($pc.Out -match '0\.0\.0\.0:\d+->' )) (Tail $pc.Out 20)

# ---------------------------------------------------------------- 7 worker + business metrics
$w1 = WaitStatus $id1 $script:TokAlice 'APPROVED' 40
Step 'worker ตัดสินเป็น APPROVED (ผ่านเส้นทางเดิม outbox -> Kafka -> worker)' ($w1 -ne $null) $(if ($w1) { $w1.Body } else { 'ไม่ถึง APPROVED ใน 40 วินาที' })
$ro = Bearer $script:TokAlice; $ro['Idempotency-Key'] = ('over-' + $k1)
$rej = Http 'POST' "$Api/claims" '{"patient_ref":"FAKE-OVER1","amount":600000000}' $ro
$wr = WaitStatus $rej.Json.id $script:TokAlice 'REJECTED' 40
Step 'เกิน 50,000 บาท -> REJECTED over_limit (กฎของ Day 8 ยังอยู่)' ($wr -ne $null -and $wr.Json.decision_reason -eq 'over_limit') $(if ($wr) { $wr.Body } else { 'ไม่ถึง REJECTED' })
$rp = Bearer $script:TokAlice; $rp['Idempotency-Key'] = ('poison-' + $k1)
$poison = Http 'POST' "$Api/claims" '{"patient_ref":"FAKE-POISON","amount":100}' $rp
Step 'ส่งเคลม poison (ต้องล้มทุกครั้ง) -> 201' ($poison.Code -eq 201) $poison.Body

# ---------------------------------------------------------------- 8 Prometheus / Grafana
Start-Sleep -Seconds 10
$tg = Http 'GET' "$Prom/api/v1/targets"
$up2 = @($tg.Json.data.activeTargets | Where-Object { $_.health -eq 'up' }).Count
Step 'Prometheus scrape api + worker ได้ (2 targets up)' ($up2 -eq 2) ((@($tg.Json.data.activeTargets | ForEach-Object { $_.labels.job + '=' + $_.health }) -join ', '))
$reqs = PromSum 'http_requests_total{route="/claims/{id}"}'
Step 'http_requests_total มี route template /claims/{id} (ไม่ใช่ id ดิบ)' ($reqs -gt 0) ("sum=$reqs")
$ids = @(PromQ 'http_requests_total') | ForEach-Object { $_.metric.route } | Sort-Object -Unique
Step 'label route มีแต่ template ไม่มี uuid' (-not ($ids -match '[0-9a-f]{8}-[0-9a-f]{4}')) ($ids -join ', ')
$cs = PromSum 'claims_submitted_total{result="created"}'
Step 'claims_submitted_total{result=created} >= 3' ($cs -ge 3) ("value=$cs")
$cd = PromSum 'claims_decided_total'
Step 'claims_decided_total (worker) > 0 และ REJECTED/over_limit มีอยู่' ($cd -gt 0 -and (PromSum 'claims_decided_total{status="REJECTED",reason="over_limit"}') -ge 1) ("decided=$cd")
$af = WaitProm 'auth_failures_total' 4 30
Step 'auth_failures_total นับครั้งที่ 401 (มี reason แยก)' ($af -ge 4 -and (PromSum 'auth_failures_total{reason="signature"}') -ge 1) ("total=$af")
$az = PromSum 'authz_denied_total{resource="claim"}'
Step 'authz_denied_total นับการเปิดเคลมคนอื่น' ($az -ge 1) ("value=$az")
$dlq = WaitProm 'consumer_messages_total{outcome="dlq"}' 1 60
Step 'เคลม poison เข้า DLQ: consumer_messages_total{outcome=dlq} >= 1 และ retried >= 2' ($dlq -ge 1 -and (PromSum 'consumer_messages_total{outcome="retried"}') -ge 2) ("dlq=$dlq retried=$(PromSum 'consumer_messages_total{outcome="retried"}')")
$pool = PromSum 'db_pool_max_conns{job="claims-api"}'
Step 'db_pool_max_conns มีค่า (pool gauge ทำงาน)' ($pool -gt 0) ("max=$pool")
$rules = Http 'GET' "$Prom/api/v1/rules"
$rn = @($rules.Json.data.groups | ForEach-Object { $_.rules } | ForEach-Object { $_.name })
Step 'Prometheus โหลด alert rules ครบ 7 ข้อ' ($rn.Count -eq 7 -and ($rn -contains 'HighErrorRate') -and ($rn -contains 'DeadLetterGrowing')) ($rn -join ', ')
# rules are evaluated every 5s, so the alerts can lag the counters a little: poll instead of reading once
$firing = @()
for ($k = 0; $k -lt 15; $k++) {
  $rules = Http 'GET' "$Prom/api/v1/rules"
  $firing = @($rules.Json.data.groups | ForEach-Object { $_.rules } | Where-Object { $_.state -eq 'firing' } | ForEach-Object { $_.name })
  if (($firing -contains 'DeadLetterGrowing') -and ($firing -contains 'AccessDeniedSeen')) { break }
  Start-Sleep -Seconds 3
}
Step 'alert ที่ควรดังตอนนี้: DeadLetterGrowing และ AccessDeniedSeen (จากที่ยิงทดสอบ)' (($firing -contains 'DeadLetterGrowing') -and ($firing -contains 'AccessDeniedSeen')) ('firing: ' + ($firing -join ', '))
$gb = [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes('admin:' + $script:GfPw))
$gh = Http 'GET' "$Gf/api/health"
Step 'Grafana ขึ้น (GET /api/health)' ($gh.Code -eq 200) $gh.Body
$gd = Http 'GET' "$Gf/api/dashboards/uid/claims-overview" $null @{ 'Authorization' = "Basic $gb" }
Step 'dashboard claims-overview ถูก provision (>= 15 panels)' ($gd.Code -eq 200 -and $gd.Json.dashboard.panels.Count -ge 15) ("code=$($gd.Code) panels=$($gd.Json.dashboard.panels.Count)")
$gn = Http 'GET' "$Gf/api/dashboards/uid/claims-overview"
Step 'Grafana ไม่ยอมให้เข้าโดยไม่ login' ($gn.Code -eq 401 -or $gn.Code -eq 302) ("code=$($gn.Code)")
$ds = Http 'GET' "$Gf/api/datasources" $null @{ 'Authorization' = "Basic $gb" }
$dn = @($ds.Json | ForEach-Object { $_.name }) -join ','
Step 'datasource Prometheus + Jaeger ถูก provision' ($dn -match 'Prometheus' -and $dn -match 'Jaeger') $dn
# query ทุก panel ผ่าน Prometheus จริง: ต้อง parse ได้ (status success) ทุกอัน
$bad = @()
foreach ($p in $gd.Json.dashboard.panels) { foreach ($t in $p.targets) {
  $q = Http 'GET' ("$Prom/api/v1/query?query=" + [uri]::EscapeDataString($t.expr))
  if (-not ($q.Json -and $q.Json.status -eq 'success')) { $bad += $p.title }
} }
Step 'ทุก query ใน dashboard รันบน Prometheus ได้ (status success)' ($bad.Count -eq 0) $(if ($bad.Count) { 'พัง: ' + ($bad -join ', ') } else { 'ทุก panel ผ่าน' })

# ---------------------------------------------------------------- 9 trace + log
Start-Sleep -Seconds 8
$tr = Http 'GET' "$Jg/api/traces?service=claims-api&limit=50&lookback=1h"
$cross = $null; $n = 0
foreach ($t in $tr.Json.data) {
  $n++
  $svc = @($t.processes.PSObject.Properties | ForEach-Object { $_.Value.serviceName }) | Sort-Object -Unique
  if (($svc -contains 'claims-api') -and ($svc -contains 'claims-worker')) { $cross = $t; break }
}
Step 'Jaeger มี trace ของ claims-api' ($tr.Code -eq 200 -and $n -gt 0) ("traces=$n")
$ops = if ($cross) { @($cross.spans | ForEach-Object { $_.operationName }) -join ' | ' } else { 'ไม่พบ trace ที่ข้าม service' }
Step 'trace เดียวผ่านทั้ง claims-api และ claims-worker (trace context ข้าม Kafka ผ่าน payload)' ($cross -ne $null) $ops
Step 'trace นั้นมี span ของ repo.Create, outbox publish และ consume' ($cross -ne $null -and $ops -match 'repo\.Create' -and $ops -match 'outbox publish' -and $ops -match 'consume ') $ops
$lg = Sh 'docker compose logs --no-color api worker'
$req = @($lg.Out -split "`r?`n" | Where-Object { $_ -match '"msg":"request"' -and $_ -match '"route":"/claims"' -and $_ -match '"method":"POST"' })
$hasIds = ($req.Count -gt 0) -and ($req[0] -match '"trace_id":"[0-9a-f]{32}"') -and ($req[0] -match '"request_id":"[0-9a-f]{32}"') -and ($req[0] -match '"span_id":"[0-9a-f]{16}"')
Step 'log ของ request เป็น JSON มี trace_id, span_id, request_id' $hasIds $(if ($req.Count) { $req[0] } else { 'ไม่พบบรรทัด request' })
$tid = ''; if ($req.Count -gt 0 -and $req[0] -match '"trace_id":"([0-9a-f]{32})"') { $tid = $Matches[1] }
$one = if ($tid) { Http 'GET' "$Jg/api/traces/$tid" } else { @{ Code = 0; Json = $null } }
Step 'trace_id ในบรรทัด log เปิดหา trace เดียวกันใน Jaeger ได้' ($one.Code -eq 200 -and $one.Json.data.Count -eq 1) ("trace_id=$tid")
$sec = @($lg.Out -split "`r?`n" | Where-Object { $_ -match '"security_event":"(auth_failed|access_denied)"' })
Step 'log มี security_event ทั้ง auth_failed และ access_denied' (($sec -join "`n") -match 'auth_failed' -and ($sec -join "`n") -match 'access_denied') (Tail ($sec -join "`n") 4)
$leak = @()
if ($lg.Out.Contains($script:TokAlice)) { $leak += 'token ของ alice' }
if ($lg.Out.Contains($script:Secret)) { $leak += 'JWT_SECRET' }
if ($lg.Out -match 'FAKE-AB12' -or $lg.Out -match 'FAKE-ZZ99') { $leak += 'patient_ref' }
Step 'log ของ api/worker ไม่มี token, JWT_SECRET หรือ patient_ref' ($leak.Count -eq 0) $(if ($leak.Count) { 'หลุด: ' + ($leak -join ', ') } else { 'สะอาด' })
$usr = Sh 'docker compose top api'
Step 'api รันเป็น non-root' ($usr.Out -match '65532') (Tail $usr.Out 4)

# ---------------------------------------------------------------- จบ
if ($Keep) {
  Write-Host 'Stack ยังรันอยู่: API http://localhost:8080  Grafana http://localhost:3001  Jaeger http://localhost:16686  Prometheus http://localhost:9091'
} else {
  $null = Sh 'docker compose down -v'
}
Finish
