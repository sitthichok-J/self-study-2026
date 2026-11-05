# security-scan.ps1 (Day 12): govulncheck + gitleaks + Trivy ผ่าน Docker เขียนผลลง notes\security-scan.md
# ต้องมี Docker Desktop และเน็ต (ดึง image + vulnerability DB) ไม่ได้แก้โค้ดให้ แค่รายงาน
param([switch]$NoPause)
$ErrorActionPreference = 'Continue'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$Out  = Join-Path $Root 'notes\security-scan.md'
$sb = New-Object System.Text.StringBuilder
[void]$sb.AppendLine('# Day 12 security scan'); [void]$sb.AppendLine('')
[void]$sb.AppendLine("Run at: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"); [void]$sb.AppendLine('')
function Sh([string]$cmd) {
  Push-Location $Root
  $bat = Join-Path $env:TEMP 'day13-scan-step.cmd'
  Set-Content -Path $bat -Value ("@echo off`r`n" + $cmd + " 2>&1") -Encoding ASCII
  $o = & cmd /c $bat | Out-String; $c = $LASTEXITCODE
  Pop-Location
  return @{ Code = $c; Out = $o.TrimEnd() }
}
function Section([string]$title, $r, [string]$meaning) {
  Write-Host ("{0}: exit {1}" -f $title, $r.Code)
  [void]$sb.AppendLine("## $title (exit code $($r.Code))"); [void]$sb.AppendLine('')
  [void]$sb.AppendLine($meaning); [void]$sb.AppendLine('')
  $l = $r.Out -split "`r?`n"; if ($l.Count -gt 80) { $l = @('... (' + ($l.Count - 80) + ' earlier lines omitted)') + ($l | Select-Object -Last 80) }
  $body = [regex]::Replace(($l -join "`n"), '-----(BEGIN|END) [A-Z ]*PRIVATE KEY-----', '[private-key-marker]')  # ผลสแกนไม่ต้องมีหัวกุญแจติดมา (ไม่งั้นเทสต์หา secret ใน repo จะเจอไฟล์ผล)
  [void]$sb.AppendLine('```'); [void]$sb.AppendLine($body); [void]$sb.AppendLine('```'); [void]$sb.AppendLine('')
  $sb.ToString() | Out-File -FilePath $Out -Encoding utf8
}
if ((Sh 'docker info').Code -ne 0) { Write-Host 'Start Docker Desktop and run again.'; exit 1 }

$g = Sh 'docker run --rm -v "%CD%:/src" -v day13-go-mod:/go/pkg/mod -v day13-go-build:/root/.cache/go-build -w /src golang:1.26 sh -c "go install golang.org/x/vuln/cmd/govulncheck@latest && govulncheck ./..."'
Section 'govulncheck ./...' $g 'exit 0 = ไม่พบช่องโหว่ที่โค้ดเรียกใช้จริง, exit 3 = พบ (ดูรายการ)'

$k = Sh 'docker run --rm -v "%CD%:/repo" zricethezav/gitleaks:latest detect --source=/repo --no-git --redact -v'
Section 'gitleaks (สแกนไฟล์ในโฟลเดอร์ ซ่อนค่า)' $k 'exit 0 = ไม่พบ secret, exit 1 = พบ'

$b = Sh 'docker build --build-arg CMD=api -t day13-api-scan .'
$sv = Sh 'docker save -o "%TEMP%\day13-api-scan.tar" day13-api-scan'
$t = Sh 'docker run --rm -v "%TEMP%:/tmp/in" -v day13-trivy-cache:/root/.cache aquasec/trivy:latest image --input /tmp/in/day13-api-scan.tar --severity HIGH,CRITICAL --scanners vuln'
Section 'Trivy: image api (HIGH,CRITICAL)' $t "exit 0 = สแกนจบ (ดูตารางว่ามีรายการหรือไม่) build: exit $($b.Code), save: exit $($sv.Code)"
$t2 = Sh 'docker run --rm -v "%CD%:/src" -v day13-trivy-cache:/root/.cache aquasec/trivy:latest fs /src --severity HIGH,CRITICAL --scanners vuln,secret,misconfig'
Section 'Trivy: fs (vuln + secret + misconfig, รวม Dockerfile และ compose)' $t2 'ดูตารางรายการท้ายผล'
Write-Host "Results written to $Out"
if (-not $NoPause) { Read-Host 'Press Enter to close' }
