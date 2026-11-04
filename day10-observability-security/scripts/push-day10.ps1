# push-day10.ps1: clone sitthichok-J/self-study-2026 to TEMP, copy day10 (without .env), update the root README row, commit, push.
# Uses whatever git credentials this PC already has. Result goes to notes\push-result.md (not committed).
param([switch]$NoPause)
$ErrorActionPreference = 'Continue'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$Out  = Join-Path $Root 'notes\push-result.md'
$log = New-Object System.Collections.Generic.List[string]
function Run([string]$cmd) {
  $bat = Join-Path $env:TEMP 'day10-push-step.cmd'
  Set-Content -Path $bat -Value ("@echo off`r`n" + $cmd + " 2>&1") -Encoding ASCII
  $o = (& cmd /c $bat | Out-String).TrimEnd(); $c = $LASTEXITCODE
  $log.Add('$ ' + $cmd); $log.Add($o); $log.Add("(exit $c)"); $log.Add('')
  Write-Host ("exit {0}: {1}" -f $c, $cmd)
  return @{ Code = $c; Out = $o }
}
function Finish([string]$status) {
  $md = @('# Day 10 push result', '', "Run at: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')", '', "status: $status", '', '```') + $log + @('```')
  $md -join "`n" | Out-File -FilePath $Out -Encoding utf8
  Write-Host "status: $status (see notes\push-result.md)"
  if (-not $NoPause) { Read-Host 'Press Enter to close' | Out-Null }
}
if ((Run 'git --version').Code -ne 0) { Finish 'FAILED: git not found on this PC'; exit 1 }
$mail = (Run 'git config --global user.email').Out
if (-not $mail) { Finish 'FAILED: git user.email is not set on this PC (git config --global user.email ...)'; exit 1 }
$tmp = Join-Path $env:TEMP 'self-study-push'
if (Test-Path $tmp) { Remove-Item -Recurse -Force $tmp }
$c = Run ('git clone https://github.com/sitthichok-J/self-study-2026.git "' + $tmp + '"')
if ($c.Code -ne 0) { Finish 'FAILED: clone'; exit 1 }
$dest = Join-Path $tmp 'day10-observability-security'
$null = Run ('robocopy "' + $Root + '" "' + $dest + '" /E /XD .git /XF .env /NFL /NDL /NJH /NJS')
if (Test-Path (Join-Path $dest '.env')) { Finish 'FAILED: .env got copied, aborting before commit'; exit 1 }
$readme = Join-Path $tmp 'README.md'
$t = [IO.File]::ReadAllText($readme, [Text.Encoding]::UTF8)
$row = '| 10 | day10-observability-security | Observability + Security: Prometheus, Grafana, Jaeger, JWT, ตรวจเจ้าของข้อมูล, สแกนช่องโหว่และแก้แล้ว รันจริงผ่าน 57/57 ขั้น (UI/Kong ของ Day 9 ยังไม่ส่ง token) |'
$t2 = $t -replace '(?m)^Day 10 \(observability/security\) ยังไม่ได้ทำ\s*$', ''
$t2 = $t2.TrimEnd() + "`n"
if ($t2 -match '(?m)^\| 9 \|.*$') { $t2 = [regex]::Replace($t2, '(?m)^(\| 9 \|.*)$', ('$1' + "`n" + $row)) }
[IO.File]::WriteAllText($readme, $t2, (New-Object Text.UTF8Encoding($false)))
$null = Run ('git -C "' + $tmp + '" add -A')
$st = Run ('git -C "' + $tmp + '" status --short')
$msg = 'Add day10 observability + security (run-all 57/57, vulnerability fixes)'
$cm = Run ('git -C "' + $tmp + '" commit -m "' + $msg + '" -m "Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>" -m "Claude-Session: https://claude.ai/code/session_01CB41SWZ6wgjyln9epg3ara"')
if ($cm.Code -ne 0) { Finish 'FAILED: commit (or nothing to commit)'; exit 1 }
$p = Run ('git -C "' + $tmp + '" push origin HEAD')
if ($p.Code -ne 0) { Finish 'FAILED: push'; exit 1 }
Finish 'OK: pushed'
