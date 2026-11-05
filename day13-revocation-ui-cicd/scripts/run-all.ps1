# run-all.ps1 (Day 13): ทดสอบ Go + เทสต์ static, สร้างคีย์ RSA, ยก stack (backend + Kong + UI + Prometheus + Grafana + Jaeger) ด้วย Docker
# แล้วตรวจ login / refresh หมุนเวียน / Kong JWT / TLS / rate limit / IDOR / metrics / trace / log จริง เขียนผลลง notes\run-results.md
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
$Gw   = 'http://localhost:8000'   # Kong (http)
$GwTls = 'https://localhost:8443' # Kong (TLS)
$Web  = 'http://localhost:3000'
$script:Results = @()
$script:Log = New-Object System.Text.StringBuilder
[void]$script:Log.AppendLine("# Day 13 run results"); [void]$script:Log.AppendLine("")
[void]$script:Log.AppendLine("Run at: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')  Machine: $env:COMPUTERNAME"); [void]$script:Log.AppendLine("")

function Save { $script:Log.ToString() | Out-File -FilePath $Out -Encoding utf8 }
function Sh([string]$cmd, [string]$dir = $Root) {
  Push-Location $dir
  $bat = Join-Path $env:TEMP 'day13-run-step.cmd'
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
  foreach ($v in @($script:Secret, $script:GfPw, $script:DemoPw, $script:TokAlice, $script:TokBob, $script:TokOps, $script:GwTok, $script:Rt1, $script:Rt2, $script:Rt3)) {
    if ($v -and $v.Length -ge 8) { $s = $s.Replace($v, '[redacted]') }
  }
  $s = [regex]::Replace($s, 'eyJ[\w-]{10,}\.[\w-]{10,}\.[\w-]*', '[jwt]')  # JWT ทุกตัวที่เผลอติดมาใน body
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
  $ws = New-Object Microsoft.PowerShell.Commands.WebRequestSession
  $hh = @{}
  foreach ($k in $headers.Keys) {
    if ($k -eq 'Cookie') {
      foreach ($part in ([string]$headers[$k]).Split(';')) {
        $nv = $part.Trim().Split('=', 2)
        if ($nv.Count -eq 2) { $ws.Cookies.Add(([uri]$url), (New-Object System.Net.Cookie($nv[0], $nv[1], '/'))) }
      }
    } else { $hh[$k] = $headers[$k] }
  }
  $headers = $hh
  try {
    if ($body) { $r = Invoke-WebRequest -UseBasicParsing -Method $method -Uri $url -Headers $headers -Body $body -ContentType 'application/json' -TimeoutSec 20 -WebSession $ws }
    else       { $r = Invoke-WebRequest -UseBasicParsing -Method $method -Uri $url -Headers $headers -TimeoutSec 20 -WebSession $ws }
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
  Set-Content -Path $envFile -Value ("GRAFANA_ADMIN_PASSWORD=" + (RandHex 12)) -Encoding ASCII
}
# .env ที่สร้างไว้ตั้งแต่ Day 10 ยังไม่มีรหัสของ demo user: เติมให้
if (-not ((Get-Content $envFile) -match '^DEMO_PASSWORD=')) { Add-Content -Path $envFile -Value ("DEMO_PASSWORD=" + (RandHex 12)) -Encoding ASCII }
$script:Secret = ''; $script:GfPw = ''; $script:DemoPw = ''; $script:GwTok = ''; $script:Rt1 = ''; $script:Rt2 = ''; $script:Rt3 = ''
foreach ($l in (Get-Content $envFile)) {
  if ($l -match '^GRAFANA_ADMIN_PASSWORD=(.+)$') { $script:GfPw = $Matches[1].Trim() }
  if ($l -match '^DEMO_PASSWORD=(.+)$') { $script:DemoPw = $Matches[1].Trim() }
}
Step '.env มีรหัส Grafana และรหัส demo user (ไม่มี JWT secret แล้ว), ไฟล์ถูก git-ignore' ($script:GfPw.Length -ge 8 -and $script:DemoPw.Length -ge 8 -and ((Get-Content (Join-Path $Root '.gitignore')) -contains '.env')) ('รหัส Grafana length=' + $script:GfPw.Length + ' รหัส demo length=' + $script:DemoPw.Length + ' (ค่าไม่แสดง)')

# คีย์ RSA ของ JWT: สร้างใหม่ทุกรอบด้วย cmd/keygen ใน Docker เก็บที่ keys\ (git-ignore) private key ไม่ไปไหนนอกจาก API
$script:SignKid = 'k1'
$keysDir = Join-Path $Root 'keys'
if (Test-Path $keysDir) { Remove-Item -Recurse -Force $keysDir }
New-Item -ItemType Directory -Force -Path $keysDir | Out-Null
$null = Sh 'docker build --build-arg CMD=keygen -t day13-keygen .'
$kg = Sh 'docker run --rm -v "%CD%\keys:/keys" day13-keygen -dir /keys -kid k1'
$kgText = ((Get-Content (Join-Path $keysDir 'k1.key') -ErrorAction SilentlyContinue) -join "`n")
Step 'สร้างคู่คีย์ RSA k1 ด้วย cmd/keygen (k1.key + k1.pub) และโฟลเดอร์ keys ถูก git-ignore' ($kg.Code -eq 0 -and (Test-Path (Join-Path $keysDir 'k1.key')) -and (Test-Path (Join-Path $keysDir 'k1.pub')) -and $kgText -match 'BEGIN PRIVATE KEY' -and ((Get-Content (Join-Path $Root '.gitignore')) -contains 'keys/')) ($kg.Out + ' (ไม่แสดงเนื้อคีย์)')
$kg2 = Sh 'docker run --rm -v "%CD%\keys:/keys" day13-keygen -dir /keys -kid k1'
Step 'keygen ไม่ยอมเขียนทับ kid เดิม (exit != 0)' ($kg2.Code -ne 0 -and $kg2.Out -match 'already exists') $kg2.Out
$PrivBody = (($kgText -split "`n") | Where-Object { $_ -notmatch '-----' }) -join ''
$PrivProbe = if ($PrivBody.Length -ge 60) { $PrivBody.Substring(20, 40) } else { '' }

# ---------------------------------------------------------------- 1 เทสต์ใน container (ไม่ต้องลง Go/Python)
$GO = 'docker run --rm -v "%CD%:/src" -v day13-go-mod:/go/pkg/mod -v day13-go-build:/root/.cache/go-build -w /src golang:1.26'
$gt = Sh "$GO go test ./..."
Step 'go test ./... (unit + baseline snapshot)' ($gt.Code -eq 0) (Tail $gt.Out 30)
$gv = Sh "$GO go vet ./..."
Step 'go vet ./...' ($gv.Code -eq 0) (Tail $gv.Out 15)
$py = Sh 'docker run --rm -v "%CD%:/src" -w /src python:3.12-alpine sh -c "pip install -q pyyaml && python -m unittest discover -s tests -v"'
Step 'เทสต์ static (compose, Prometheus, dashboard, migration, secrets)' ($py.Code -eq 0 -and $py.Out -match 'OK') (Tail $py.Out 12)
if ($gt.Code -ne 0) { Write-Host 'Go tests failed, stopping.'; Finish; exit 1 }

# ---------------------------------------------------------------- 2 ยก stack
$null = Sh 'docker compose -p day08 down -v'
$null = Sh 'docker compose -p day09 down -v'   # stack เก่าที่อาจค้างพอร์ต 8080/3000/9094 อยู่
$null = Sh 'docker compose -p day10 down -v'
$null = Sh 'docker compose -p day12 down -v'   # Day 13: stack ของ Day 12 (run-all-keep) ถ้ายังค้างอยู่
$null = Sh 'docker compose down -v'
$rg = Sh 'powershell -NoProfile -ExecutionPolicy Bypass -File scripts\render-gateway.ps1'
Step 'render config ของ Kong (ใส่ public key ทุก kid) และสร้าง cert สำหรับ localhost' ($rg.Code -eq 0 -and (Test-Path (Join-Path $Root 'gateway\.rendered\kong.yml')) -and (Test-Path (Join-Path $Root 'certs\server.crt'))) (Tail $rg.Out 8)
$up = Sh 'docker compose up -d --build'
Step 'docker compose up --build (api, worker, kong, web, postgres, valkey, kafka, prometheus, grafana, jaeger)' ($up.Code -eq 0) (Tail $up.Out 25)
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
$mu = Sh 'docker compose exec -T postgres psql -U claims -d claims -tAc "select count(*) from information_schema.tables where table_name in (''users'',''refresh_tokens'')"'
Step 'migration 000003: ตาราง users และ refresh_tokens มีอยู่' ($mu.Out.Trim() -eq '2') $mu.Out
$gwReady = $false
for ($i = 0; $i -lt 30; $i++) { $r = Http 'GET' "$Web/healthz"; $g = Http 'GET' "$Gw/api/claims"; if ($r.Code -eq 200 -and $g.Code -eq 401) { $gwReady = $true; break }; Start-Sleep -Seconds 2 }
Step 'Kong และ UI ขึ้น (UI /healthz 200, Kong ตอบ 401 เมื่อไม่มี token)' $gwReady ("web=$($r.Code) gateway=$($g.Code) body=$($g.Body)")
$ri = Sh 'docker run --rm --network day13_default -e REDIS_TEST_ADDR=valkey:6379 -v "%CD%:/src" -v day13-go-mod:/go/pkg/mod -v day13-go-build:/root/.cache/go-build -w /src golang:1.26 go test -count=1 -v -run "Revoker|Limiter" ./internal/adapter/redis'
Step 'เทสต์ adapter กับ valkey จริง (บัญชีเพิกถอน token, cut ต่อ user, rate limit) ไม่ถูก skip' ($ri.Code -eq 0 -and $ri.Out -match '--- PASS: TestRevokerTokenAndExpiry' -and $ri.Out -match '--- PASS: TestRevokerUserCut' -and $ri.Out -match '--- PASS: TestLimiterWindow' -and $ri.Out -notmatch 'SKIP') (Tail $ri.Out 25)

# ---------------------------------------------------------------- 3 token
$null = Sh 'docker build --build-arg CMD=mint -t day13-mint .'
function Mint([string]$sub, [string]$role = 'user') { $m = Sh "docker run --rm -v ""%CD%\keys:/keys:ro"" -e JWT_KEYS_DIR=/keys -e JWT_SIGNING_KID=$($script:SignKid) day13-mint -sub $sub -role $role"; return $m.Out.Trim() }
$script:TokAlice = Mint 'alice'; $script:TokBob = Mint 'bob'; $script:TokOps = Mint 'ops' 'staff'
$okTok = ($script:TokAlice -match '^[\w-]+\.[\w-]+\.[\w-]+$') -and ($script:TokBob -match '^[\w-]+\.[\w-]+\.[\w-]+$') -and ($script:TokOps -match '^[\w-]+\.[\w-]+\.[\w-]+$')
Step 'cmd/mint ออก token ได้ (alice, bob, ops=staff)' $okTok 'token ไม่แสดงในไฟล์ผล'
$nokey = Sh 'docker run --rm -e JWT_KEYS_DIR=/nonexistent day13-mint -sub x'
Step 'ไม่มีไดเรกทอรีคีย์ -> mint ปฏิเสธ (exit != 0) ไม่ออก token' ($nokey.Code -ne 0 -and $nokey.Out -notmatch 'eyJ') $nokey.Out
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
if ($PrivProbe -and $lg.Out.Contains($PrivProbe)) { $leak += 'private key' }
if ($lg.Out.Contains($script:DemoPw)) { $leak += 'DEMO_PASSWORD' }
if ($lg.Out -match 'FAKE-AB12' -or $lg.Out -match 'FAKE-ZZ99') { $leak += 'patient_ref' }
Step 'log ของ api/worker ไม่มี token, private key หรือ patient_ref' ($leak.Count -eq 0) $(if ($leak.Count) { 'หลุด: ' + ($leak -join ', ') } else { 'สะอาด' })
$usr = Sh 'docker compose top api'
Step 'api รันเป็น non-root' ($usr.Out -match '65532') (Tail $usr.Out 4)

# ---------------------------------------------------------------- 9.5 Auth ผ่าน Kong (login -> refresh หมุน -> reuse -> logout -> TLS -> rate limit)
function CookieOf($r) { $c = Hd $r 'Set-Cookie'; if ($c -match 'refresh_token=([^;]*)') { return $Matches[1] }; return '' }
function Csrf([string]$cookie) { $h = @{ 'X-Requested-With' = 'claims-web' }; if ($cookie) { $h['Cookie'] = "refresh_token=$cookie" }; return $h }
function Sub([string]$jwt) { try { $p = $jwt.Split('.')[1].Replace('-', '+').Replace('_', '/'); while ($p.Length % 4) { $p += '=' }; return ([Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($p)) | ConvertFrom-Json).sub } catch { return '' } }
function LoginBody([string]$u, [string]$p) { return (@{ username = $u; password = $p } | ConvertTo-Json -Compress) }
$LG = "$Gw/api/auth"
$sw = [Diagnostics.Stopwatch]::StartNew()
$bad1 = Http 'POST' "$LG/login" (LoginBody 'alice' 'wrong-password-123')
$t1 = $sw.ElapsedMilliseconds; $sw.Restart()
$bad2 = Http 'POST' "$LG/login" (LoginBody 'nobody-here' 'wrong-password-123')
$t2 = $sw.ElapsedMilliseconds
Step 'login ผิด: user มีอยู่กับไม่มีอยู่ ตอบ 401 ข้อความเดียวกันทุกตัวอักษร' ($bad1.Code -eq 401 -and $bad2.Code -eq 401 -and $bad1.Body -eq $bad2.Body) ("code=$($bad1.Code)/$($bad2.Code) body=$($bad1.Body) เวลา ${t1}ms vs ${t2}ms (ตัวอย่างเดียว ไม่ใช่การพิสูจน์เรื่อง timing)")
$li = Http 'POST' "$LG/login" (LoginBody 'alice' $script:DemoPw)
$script:Rt1 = CookieOf $li; $script:GwTok = [string]$li.Json.access_token
$ck = Hd $li 'Set-Cookie'
Step 'login ผ่าน Kong -> 200 access token (sub=alice) + refresh cookie HttpOnly; Secure; SameSite=Strict; Path=/api/auth' ($li.Code -eq 200 -and (Sub $script:GwTok) -eq 'alice' -and $script:Rt1.Length -ge 32 -and $ck -match 'HttpOnly' -and $ck -match 'Secure' -and $ck -match 'SameSite=Strict' -and $ck -match 'Path=/api/auth') ("code=$($li.Code) sub=$(Sub $script:GwTok) expires_in=$($li.Json.expires_in) cookie(ไม่แสดงค่า)=" + ($ck -replace 'refresh_token=[^;]*', 'refresh_token=[value]'))
Step 'refresh token ไม่อยู่ใน body ของ login (อยู่ใน cookie เท่านั้น)' ($script:Rt1.Length -ge 32 -and -not $li.Body.Contains($script:Rt1)) 'body มีแค่ access_token, token_type, expires_in'
$hash = Sh 'docker compose exec -T postgres psql -U claims -d claims -tAc "select count(*) from refresh_tokens"'
$sha = [BitConverter]::ToString((New-Object Security.Cryptography.SHA256Managed).ComputeHash([Text.Encoding]::UTF8.GetBytes($script:Rt1))).Replace('-', '').ToLower()
$rawIn = Sh ('docker compose exec -T postgres psql -U claims -d claims -tAc "select count(*) from refresh_tokens where token_hash = ' + "'" + $script:Rt1 + "'" + '"')
$hashIn = Sh ('docker compose exec -T postgres psql -U claims -d claims -tAc "select count(*) from refresh_tokens where token_hash = ' + "'" + $sha + "'" + '"')
Step 'DB เก็บเฉพาะ SHA-256 ของ refresh token (ค่าดิบค้นไม่เจอ, ค่า hash เจอ 1 แถว)' ($rawIn.Out.Trim() -eq '0' -and $hashIn.Out.Trim() -eq '1') ("rows=$($hash.Out.Trim()) raw=$($rawIn.Out.Trim()) hash=$($hashIn.Out.Trim())")

$g0 = Http 'GET' "$Gw/api/claims"
Step 'ผ่าน Kong ไม่มี token -> 401 (ตอบโดย Kong ไม่ถึง API)' ($g0.Code -eq 401 -and $g0.Body -match '"message"') $g0.Body
$g1 = Http 'GET' "$Gw/api/claims" $null (Bearer $script:GwTok)
$foreign = @($g1.Json.items | Where-Object { $_.owner_id -ne 'alice' }).Count
Step 'ผ่าน Kong ด้วย access token ของ alice -> 200 เห็นเฉพาะของตัวเอง' ($g1.Code -eq 200 -and $foreign -eq 0) ("code=$($g1.Code) items=$($g1.Json.items.Count) ของคนอื่น=$foreign")
$gp = $script:GwTok.Split('.')
$g2 = Http 'GET' "$Gw/api/claims" $null (Bearer ($gp[0] + '.' + $gp[1] + 'x.' + $gp[2]))
Step 'ผ่าน Kong payload ถูกแก้ -> 401 จาก Kong' ($g2.Code -eq 401 -and $g2.Body -match '"message"') $g2.Body
$g3 = Http 'GET' "$Gw/api/claims" $null (Bearer ('eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.' + $gp[1] + '.'))
Step 'ผ่าน Kong alg=none -> 401 จาก Kong' ($g3.Code -eq 401) $g3.Body
$qt = Mint 'qs-test'; $script:QsTok = $qt
$g4 = Http 'GET' "$Gw/api/claims?jwt=$qt"
Step 'Kong ไม่รับ token ใน query string (ใช้ token ที่ออกใหม่ ไม่ใช่ของ session)' ($g4.Code -eq 401) ("code=$($g4.Code)")
$gm = Http 'GET' "$Gw/metrics"
$gHealthDirect = Http 'GET' "$Api/healthz"
Step 'route ที่ Kong ไม่ได้เปิด (/metrics) ได้ 404 ส่วน API ตรงตอบได้เฉพาะผ่าน 127.0.0.1' ($gm.Code -eq 404 -and $gHealthDirect.Code -eq 200) ("kong /metrics=$($gm.Code) api direct /healthz=$($gHealthDirect.Code)")

$nc = Http 'POST' "$LG/refresh" $null @{ 'Cookie' = "refresh_token=$($script:Rt1)" }
Step 'refresh ไม่มี header X-Requested-With -> 403 (กัน CSRF)' ($nc.Code -eq 403) $nc.Body
$rf = Http 'POST' "$LG/refresh" $null (Csrf $script:Rt1)
$script:Rt2 = CookieOf $rf
Step 'refresh -> 200 access token ใหม่ (sub=alice) และ cookie ใหม่ไม่ซ้ำของเดิม (หมุนเวียน)' ($rf.Code -eq 200 -and (Sub ([string]$rf.Json.access_token)) -eq 'alice' -and $script:Rt2.Length -ge 32 -and $script:Rt2 -ne $script:Rt1) ("code=$($rf.Code) sub=$(Sub ([string]$rf.Json.access_token)) cookie ใหม่ต่างจากเดิม=$($script:Rt2 -ne $script:Rt1) body=$($rf.Body)")
$dl = Http 'POST' "$Api/auth/login" (LoginBody 'alice' $script:DemoPw)
$dc = CookieOf $dl
$dr = Http 'POST' "$Api/auth/refresh" $null (Csrf $dc)
Step 'เทียบ: refresh ยิงตรงเข้า API (ไม่ผ่าน Kong) หมุน token ได้ -> 200' ($dr.Code -eq 200 -and (CookieOf $dr) -ne $dc -and (CookieOf $dr).Length -ge 32) ("login=$($dl.Code) refresh=$($dr.Code) body=$($dr.Body)")
$re = Http 'POST' "$LG/refresh" $null (Csrf $script:Rt1)
Step 'ใช้ refresh token เดิมซ้ำ -> 401 (ตรวจพบการใช้ซ้ำ)' ($re.Code -eq 401) ("code=$($re.Code) body=$($re.Body)")
$re2 = Http 'POST' "$LG/refresh" $null (Csrf $script:Rt2)
Step 'หลังตรวจพบการใช้ซ้ำ token ใหม่ของสายเดียวกันถูกเพิกถอนด้วย -> 401' ($re2.Code -eq 401) ("code=$($re2.Code)")
$fam = Sh 'docker compose exec -T postgres psql -U claims -d claims -tAc "select count(*) filter (where revoked_at is not null), count(*), count(*) filter (where used_at is not null) from refresh_tokens where username = ''alice''"'
Step 'DB: แถว refresh ของ alice ที่ถูกเพิกถอนมีอย่างน้อย 1 แถว' ($fam.Out.Trim() -match '^[1-9]\d*\|[1-9]\d*\|[1-9]\d*$') ('revoked|total|used = ' + $fam.Out.Trim())
$re3 = Http 'POST' "$LG/refresh" $null (Csrf '')
Step 'refresh ไม่มี cookie -> 401' ($re3.Code -eq 401) ("code=$($re3.Code)")

$l2 = Http 'POST' "$LG/login" (LoginBody 'bob' $script:DemoPw)
$script:Rt3 = CookieOf $l2
$swap = Http 'POST' "$LG/refresh" $null (Csrf $script:Rt3)
Step 'refresh ของ bob ได้ token ในนาม bob เท่านั้น (sub มาจากแถวใน DB ไม่ใช่จาก client)' ($swap.Code -eq 200 -and (Sub ([string]$swap.Json.access_token)) -eq 'bob') ("sub=$(Sub ([string]$swap.Json.access_token))")
$script:Rt3 = CookieOf $swap
$lo = Http 'POST' "$LG/logout" $null (Csrf $script:Rt3)
$after = Http 'POST' "$LG/refresh" $null (Csrf $script:Rt3)
Step 'logout -> 204 ล้าง cookie และ refresh token นั้นใช้ต่อไม่ได้ (401)' ($lo.Code -eq 204 -and (Hd $lo 'Set-Cookie') -match 'Max-Age=0|expires=' -and $after.Code -eq 401) ("logout=$($lo.Code) refresh_after=$($after.Code)")

$pf = Http 'OPTIONS' "$LG/login" $null @{ 'Origin' = 'http://localhost:3000'; 'Access-Control-Request-Method' = 'POST'; 'Access-Control-Request-Headers' = 'content-type,x-requested-with' }
$pe = Http 'OPTIONS' "$LG/login" $null @{ 'Origin' = 'http://evil.example'; 'Access-Control-Request-Method' = 'POST' }
Step 'CORS ของ /api/auth: origin ของ UI ได้ credentials=true, origin อื่นไม่ได้ header allow' ((Hd $pf 'Access-Control-Allow-Origin') -eq 'http://localhost:3000' -and (Hd $pf 'Access-Control-Allow-Credentials') -eq 'true' -and (Hd $pe 'Access-Control-Allow-Origin') -ne 'http://evil.example') ("ui: origin=$(Hd $pf 'Access-Control-Allow-Origin') cred=$(Hd $pf 'Access-Control-Allow-Credentials')  evil: origin=$(Hd $pe 'Access-Control-Allow-Origin')")

$cacert = Join-Path $Root 'certs\server.crt'
$tls = Sh ('curl.exe -s --cacert "' + $cacert + '" -o NUL -w "%%{http_code} verify=%%{ssl_verify_result}" ' + "$GwTls/api/claims")
Step 'TLS ที่ Kong: cert ตรวจผ่านเมื่อ pin cert ของเครื่อง และตอบ 401 เมื่อไม่มี token' ($tls.Out -match '^401 verify=0') $tls.Out
$tls3 = Sh ('curl.exe -s -o NUL -w "%%{http_code} verify=%%{ssl_verify_result}" ' + "$GwTls/api/claims")
Step 'TLS: ถ้าไม่ pin cert (self-signed) curl ปฏิเสธการเชื่อมต่อ (http_code 000)' ($tls3.Out -match '^000') ($tls3.Out + ' (exit ' + $tls3.Code + ')')
$ui = Http 'GET' "$Web/"
Step 'UI (nginx) เสิร์ฟหน้า SPA และมี security header' ($ui.Code -eq 200 -and $ui.Body -match 'id="app"' -and (Hd $ui 'X-Frame-Options') -eq 'DENY') ("code=$($ui.Code) x-frame=$(Hd $ui 'X-Frame-Options')")

# rate limit ทำเป็นขั้นสุดท้ายของส่วนนี้ เพราะจะล็อก IP นี้ที่ /api/auth นาน 1 นาที
$codes = @(); $last = $null
for ($i = 0; $i -lt 30; $i++) { $x = Http 'POST' "$LG/login" (LoginBody 'alice' 'wrong-password-123'); $codes += $x.Code; if ($x.Code -eq 429) { $last = $x } }
$n429 = @($codes | Where-Object { $_ -eq 429 }).Count
Step 'rate limit ที่ /api/auth/login: ยิงผิด 30 ครั้งติดกัน -> เจอ 429 พร้อมโควต้าใน header' ($n429 -ge 1 -and $last -ne $null -and (Hd $last 'X-RateLimit-Remaining-Minute') -ne '') ("401=$(@($codes | Where-Object { $_ -eq 401 }).Count) 429=$n429 header-remaining=$(if ($last) { Hd $last 'X-RateLimit-Remaining-Minute' }) retry-after=$(if ($last) { Hd $last 'Retry-After' })")
$klog = Sh 'docker compose logs --no-color gateway api'
$kleak = 0; $kwho = @()
foreach ($pair in @(@('refresh1', $script:Rt1), @('refresh2', $script:Rt2), @('refresh3', $script:Rt3), @('demo-password', $script:DemoPw), @('access-token', $script:GwTok))) { if ($pair[1] -and $klog.Out.Contains($pair[1])) { $kleak++; $kwho += $pair[0] } }
Step 'log ของ Kong และ API ไม่มี refresh token, access token หรือรหัสผ่าน demo' ($kleak -eq 0) $(if ($kleak) { "หลุด $kleak ค่า: " + ($kwho -join ', ') + ' | บรรทัดที่เกี่ยว: ' + ((($klog.Out -split "`r?`n") | Where-Object { ($script:GwTok -and $_.Contains($script:GwTok)) -or ($script:Rt1 -and $_.Contains($script:Rt1)) } | Select-Object -First 2 | ForEach-Object { (Scrub $_) }) -join ' // ') } else { 'สะอาด' })
$sev = @($klog.Out -split "`r?`n" | Where-Object { $_ -match '"security_event":"(login_failed|refresh_reuse|csrf_header_missing|refresh_failed)"' })
Step 'log API มี security_event: login_failed, refresh_reuse, csrf_header_missing' (($sev -join "`n") -match 'login_failed' -and ($sev -join "`n") -match 'refresh_reuse' -and ($sev -join "`n") -match 'csrf_header_missing') ('login_failed=' + @($sev | Where-Object { $_ -match 'login_failed' }).Count + ' refresh_reuse=' + @($sev | Where-Object { $_ -match 'refresh_reuse' }).Count + ' csrf_header_missing=' + @($sev | Where-Object { $_ -match 'csrf_header_missing' }).Count + ' refresh_failed=' + @($sev | Where-Object { $_ -match 'refresh_failed' }).Count)

# ---------------------------------------------------------------- 10 Day 12: RS256, หมุน key, log ไม่มี query string, 429 ที่ UI อ่านได้
function B64u([byte[]]$b) { return [Convert]::ToBase64String($b).TrimEnd('=').Replace('+', '-').Replace('/', '_') }
function B64uStr([string]$t) { return B64u ([Text.Encoding]::UTF8.GetBytes($t)) }
function Hdr([string]$jwt) { try { $q = $jwt.Split('.')[0].Replace('-', '+').Replace('_', '/'); while ($q.Length % 4) { $q += '=' }; return ([Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($q)) | ConvertFrom-Json) } catch { return $null } }
function Recreate([string]$kid) {
  $env:JWT_SIGNING_KID = $kid
  $r1 = Sh 'powershell -NoProfile -ExecutionPolicy Bypass -File scripts\render-gateway.ps1'
  $r2 = Sh 'docker compose up -d --no-deps --force-recreate api gateway'
  $ok = $false
  for ($i = 0; $i -lt 60; $i++) { $a = Http 'GET' "$Api/readyz"; $g = Http 'GET' "$Gw/api/claims"; if ($a.Code -eq 200 -and $g.Code -eq 401) { $ok = $true; break }; Start-Sleep -Seconds 2 }
  return @{ Ok = ($ok -and $r1.Code -eq 0 -and $r2.Code -eq 0); Out = (Tail ($r1.Out + "`n" + $r2.Out) 8) }
}
$h1 = Hdr $script:TokAlice; $h2 = Hdr $script:GwTok
Step 'access token เป็น RS256 มี kid=k1 (ทั้ง token จาก mint และจาก login ผ่าน Kong)' ($h1 -ne $null -and $h2 -ne $null -and $h1.alg -eq 'RS256' -and $h1.kid -eq 'k1' -and $h2.alg -eq 'RS256' -and $h2.kid -eq 'k1') ("mint: alg=$($h1.alg) kid=$($h1.kid) | login: alg=$($h2.alg) kid=$($h2.kid)")
$good2 = Mint 'alice'
$gk = Http 'GET' "$Gw/api/claims" $null (Bearer $good2)
Step 'token RS256 ที่ถูกต้องผ่าน Kong -> 200 (ใช้เทียบกับ token ปลอมด้านล่าง)' ($gk.Code -eq 200) ("code=$($gk.Code)")
# โจมตี key confusion: เซ็นด้วย HMAC-SHA256 โดยใช้ public key เป็น secret แล้วอ้างว่าเป็น HS256 และเลื่อนตัวเองเป็น staff
$now = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
$pl = '{"sub":"alice","iss":"claim-tracker","aud":"claims-api","role":"staff","iat":' + $now + ',"exp":' + ($now + 600) + '}'
$in = (B64uStr '{"alg":"HS256","typ":"JWT","kid":"k1"}') + '.' + (B64uStr $pl)
$hm = New-Object Security.Cryptography.HMACSHA256 (,[IO.File]::ReadAllBytes((Join-Path $keysDir 'k1.pub')))
$forged = $in + '.' + (B64u ($hm.ComputeHash([Text.Encoding]::ASCII.GetBytes($in))))
$f1 = Http 'GET' "$Gw/api/claims" $null (Bearer $forged)
$f2 = Http 'GET' "$Api/claims" $null (Bearer $forged)
Step 'โจมตี key confusion (HS256 เซ็นด้วย public key เป็น secret, อ้าง role=staff) -> 401 ทั้งที่ Kong และ API ตรง' ($f1.Code -eq 401 -and $f2.Code -eq 401) ("kong=$($f1.Code) api=$($f2.Code)")
$fp = $good2.Split('.')
$badKid = (B64uStr '{"alg":"RS256","typ":"JWT","kid":"k9"}') + '.' + $fp[1] + '.' + $fp[2]
$u1 = Http 'GET' "$Gw/api/claims" $null (Bearer $badKid)
$u2 = Http 'GET' "$Api/claims" $null (Bearer $badKid)
Step 'token ที่อ้าง kid ที่ไม่รู้จัก (k9) -> 401 ทั้งที่ Kong และ API ตรง' ($u1.Code -eq 401 -and $u2.Code -eq 401) ("kong=$($u1.Code) api=$($u2.Code)")

# Kong access log ไม่ควรมี query string แต่ยังต้องมีบรรทัด request
$canary = 'canary-leak-check-' + [guid]::NewGuid().ToString('N').Substring(0, 8)
$null = Http 'GET' "$Gw/api/claims?access_token=$canary"
Start-Sleep -Seconds 2
$gl = Sh 'docker compose logs --no-color gateway'
$hasPath = $gl.Out -match '"GET /api/claims HTTP/[\d.]+" 401'
$hasQs = $gl.Out.Contains($canary) -or ($script:QsTok -and $gl.Out.Contains($script:QsTok)) -or ($gl.Out -match 'access_token=') -or ($gl.Out -match '/api/claims\?')
Step 'Kong access log ใช้ format noquery: ยังมีบรรทัด request (path) แต่ไม่มี query string เลย (canary และ token ที่ส่งทาง ?jwt= ก่อนหน้าไม่อยู่ใน log)' ($hasPath -and -not $hasQs) ("มีบรรทัด request=$hasPath พบ query/canary/token ใน log=$hasQs | ตัวอย่าง: " + (@($gl.Out -split "`r?`n" | Where-Object { $_ -match 'GET /api/claims' } | Select-Object -First 2) -join ' || '))

# 429 ต้องอ่านเวลารอได้จาก browser ของ UI (ข้าม origin ต้องมี Access-Control-Expose-Headers)
$o429 = $null
for ($i = 0; $i -lt 40; $i++) { $x = Http 'POST' "$LG/login" (LoginBody 'alice' 'wrong-password-123') @{ 'Origin' = 'http://localhost:3000' }; if ($x.Code -eq 429) { $o429 = $x; break } }
Step '429 ที่ /api/auth ส่ง Retry-After และ Access-Control-Expose-Headers ให้ origin ของ UI อ่านได้' ($o429 -ne $null -and (Hd $o429 'Retry-After') -match '^\d+$' -and (Hd $o429 'Access-Control-Expose-Headers') -match 'Retry-After' -and (Hd $o429 'Access-Control-Allow-Origin') -eq 'http://localhost:3000') $(if ($o429) { "code=429 Retry-After=$(Hd $o429 'Retry-After') Expose=$(Hd $o429 'Access-Control-Expose-Headers') Allow-Origin=$(Hd $o429 'Access-Control-Allow-Origin') body=$($o429.Body)" } else { 'ไม่เจอ 429 ใน 40 ครั้ง' })

# หมุน key: เพิ่ม k2, ให้ API เซ็นด้วย k2 แต่ยังรับ k1, แล้วถอด k1
$oldTok = Mint 'alice'
$rk = Sh 'docker run --rm -v "%CD%\keys:/keys" day13-keygen -dir /keys -kid k2'
$rr = Recreate 'k2'
$script:SignKid = 'k2'
Step 'หมุน key ขั้น 1: เพิ่ม k2, JWT_SIGNING_KID=k2, render Kong ใหม่ (รู้ทั้ง k1 และ k2) และยก api+gateway ใหม่' ($rk.Code -eq 0 -and $rr.Ok) ($rk.Out + "`n" + $rr.Out)
$c1 = Http 'GET' "$Gw/api/claims" $null (Bearer $oldTok)
Step 'ระหว่างหมุน key: token เดิมที่ออกด้วย k1 ยังใช้ผ่าน Kong ได้ (200)' ($c1.Code -eq 200) ("code=$($c1.Code)")
$nl = Http 'POST' "$LG/login" (LoginBody 'alice' $script:DemoPw)
$newTok = [string]$nl.Json.access_token; $nh = Hdr $newTok
$c2 = Http 'GET' "$Gw/api/claims" $null (Bearer $newTok)
Step 'login หลังหมุน key ได้ access token ที่ kid=k2 และใช้ผ่าน Kong ได้ (200)' ($nl.Code -eq 200 -and $nh -ne $null -and $nh.kid -eq 'k2' -and $nh.alg -eq 'RS256' -and $c2.Code -eq 200) ("login=$($nl.Code) kid=$($nh.kid) claims=$($c2.Code)")
$al = Sh 'docker compose logs --no-color api'
$kl = @($al.Out -split "`r?`n" | Where-Object { $_ -match 'jwt keys loaded' }) | Select-Object -Last 1
Step 'log ของ API หลัง restart บอก signing_kid=k2 และ accepted_kids มี k1 กับ k2 (ไม่มีเนื้อคีย์)' ($kl -match '"signing_kid":"k2"' -and $kl -match '"accepted_kids":\["k1","k2"\]') ([string]$kl)
Remove-Item (Join-Path $keysDir 'k1.pub'), (Join-Path $keysDir 'k1.key') -Force -ErrorAction SilentlyContinue
$rr2 = Recreate 'k2'
$d1 = Http 'GET' "$Gw/api/claims" $null (Bearer $oldTok)
$d2 = Http 'GET' "$Api/claims" $null (Bearer $oldTok)
$d3 = Http 'GET' "$Gw/api/claims" $null (Bearer $newTok)
Step 'หมุน key ขั้น 2: ถอด k1 -> token เดิมของ k1 ได้ 401 ทั้ง Kong และ API ตรง ส่วน token ของ k2 ยังได้ 200' ($rr2.Ok -and $d1.Code -eq 401 -and $d2.Code -eq 401 -and $d3.Code -eq 200 -and -not (Test-Path (Join-Path $keysDir 'k1.pub'))) ("k1 ผ่าน Kong=$($d1.Code) k1 ตรง API=$($d2.Code) k2 ผ่าน Kong=$($d3.Code)`n" + $rr2.Out)
$bl = Http 'POST' "$LG/login" (LoginBody 'bob' $script:DemoPw)
$bh = Hdr ([string]$bl.Json.access_token)
Step 'หลังถอด k1 login ยังใช้ได้ (refresh token ใน DB ไม่ผูกกับ key) และได้ token kid=k2' ($bl.Code -eq 200 -and $bh.kid -eq 'k2') ("login=$($bl.Code) kid=$($bh.kid)")
$al2 = Sh 'docker compose logs --no-color gateway api'
$pk2 = (((Get-Content (Join-Path $keysDir 'k2.key')) | Where-Object { $_ -notmatch '-----' }) -join '')
Step 'log ทั้ง Kong และ API ไม่มีเนื้อ private key ของ k1 และ k2' (-not ($PrivProbe -and $al2.Out.Contains($PrivProbe)) -and -not ($pk2.Length -ge 60 -and $al2.Out.Contains($pk2.Substring(20, 40)))) 'สะอาด'
Remove-Item Env:JWT_SIGNING_KID -ErrorAction SilentlyContinue


# ---------------------------------------------------------------- 11 Day 13: เพิกถอน access token, rate limit ต่อ user, ตรวจเคลมโดยเจ้าหน้าที่, เครื่องมือดูข้อมูล
function LoginVia([string]$user) {   # login ผ่าน Kong; Kong จำกัด 20 ครั้ง/นาที/IP จึงรอถ้าโดน 429
  for ($i = 0; $i -lt 35; $i++) { $x = Http 'POST' "$LG/login" (LoginBody $user $script:DemoPw); if ($x.Code -ne 429) { return $x }; Start-Sleep -Seconds 2 }
  return $x
}
function Valkey([string]$args1) { return (Sh ('docker compose exec -T valkey valkey-cli ' + $args1)) }

$env:JWT_SIGNING_KID = 'k2'   # Day 12 ถอด k1 ไปแล้ว: ทุกครั้งที่ยก api ใหม่ต้องชี้ไป k2
$la = LoginVia 'alice'; $tA = [string]$la.Json.access_token; $cA = CookieOf $la
$g0 = Http 'GET' "$Gw/api/claims" $null (Bearer $tA)
Step 'revocation 1: login ผ่าน Kong ได้ access token และใช้งานผ่าน Kong ได้ (200)' ($la.Code -eq 200 -and $g0.Code -eq 200) ("login=$($la.Code) claims=$($g0.Code)")
$lh = Csrf $cA; $lh['Authorization'] = "Bearer $tA"
$lo2 = Http 'POST' "$LG/logout" $null $lh
Step 'revocation 2: logout พร้อม Bearer ของ access token -> 204 และล้าง cookie' ($lo2.Code -eq 204 -and (Hd $lo2 'Set-Cookie') -match 'Max-Age=0|expires=') ("code=$($lo2.Code)")
$rv1 = Http 'GET' "$Gw/api/claims" $null (Bearer $tA)
$rv2 = Http 'GET' "$Api/claims" $null (Bearer $tA)
Step 'revocation 3: access token เดิม (ลายเซ็นและ exp ยังถูก) หลัง logout -> 401 ทันที ทั้งผ่าน Kong และยิงตรง API' ($rv1.Code -eq 401 -and $rv2.Code -eq 401 -and $rv1.Body -match '"error":"unauthorized"') ("kong=$($rv1.Code) body=$($rv1.Body) api=$($rv2.Code)")
$la2 = LoginVia 'alice'; $tA2 = [string]$la2.Json.access_token
$g3 = Http 'GET' "$Gw/api/claims" $null (Bearer $tA2)
Step 'revocation 4: login ใหม่ของ user เดียวกันได้ token ใหม่ที่ใช้งานได้ (เพิกถอนเฉพาะ token ที่ logout)' ($la2.Code -eq 200 -and $g3.Code -eq 200 -and $tA2 -ne $tA) ("login=$($la2.Code) claims=$($g3.Code)")
$keys = Valkey '--scan --pattern "rev:jti:*"'
$k1key = (($keys.Out -split "`r?`n") | Where-Object { $_ -match '^rev:jti:' } | Select-Object -First 1)
$ttl = if ($k1key) { [int]((Valkey ('ttl ' + $k1key.Trim())).Out.Trim()) } else { -1 }
Step 'revocation 5: ใน valkey มีรายการ rev:jti:<jti> ที่มี TTL (หมดเองไม่เกินอายุ token + leeway) ไม่มีตัว token เก็บอยู่' ($k1key -and $ttl -gt 0 -and $ttl -le 700 -and -not $keys.Out.Contains($tA)) ("key(jti เท่านั้น)=$k1key ttl=${ttl}s")
$rvm = WaitProm 'auth_failures_total{reason="revoked"}' 1 40
$alog = Sh 'docker compose logs --no-color api'
Step 'revocation 6: metric auth_failures_total{reason=revoked} นับ และ log API มี security_event token_revoked' ($rvm -ge 1 -and $alog.Out -match '"security_event":"token_revoked"' -and $alog.Out -match '"reason":"revoked"') ("revoked=$rvm")

# refresh token ถูกใช้ซ้ำ -> access token ที่ออกไปแล้วของ user นั้นตายด้วย
$lb = LoginVia 'bob'; $tB = [string]$lb.Json.access_token; $cB1 = CookieOf $lb
$rfb = Http 'POST' "$LG/refresh" $null (Csrf $cB1)
$cB2 = CookieOf $rfb; $tB2 = [string]$rfb.Json.access_token
$preB = Http 'GET' "$Api/claims" $null (Bearer $tB2)
$reuse = Http 'POST' "$LG/refresh" $null (Csrf $cB1)
$postB1 = Http 'GET' "$Api/claims" $null (Bearer $tB)
$postB2 = Http 'GET' "$Api/claims" $null (Bearer $tB2)
Step 'revocation 7: ใช้ refresh token ซ้ำ -> 401 และ access token ทุกตัวของ bob ที่ออกก่อนหน้า (รวมตัวที่เพิ่ง refresh) ตายด้วย (401)' ($rfb.Code -eq 200 -and $preB.Code -eq 200 -and $reuse.Code -eq 401 -and $postB1.Code -eq 401 -and $postB2.Code -eq 401) ("refresh=$($rfb.Code) ก่อนซ้ำ=$($preB.Code) ซ้ำ=$($reuse.Code) token เก่า=$($postB1.Code) token จาก refresh=$($postB2.Code)")
$cut = Valkey 'get rev:user:bob'
Step 'revocation 8: valkey มี rev:user:bob (เวลา cut เป็น unix) และ user อื่น (alice) ไม่ถูกกระทบ' ($cut.Out.Trim() -match '^\d{10}$' -and (Http 'GET' "$Api/claims" $null (Bearer $tA2)).Code -eq 200) ("rev:user:bob=$($cut.Out.Trim())")
Start-Sleep -Seconds 2   # iat มีหน่วยวินาที: token ที่ออกในวินาทีเดียวกับ cut ถูกบล็อกด้วย จึงรอข้ามวินาที
$lb2 = LoginVia 'bob'; $tB3 = [string]$lb2.Json.access_token
$nb = Http 'GET' "$Api/claims" $null (Bearer $tB3)
Step 'revocation 9: bob login ใหม่หลังเหตุการณ์ได้ token ใหม่ที่ใช้งานได้ (cut ไม่ล็อก user ถาวร)' ($lb2.Code -eq 200 -and $nb.Code -eq 200) ("login=$($lb2.Code) claims=$($nb.Code)")

# valkey ล่ม: ตรวจ blocklist ไม่ได้ -> ปิดไว้ก่อน (503) ไม่ใช่ปล่อยผ่าน; กลับมาแล้ว blocklist ยังอยู่ (AOF)
$null = Sh 'docker compose stop valkey'
$d503 = Http 'GET' "$Api/claims" $null (Bearer $tA2)
$dh = Http 'GET' "$Api/healthz"
$dbad = Http 'GET' "$Api/claims" $null (Bearer 'garbage')
Step 'valkey ล่ม: คำขอที่มี token ถูกต้อง -> 503 + Retry-After (fail closed) ส่วน /healthz ยัง 200 และ token มั่ว ยัง 401' ($d503.Code -eq 503 -and (Hd $d503 'Retry-After') -match '^\d+$' -and $dh.Code -eq 200 -and $dbad.Code -eq 401) ("claims=$($d503.Code) retry-after=$(Hd $d503 'Retry-After') healthz=$($dh.Code) garbage=$($dbad.Code)")
$null = Sh 'docker compose start valkey'
$vok = $false
for ($i = 0; $i -lt 30; $i++) { $x = Http 'GET' "$Api/claims" $null (Bearer $tA2); if ($x.Code -eq 200) { $vok = $true; break }; Start-Sleep -Seconds 2 }
$still = Http 'GET' "$Api/claims" $null (Bearer $tA)
Step 'valkey กลับมา: คำขอปกติ 200 อีกครั้ง และ token ที่ logout ไปก่อน restart ยัง 401 (blocklist อยู่รอดด้วย AOF)' ($vok -and $still.Code -eq 401) ("normal=$vok revoked-token-after-restart=$($still.Code)")

# rate limit ต่อ user: ลดเพดานเหลือ 40/นาทีเฉพาะตอนนี้ (ค่า default ของ compose คือ 100) แล้วยิงตรง API ด้วย user เฉพาะกิจ เพื่อไม่ให้ชนเพดานต่อ IP ของ Kong
$env:USER_RATE_LIMIT = '40'
$rc = Recreate 'k2'
Step 'ยก api ใหม่ด้วย USER_RATE_LIMIT=40 (เฉพาะรอบทดสอบ)' $rc.Ok $rc.Out
$tR = Mint 'rl-user'
$rl = @(); $lastRl = $null
for ($i = 0; $i -lt 45; $i++) { $x = Http 'GET' "$Api/claims" $null (Bearer $tR); $rl += $x.Code; if ($x.Code -eq 429) { $lastRl = $x } }
$ok30 = @($rl | Where-Object { $_ -eq 200 }).Count; $n429u = @($rl | Where-Object { $_ -eq 429 }).Count
Step 'rate limit ต่อ user: เพดาน 40/นาที -> 40 ครั้งแรกผ่าน อีก 5 ครั้งโดน 429 พร้อม Retry-After' ($ok30 -eq 40 -and $n429u -eq 5 -and $lastRl -ne $null -and (Hd $lastRl 'Retry-After') -match '^\d+$' -and $lastRl.Body -match '"error":"too many requests"') ("200=$ok30 429=$n429u retry-after=$(Hd $lastRl 'Retry-After') body=$($lastRl.Body)")
$other = Http 'GET' "$Api/claims" $null (Bearer (Mint 'someone-else'))
$hz = Http 'GET' "$Api/healthz"
Step 'rate limit ต่อ user: user อื่นไม่โดนด้วย และ /healthz ไม่นับโควต้า' ($other.Code -eq 200 -and $hz.Code -eq 200) ("other=$($other.Code) healthz=$($hz.Code)")

# ตรวจเคลมโดยเจ้าหน้าที่: ยอด 20,000.01-50,000 บาท worker ไม่ตัดสินเอง
$tAl = Mint 'alice'; $tOps = Mint 'ops' 'staff'
$ra = Bearer $tAl; $ra['Idempotency-Key'] = 'rev-' + $k1
$mc = Http 'POST' "$Api/claims" '{"patient_ref":"FAKE-REV01","amount":3000000}' $ra
$mid = [string]$mc.Json.id
$rv = $null
for ($i = 0; $i -lt 20; $i++) { $x = Http 'GET' "$Api/claims/$mid" $null (Bearer $tAl); if ($x.Json -and $x.Json.decision_reason -eq 'needs_review') { $rv = $x; break }; Start-Sleep -Seconds 2 }
Step 'ตรวจเคลม 1: เคลม 30,000 บาท -> worker ตั้ง PENDING เหตุผล needs_review และไม่เปลี่ยนต่อ' ($mc.Code -eq 201 -and $rv -ne $null -and $rv.Json.status -eq 'PENDING') $(if ($rv) { $rv.Body } else { 'ไม่ถึง needs_review ใน 40 วินาที' })
$a403 = Http 'POST' "$Api/claims/$mid/approve" $null (Bearer $tAl)
Step 'ตรวจเคลม 2: เจ้าของเคลม (ไม่ใช่ staff) กดอนุมัติเอง -> 403 และสถานะไม่เปลี่ยน' ($a403.Code -eq 403 -and (Http 'GET' "$Api/claims/$mid" $null (Bearer $tAl)).Json.status -eq 'PENDING') ("code=$($a403.Code) body=$($a403.Body)")
$a200 = Http 'POST' "$Api/claims/$mid/approve" $null (Bearer $tOps)
Step 'ตรวจเคลม 3: staff อนุมัติ -> 200 APPROVED เหตุผล manual_approved' ($a200.Code -eq 200 -and $a200.Json.status -eq 'APPROVED' -and $a200.Json.decision_reason -eq 'manual_approved') $a200.Body
$a409 = Http 'POST' "$Api/claims/$mid/reject" $null (Bearer $tOps)
$aget = Http 'GET' "$Api/claims/$mid" $null (Bearer $tAl)
Step 'ตรวจเคลม 4: staff กดไม่อนุมัติซ้ำหลังอนุมัติแล้ว -> 409 และผลแรกยังอยู่ (เจ้าของเห็น APPROVED)' ($a409.Code -eq 409 -and $aget.Json.status -eq 'APPROVED') ("reject=$($a409.Code) status=$($aget.Json.status)")
$edge = Http 'POST' "$Api/claims" '{"patient_ref":"FAKE-EDGE1","amount":2000000}' (@{ 'Authorization' = "Bearer $tAl"; 'Idempotency-Key' = ('edge-' + $k1) })
$edgeOk = WaitStatus ([string]$edge.Json.id) $tAl 'APPROVED' 40
Step 'ตรวจเคลม 5: ยอด 20,000 บาทพอดียังอนุมัติอัตโนมัติ (ขอบเขตของกฎใหม่)' ($edge.Code -eq 201 -and $edgeOk -ne $null -and $edgeOk.Json.decision_reason -eq 'ok') $(if ($edgeOk) { $edgeOk.Body } else { 'ไม่ถึง APPROVED' })
$rj = Http 'POST' "$Api/claims" '{"patient_ref":"FAKE-REV02","amount":4000000}' (@{ 'Authorization' = "Bearer $tAl"; 'Idempotency-Key' = ('rev2-' + $k1) })
$rjid = [string]$rj.Json.id
for ($i = 0; $i -lt 20; $i++) { $x = Http 'GET' "$Api/claims/$rjid" $null (Bearer $tAl); if ($x.Json -and $x.Json.decision_reason -eq 'needs_review') { break }; Start-Sleep -Seconds 2 }
$rjr = Http 'POST' "$Api/claims/$rjid/reject" $null (Bearer $tOps)
Step 'ตรวจเคลม 6: staff ไม่อนุมัติเคลมที่รอตรวจ -> 200 REJECTED เหตุผล manual_rejected' ($rjr.Code -eq 200 -and $rjr.Json.status -eq 'REJECTED' -and $rjr.Json.decision_reason -eq 'manual_rejected') $rjr.Body
$mlog = Sh 'docker compose logs --no-color api worker'
Step 'ตรวจเคลม 7: log บันทึกการตัดสินโดยคน (claim decided manually) ไม่มี patient_ref' ($mlog.Out -match 'claim decided manually' -and $mlog.Out -notmatch 'FAKE-REV0') ('พบบรรทัด claim decided manually; ไม่พบ patient_ref ใน log')

# เครื่องมือดูข้อมูล (ผูก 127.0.0.1 เท่านั้น)
$km = Http 'GET' 'http://localhost:8002/'
$ka = Http 'GET' 'http://localhost:8001/services'
$kn = @($ka.Json.data | ForEach-Object { $_.name }) -join ', '
Step 'Kong Manager (:8002) เปิดได้ และ Admin API (:8001) เห็น service auth-api, claims-api (อ่านอย่างเดียว ผูก 127.0.0.1)' ($km.Code -eq 200 -and $ka.Code -eq 200 -and $kn -match 'auth-api' -and $kn -match 'claims-api') ("manager=$($km.Code) admin=$($ka.Code) services=$kn")
$ku = Http 'GET' 'http://localhost:8085/api/clusters'
Step 'Kafka UI (:8085) เห็น cluster claims-local ในโหมด read-only' ($ku.Code -eq 200 -and $ku.Body -match 'claims-local' -and $ku.Body -match '"readOnly":true') ("code=$($ku.Code) body=" + $ku.Body.Substring(0, [Math]::Min(200, $ku.Body.Length)))
$kt = Http 'GET' 'http://localhost:8085/api/clusters/claims-local/topics'
Step 'Kafka UI เห็น topic claims.submitted และ claims.submitted.dlq' ($kt.Code -eq 200 -and $kt.Body -match 'claims\.submitted' -and $kt.Body -match 'claims\.submitted\.dlq') ("code=$($kt.Code)")
Remove-Item Env:USER_RATE_LIMIT -ErrorAction SilentlyContinue
Remove-Item Env:JWT_SIGNING_KID -ErrorAction SilentlyContinue

# ---------------------------------------------------------------- จบ
if ($Keep) {
  Write-Host 'Stack ยังรันอยู่: UI http://localhost:3000  API http://localhost:8080  Kong Manager http://localhost:8002  Kafka UI http://localhost:8085  Grafana http://localhost:3001  Jaeger http://localhost:16686  Prometheus http://localhost:9091'
} else {
  $null = Sh 'docker compose down -v'
}
Finish
