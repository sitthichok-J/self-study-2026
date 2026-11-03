# run-all.ps1 (Day 9): build และทดสอบ frontend, ยก stack ทั้งหมด (Day 8 backend + Kong gateway + web) ด้วย Docker
# ตรวจผ่านหน้า gateway แล้วเขียนผลลง notes\run-results.md ต้องมี Docker Desktop (เปิดอยู่) อย่างเดียว
# ใส่ -Keep เพื่อปล่อย stack ไว้หลังรันเสร็จ  ต้องมีโฟลเดอร์ ..\day08-backend อยู่ข้างๆ
param([switch]$Keep)
$ErrorActionPreference = 'Continue'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$Out  = Join-Path $Root 'notes\run-results.md'
$Gw   = 'http://localhost:8000'      # gateway
$Api  = "$Gw/api"
$Web  = 'http://localhost:3000'
$script:Results = @()
$script:Log = New-Object System.Text.StringBuilder
[void]$script:Log.AppendLine("# Day 9 run results"); [void]$script:Log.AppendLine("")
[void]$script:Log.AppendLine("Run at: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')  Machine: $env:COMPUTERNAME"); [void]$script:Log.AppendLine("")

function Save { $script:Log.ToString() | Out-File -FilePath $Out -Encoding utf8 }
function Sh([string]$cmd, [string]$dir = $Root) {
  Push-Location $dir
  $bat = Join-Path $env:TEMP 'day09-run-step.cmd'
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
# HTTP helper: Url เต็ม, ส่งกลับ Code, Body, Json, Headers (ค่า header ดึงด้วย Hd)
function Http([string]$method, [string]$url, [string]$body = $null, [hashtable]$headers = @{}) {
  try {
    if ($body) { $r = Invoke-WebRequest -UseBasicParsing -Method $method -Uri $url -Headers $headers -Body $body -ContentType 'application/json' -TimeoutSec 20 }
    else       { $r = Invoke-WebRequest -UseBasicParsing -Method $method -Uri $url -Headers $headers -TimeoutSec 20 }
    $code = [int]$r.StatusCode; $txt = $r.Content; $hd = $r.Headers
  } catch {
    $resp = $_.Exception.Response
    if ($resp -eq $null) { return @{ Code = 0; Body = $_.Exception.Message; Json = $null; Headers = @{} } }
    $code = [int]$resp.StatusCode
    # PowerShell 5.1 เก็บ body ของ error ไว้ใน ErrorDetails ถ้าไปอ่านจาก stream ตรงๆ มักได้ค่าว่าง
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
function WaitStatus([string]$id, [string]$want, [int]$secs) {
  $end = (Get-Date).AddSeconds($secs)
  while ((Get-Date) -lt $end) {
    $r = Http 'GET' "$Api/claims/$id"
    if ($r.Json -and $r.Json.status -eq $want) { return $r }
    Start-Sleep -Seconds 2
  }
  return $null
}

# ---------------------------------------------------------------- 0 ข้อกำหนดเบื้องต้น
$dockerOk = (Sh 'docker info').Code -eq 0
Step 'Docker running' $dockerOk ((Sh 'docker version --format "client {{.Client.Version}} / server {{.Server.Version}}"').Out)
if (-not $dockerOk) { Write-Host 'Start Docker Desktop and run again.'; Finish; exit 1 }
$d8 = Test-Path (Join-Path $Root '..\day08-backend\docker-compose.yml')
Step 'day08-backend อยู่ข้างๆ (compose include)' $d8 ((Resolve-Path (Join-Path $Root '..') ).Path)
if (-not $d8) { Finish; exit 1 }
$cv = Sh 'docker compose version --short'
Step 'Docker Compose 2.20+ (รองรับ include)' ($cv.Code -eq 0 -and [version]($cv.Out -replace '^v','' -replace '[^0-9.].*$','') -ge [version]'2.20') $cv.Out

# ---------------------------------------------------------------- 1 frontend: typecheck, test, build (Node ใน container)
# node_modules อยู่ใน volume ของ Docker ไม่ลงไฟล์ Linux ในโฟลเดอร์ของคุณ
$NODE = 'docker run --rm -v "%CD%\web:/app" -v day09-node-modules:/app/node_modules -v day09-npm-cache:/root/.npm -w /app node:22-alpine'
$ci = Sh "$NODE npm ci --no-audit --no-fund"
Step 'npm ci' ($ci.Code -eq 0) (Tail $ci.Out 12)
$tc = Sh "$NODE npm run typecheck"
Step 'vue-tsc typecheck' ($tc.Code -eq 0) (Tail $tc.Out 20)
$vt = Sh "$NODE npm test"
Step 'vitest (unit + baseline)' ($vt.Code -eq 0 -and $vt.Out -match 'passed') (Tail $vt.Out 20)
$bd = Sh "$NODE npm run build"
Step 'vite build' ($bd.Code -eq 0) (Tail $bd.Out 15)
if ($tc.Code -ne 0 -or $vt.Code -ne 0) { Write-Host 'Frontend checks failed, stopping.'; Finish; exit 1 }

# ---------------------------------------------------------------- 2 ยก stack
# stack ของ Day 8 ใช้พอร์ต 8080/5432/6379/9094 ชุดเดียวกัน ปิดของเดิมก่อนกันพอร์ตชน
$null = Sh 'docker compose -p day08 -f ..\day08-backend\docker-compose.yml down -v'
$null = Sh 'docker compose down -v'
$up = Sh 'docker compose up -d --build'
Step 'docker compose up --build (backend + gateway + web)' ($up.Code -eq 0) (Tail $up.Out 25)
$ready = $false
for ($i = 0; $i -lt 80; $i++) { $r = Http 'GET' "$Api/readyz"; if ($r.Code -eq 200) { $ready = $true; break }; Start-Sleep -Seconds 3 }
$ps = Sh 'docker compose ps -a'
Step 'api ready ผ่าน gateway (GET /api/readyz 200)' $ready (Tail $ps.Out 18)
if (-not $ready) {
  $lg = Sh 'docker compose logs --tail 40 gateway api worker web'
  Step 'logs after failed start' $false (Tail $lg.Out 80)
  $null = Sh 'docker compose down -v'; Finish; exit 1
}

# ---------------------------------------------------------------- 3 gateway: routing และ flow ของ backend
$k1 = 'run-' + [guid]::NewGuid().ToString('N').Substring(0, 12)
$good = '{"patient_ref":"FAKE-AB12","amount":150000}'
$idk = @{ 'Idempotency-Key' = $k1 }
$r1 = Http 'POST' "$Api/claims" $good $idk
Step 'POST /api/claims ผ่าน gateway -> 201 PENDING' ($r1.Code -eq 201 -and $r1.Json.status -eq 'PENDING') $r1.Body
$id1 = $r1.Json.id
$w1 = WaitStatus $id1 'APPROVED' 40
Step 'worker ตัดสินเป็น APPROVED (GET ผ่าน gateway)' ($w1 -ne $null) $(if ($w1) { $w1.Body } else { 'ไม่ถึง APPROVED ใน 40 วินาที' })

$r2 = Http 'POST' "$Api/claims" $good $idk
Step 'Idempotency-Key ซ้ำ -> 200 id เดิม และ header Idempotent-Replayed ผ่าน gateway มาถึง' ($r2.Code -eq 200 -and $r2.Json.id -eq $id1 -and (Hd $r2 'Idempotent-Replayed') -eq 'true') ("code=$($r2.Code) replayed=$(Hd $r2 'Idempotent-Replayed')")
$r3 = Http 'POST' "$Api/claims" '{"patient_ref":"FAKE-AB12","amount":999}' $idk
Step 'key เดิม body ต่าง -> 409 (ผ่าน gateway)' ($r3.Code -eq 409) $r3.Body
$r4 = Http 'POST' "$Api/claims" $good
Step 'ไม่ส่ง Idempotency-Key -> 400' ($r4.Code -eq 400) $r4.Body

$big = Http 'POST' "$Api/claims" ('{"patient_ref":"FAKE-AB12","amount":1,"pad":"' + ('x' * 2097152) + '"}') @{ 'Idempotency-Key' = ('big-' + $k1) }
Step 'body เกิน 1 MB ถูกตัดที่ gateway -> 413' ($big.Code -eq 413) ("code=$($big.Code) " + $big.Body)

$rej = Http 'POST' "$Api/claims" '{"patient_ref":"FAKE-OVER1","amount":600000000}' @{ 'Idempotency-Key' = ('over-' + $k1) }
$wr = if ($rej.Json) { WaitStatus $rej.Json.id 'REJECTED' 40 } else { $null }
Step 'ยอดเกินวงเงิน -> REJECTED เหตุผล over_limit (แสดงใน UI ได้)' ($wr -ne $null -and $wr.Json.decision_reason -eq 'over_limit') $(if ($wr) { $wr.Body } else { $rej.Body })

$lst = Http 'GET' "$Api/claims?status=APPROVED&limit=11&offset=0"
Step 'GET /api/claims?status=&limit=11 (ที่ UI ใช้ขอ limit+1)' ($lst.Code -eq 200 -and $lst.Json.limit -eq 11 -and @($lst.Json.items).Count -ge 1) (Tail $lst.Body 6)

$nf = Http 'GET' "$Gw/nope"
Step 'path นอก /api -> 404 จาก Kong (ไม่ถึง backend)' ($nf.Code -eq 404 -and $nf.Body -match 'no Route matched') $nf.Body
$del = Http 'DELETE' "$Api/claims/$id1"
Step 'method ที่ route ไม่อนุญาต (DELETE) -> 404 จาก Kong' ($del.Code -eq 404 -and $del.Body -match 'no Route matched') $del.Body

# ---------------------------------------------------------------- 4 CORS
$pf = Http 'OPTIONS' "$Api/claims" $null @{ 'Origin' = 'http://localhost:3000'; 'Access-Control-Request-Method' = 'POST'; 'Access-Control-Request-Headers' = 'idempotency-key,content-type' }
$allowH = (Hd $pf 'Access-Control-Allow-Headers').ToLower()
Step 'CORS preflight จาก web (localhost:3000) ผ่าน และอนุญาต Idempotency-Key' (($pf.Code -in 200, 204) -and (Hd $pf 'Access-Control-Allow-Origin') -eq 'http://localhost:3000' -and $allowH.Contains('idempotency-key')) ("code=$($pf.Code) origin=$(Hd $pf 'Access-Control-Allow-Origin') headers=$allowH")
$pe = Http 'OPTIONS' "$Api/claims" $null @{ 'Origin' = 'http://evil.example'; 'Access-Control-Request-Method' = 'POST' }
# Kong ที่ตั้ง origin เดียวจะตอบ origin ที่ตั้งไว้เสมอ ไม่ว่า request มาจากไหน browser เป็นฝ่ายเทียบแล้วปฏิเสธเอง
# จึงตรวจว่า evil.example ไม่ถูกอนุญาตและไม่ใช่ wildcard
$eo = Hd $pe 'Access-Control-Allow-Origin'
Step 'origin อื่น (evil.example) ไม่ถูกอนุญาต: ไม่มี * และไม่สะท้อน origin กลับ' ($eo -ne '*' -and $eo -ne 'http://evil.example') ("code=$($pe.Code) allow-origin='$eo'")
$ro = Http 'GET' "$Api/claims/$id1" $null @{ 'Origin' = 'http://localhost:3000' }
$exp = (Hd $ro 'Access-Control-Expose-Headers').ToLower()
Step 'response จริงมี Allow-Origin และ expose Idempotent-Replayed ให้ JS อ่านได้' ((Hd $ro 'Access-Control-Allow-Origin') -eq 'http://localhost:3000' -and $exp.Contains('idempotent-replayed')) ("origin=$(Hd $ro 'Access-Control-Allow-Origin') expose=$exp")

# ---------------------------------------------------------------- 5 web (nginx)
$ix = Http 'GET' "$Web/"
Step 'web: GET / ส่ง index.html' ($ix.Code -eq 200 -and $ix.Body -match '<div id="app">') (Tail $ix.Body 8)
$deep = Http 'GET' "$Web/claims/$id1"
Step 'web: deep link /claims/<id> ได้ index.html (SPA fallback)' ($deep.Code -eq 200 -and $deep.Body -match '<div id="app">') ("code=$($deep.Code)")
$m = [regex]::Match($ix.Body, 'src="(/assets/[^"]+\.js)"')
$js = if ($m.Success) { Http 'GET' ($Web + $m.Groups[1].Value) } else { @{ Code = 0; Body = ''; Headers = @{} } }
Step 'web: asset js ได้ 200 พร้อม Cache-Control immutable' ($js.Code -eq 200 -and (Hd $js 'Cache-Control') -match 'immutable') ("file=$($m.Groups[1].Value) cache=$(Hd $js 'Cache-Control')")
# Vite แยกโค้ดเป็นหลาย chunk: ที่อยู่ gateway อยู่ใน chunk ที่ใช้ร่วมกัน ไม่ใช่ไฟล์ entry จึงไล่ดูทุก chunk ที่ entry อ้างถึง
$baked = ($js.Body -match 'localhost:8000/api'); $where = 'entry'
if (-not $baked) {
  foreach ($c in @([regex]::Matches($js.Body, 'assets/[A-Za-z0-9_.-]+\.js') | ForEach-Object { $_.Value } | Select-Object -Unique)) {
    $b = Http 'GET' "$Web/$c"
    if ($b.Body -match 'localhost:8000/api') { $baked = $true; $where = $c; break }
  }
}
Step 'web: bundle ฝังที่อยู่ gateway (localhost:8000/api)' $baked ("found=$baked in $where")
Step 'web: security header (nosniff) และ /healthz' ((Hd $ix 'X-Content-Type-Options') -eq 'nosniff' -and (Http 'GET' "$Web/healthz").Code -eq 200) ("nosniff=$(Hd $ix 'X-Content-Type-Options')")

# ---------------------------------------------------------------- 6 backend ล่ม -> gateway ตอบ 5xx แล้วกลับมาเอง
$null = Sh 'docker compose stop api'
Start-Sleep -Seconds 2
$dn = Http 'GET' "$Api/claims/$id1"
Step 'api ดับ: gateway ตอบ 502/503/504 (UI แสดง error ได้ ไม่ค้าง)' ($dn.Code -in 502, 503, 504) ("code=$($dn.Code) " + $dn.Body)
$null = Sh 'docker compose start api'
$back = $false
for ($i = 0; $i -lt 30; $i++) { if ((Http 'GET' "$Api/readyz").Code -eq 200) { $back = $true; break }; Start-Sleep -Seconds 2 }
Step 'api กลับมา: ผ่าน gateway ใช้ได้อีกครั้ง' $back ''

# ---------------------------------------------------------------- 7 rate limit (ทำท้ายสุด เพราะจะโดนบล็อกจนครบนาที)
$codes = @(); $first429 = $null
for ($i = 1; $i -le 140; $i++) {
  $x = Http 'GET' "$Api/healthz"
  $codes += $x.Code
  if ($x.Code -eq 429 -and $first429 -eq $null) { $first429 = @{ At = $i; Retry = (Hd $x 'Retry-After'); Body = $x.Body; Remaining = (Hd $x 'X-RateLimit-Remaining-Minute') } }
}
$n429 = @($codes | Where-Object { $_ -eq 429 }).Count
Step 'rate limit 120/นาที: เกินแล้วได้ 429 + Retry-After' ($first429 -ne $null -and $first429.Retry -ne '') $(if ($first429) { "ครั้งแรกที่ 429 คือคำขอที่ $($first429.At) (รวม 429 ทั้งหมด $n429 จาก 140) Retry-After=$($first429.Retry) body=$($first429.Body)" } else { 'ไม่พบ 429 ใน 140 คำขอ' })
$pf2 = Http 'GET' "$Api/healthz" $null @{ 'Origin' = 'http://localhost:3000' }
Step 'ตอน 429 ยังมี CORS header (browser อ่านข้อความ error ได้)' ($pf2.Code -eq 429 -and (Hd $pf2 'Access-Control-Allow-Origin') -eq 'http://localhost:3000') ("code=$($pf2.Code) origin=$(Hd $pf2 'Access-Control-Allow-Origin')")

# ---------------------------------------------------------------- 8 เก็บกวาด
if ($Keep) {
  Write-Host 'รอให้ rate limit รีเซ็ต (62 วินาที) ก่อนปล่อย stack ไว้...'; Start-Sleep -Seconds 62
  Step 'ปล่อย stack ไว้ (-Keep): เปิด http://localhost:3000' $true 'ปิดด้วย docker compose down -v'
} else {
  $dn2 = Sh 'docker compose down -v'
  $left = (Sh 'docker ps -a --filter "label=com.docker.compose.project=day09" -q').Out.Trim()
  Step 'docker compose down -v (ไม่เหลือ container)' ($dn2.Code -eq 0 -and $left -eq '') (Tail $dn2.Out 8)
}
Finish
