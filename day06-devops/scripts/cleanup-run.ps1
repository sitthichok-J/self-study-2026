# stops a stuck run-all.ps1 and leftover claims containers
Get-CimInstance Win32_Process -Filter "Name='powershell.exe'" | Where-Object { $_.CommandLine -like '*run-all.ps1*' } | ForEach-Object { Stop-Process -Id $_.ProcessId -Force }
$ids = docker ps -aq --filter ancestor=claims-api:dev
if ($ids) { docker rm -f $ids }
docker compose -p claims down -v 2>$null
kind delete cluster --name claims 2>$null
Write-Host 'cleaned'
