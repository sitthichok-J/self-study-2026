# สร้าง gateway/.rendered/kong.yml จากเทมเพลตด้วย JWT_SECRET ใน .env และสร้าง cert ของเครื่องนี้ถ้ายังไม่มี
# ไม่พิมพ์ secret ลงจอ  ต้องมี Docker (ใช้ image alpine/openssl สร้าง cert)
$ErrorActionPreference = 'Stop'
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$secret = ''
foreach ($l in (Get-Content (Join-Path $Root '.env'))) { if ($l -match '^JWT_SECRET=(.+)$') { $secret = $Matches[1].Trim() } }
if ($secret.Length -lt 32) { throw 'JWT_SECRET ใน .env สั้นกว่า 32 ตัวอักษร' }
$dir = Join-Path $Root 'gateway\.rendered'
New-Item -ItemType Directory -Force -Path $dir | Out-Null
$tpl = [IO.File]::ReadAllText((Join-Path $Root 'gateway\kong.yml.tpl'))
[IO.File]::WriteAllText((Join-Path $dir 'kong.yml'), $tpl.Replace('__JWT_SECRET__', $secret), (New-Object Text.UTF8Encoding $false))

$certs = Join-Path $Root 'certs'
New-Item -ItemType Directory -Force -Path $certs | Out-Null
if (-not ((Test-Path (Join-Path $certs 'server.crt')) -and (Test-Path (Join-Path $certs 'server.key')))) {
  # cert self-signed อายุ 30 วัน สำหรับ localhost เท่านั้น (key อ่านได้ทุกคน เพราะ Kong ใน container รันเป็น user อื่น: ใช้เฉพาะเครื่องตัวเอง)
  docker run --rm -v "${certs}:/certs" --entrypoint sh alpine/openssl -c 'openssl req -x509 -newkey rsa:2048 -nodes -keyout /certs/server.key -out /certs/server.crt -days 30 -subj "/CN=localhost" -addext "subjectAltName=DNS:localhost,IP:127.0.0.1" && chmod 644 /certs/server.key /certs/server.crt'
  if ($LASTEXITCODE -ne 0) { throw 'สร้าง cert ไม่สำเร็จ' }
}
Write-Host 'rendered gateway/.rendered/kong.yml and certs/'
