# สร้าง gateway/.rendered/kong.yml จากเทมเพลตด้วย public key ทุกตัวใน keys\*.pub (RS256, หนึ่ง kid ต่อหนึ่งรายการ)
# และสร้าง cert ของเครื่องนี้ถ้ายังไม่มี  Kong ไม่เคยเห็น private key  ต้องมี Docker (ใช้ image alpine/openssl สร้าง cert)
$ErrorActionPreference = 'Stop'
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$keysDir = Join-Path $Root 'keys'
$pubs = @(Get-ChildItem -Path $keysDir -Filter '*.pub' -ErrorAction SilentlyContinue | Sort-Object Name)
if ($pubs.Count -lt 1) { throw 'ไม่พบ keys\*.pub (รัน run-all.ps1 หรือ docker run ... day12-keygen ก่อน)' }
$items = foreach ($f in $pubs) {
  $kid = $f.BaseName
  if ($kid -notmatch '^[A-Za-z0-9_-]{1,32}$') { throw "kid ไม่ถูกต้อง: $kid" }
  $pem = [IO.File]::ReadAllText($f.FullName).Trim() -replace "`r", ''
  if ($pem -notmatch '^-----BEGIN PUBLIC KEY-----') { throw "$($f.Name) ไม่ใช่ public key (PKIX PEM)" }
  $body = ($pem -split "`n" | ForEach-Object { '          ' + $_ }) -join "`n"
  "      - key: $kid`n        algorithm: RS256`n        rsa_public_key: |`n$body"
}
$dir = Join-Path $Root 'gateway\.rendered'
New-Item -ItemType Directory -Force -Path $dir | Out-Null
$tpl = [IO.File]::ReadAllText((Join-Path $Root 'gateway\kong.yml.tpl'))
$block = ($items -join "`n")
$pat = '(?s)      # __JWT_KEYS_BEGIN__.*?# __JWT_KEYS_END__'
if (-not [regex]::IsMatch($tpl, $pat)) { throw 'kong.yml.tpl ไม่มี marker __JWT_KEYS_BEGIN__ / __JWT_KEYS_END__' }
$out = [regex]::Replace($tpl, $pat, { param($m) $block })
[IO.File]::WriteAllText((Join-Path $dir 'kong.yml'), $out, (New-Object Text.UTF8Encoding $false))

$certs = Join-Path $Root 'certs'
New-Item -ItemType Directory -Force -Path $certs | Out-Null
if (-not ((Test-Path (Join-Path $certs 'server.crt')) -and (Test-Path (Join-Path $certs 'server.key')))) {
  # cert self-signed อายุ 30 วัน สำหรับ localhost เท่านั้น (key อ่านได้ทุกคน เพราะ Kong ใน container รันเป็น user อื่น: ใช้เฉพาะเครื่องตัวเอง)
  docker run --rm -v "${certs}:/certs" --entrypoint sh alpine/openssl -c 'openssl req -x509 -newkey rsa:2048 -nodes -keyout /certs/server.key -out /certs/server.crt -days 30 -subj "/CN=localhost" -addext "subjectAltName=DNS:localhost,IP:127.0.0.1" && chmod 644 /certs/server.key /certs/server.crt'
  if ($LASTEXITCODE -ne 0) { throw 'สร้าง cert ไม่สำเร็จ' }
}
Write-Host ('rendered gateway/.rendered/kong.yml with kids: ' + (($pubs | ForEach-Object { $_.BaseName }) -join ', '))
