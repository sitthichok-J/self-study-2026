# run-all.ps1 (Day 7): runs terraform on LocalStack inside Docker and writes notes\run-results.md
# Needs only Docker Desktop (running). terraform and the AWS CLI run as containers, nothing is installed.
# The AWS provider download is large the first time. Cleans up everything it creates.
$ErrorActionPreference = 'Continue'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$Out  = Join-Path $Root 'notes\run-results.md'
$TF   = 'docker run --rm --network day07_default -v "%CD%:/work" -w /work -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=ap-southeast-1 hashicorp/terraform:1.9'
$CLI  = 'docker run --rm --network day07_default -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=ap-southeast-1 amazon/aws-cli --endpoint-url http://localstack:4566'
$script:Results = @()
$script:Log = New-Object System.Text.StringBuilder
[void]$script:Log.AppendLine("# Day 7 run results"); [void]$script:Log.AppendLine("")
[void]$script:Log.AppendLine("Run at: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')  Machine: $env:COMPUTERNAME"); [void]$script:Log.AppendLine("")

function Save { $script:Log.ToString() | Out-File -FilePath $Out -Encoding utf8 }
function Sh([string]$cmd, [string]$dir = $Root) {
  Push-Location $dir
  $bat = Join-Path $env:TEMP 'day07-run-step.cmd'
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

# ---------------------------------------------------------------- 0 docker + localstack
$dockerOk = (Sh 'docker info').Code -eq 0
Step 'Docker running' $dockerOk ((Sh 'docker version --format "client {{.Client.Version}} / server {{.Server.Version}}"').Out)
if (-not $dockerOk) { Write-Host 'Start Docker Desktop and run again.'; Finish; exit 1 }

$up = Sh 'docker compose up -d'
$ok = $false
for ($i = 0; $i -lt 40; $i++) {
  try { $r = Invoke-WebRequest -UseBasicParsing -Uri 'http://localhost:4566/_localstack/health' -TimeoutSec 5; if ($r.StatusCode -eq 200) { $ok = $true; break } } catch {}
  Start-Sleep -Seconds 3
}
Step 'LocalStack up and healthy' ($up.Code -eq 0 -and $ok) ((Tail $up.Out 8))
if (-not $ok) { $null = Sh 'docker compose down -v'; Finish; exit 1 }

# ---------------------------------------------------------------- 1 infra (S3 + SQS + IAM)
$infra = Join-Path $Root 'infra'
$fmt = Sh "$TF fmt -check -recursive" $infra
Step 'terraform fmt -check (infra)' ($fmt.Code -eq 0) ("exit " + $fmt.Code + "`n" + (Tail $fmt.Out 15))
$init = Sh "$TF init -no-color" $infra
Step 'terraform init (infra)' ($init.Code -eq 0) (Tail $init.Out 15)
$val = Sh "$TF validate -no-color" $infra
Step 'terraform validate (infra)' ($val.Code -eq 0) (Tail $val.Out 20)
$plan = Sh "$TF plan -no-color -out=tfplan" $infra
Step 'terraform plan (infra)' ($plan.Code -eq 0 -and $plan.Out -match 'Plan: 10 to add') ((Tail $plan.Out 12))
$apply = Sh "$TF apply -no-color tfplan" $infra
Step 'terraform apply (infra)' ($apply.Code -eq 0) (Tail $apply.Out 12)
$again = Sh "$TF plan -no-color -detailed-exitcode" $infra
Step 'plan after apply shows no changes (exit 0)' ($again.Code -eq 0) ("exit " + $again.Code + "`n" + (Tail $again.Out 8))

if ($apply.Code -eq 0) {
  $bucket = (Sh "$TF output -raw bucket_name" $infra).Out.Trim()
  $queue  = (Sh "$TF output -raw queue_url" $infra).Out.Trim()
  $pab = Sh "$CLI s3api get-public-access-block --bucket $bucket"
  Step 'AWS CLI: bucket has public access fully blocked' (($pab.Out -match 'BlockPublicAcls": true') -and ($pab.Out -notmatch 'false')) $pab.Out
  $ver = Sh "$CLI s3api get-bucket-versioning --bucket $bucket"
  Step 'AWS CLI: bucket versioning enabled' ($ver.Out -match 'Enabled') $ver.Out
  $qa = Sh "$CLI sqs get-queue-attributes --queue-url $queue --attribute-names RedrivePolicy VisibilityTimeout"
  Step 'AWS CLI: queue has redrive policy to DLQ (maxReceiveCount 3)' (($qa.Out -match 'deadLetterTargetArn') -and ($qa.Out -match 'maxReceiveCount.{1,6}3')) $qa.Out
  $snd = Sh "$CLI sqs send-message --queue-url $queue --message-body hello-day07"
  $rcv = Sh "$CLI sqs receive-message --queue-url $queue --wait-time-seconds 2"
  Step 'SQS: send and receive a message' (($snd.Code -eq 0) -and ($rcv.Out -match 'hello-day07')) ((Tail $snd.Out 6) + "`n" + (Tail $rcv.Out 12))
  $pol = Sh "$CLI iam list-policies --scope Local"
  Step 'IAM: consumer and producer policies exist' (($pol.Out -match 'claims-consumer') -and ($pol.Out -match 'claims-producer')) (Tail $pol.Out 25)
}

# ---------------------------------------------------------------- 2 destroy and confirm empty
$dest = Sh "$TF destroy -no-color -auto-approve" $infra
Step 'terraform destroy (infra)' ($dest.Code -eq 0) (Tail $dest.Out 8)
$left = Sh "$CLI s3api list-buckets"
$leftQ = Sh "$CLI sqs list-queues"
Step 'nothing left in LocalStack after destroy' (($left.Out -notmatch 'selfstudy') -and ($leftQ.Out -notmatch 'selfstudy')) ($left.Out + "`n" + $leftQ.Out)

# ---------------------------------------------------------------- 3 vpc: plan only
$vpc = Join-Path $Root 'infra-vpc'
$vi = Sh "$TF init -no-color" $vpc
Step 'terraform init (infra-vpc)' ($vi.Code -eq 0) (Tail $vi.Out 10)
$vv = Sh "$TF validate -no-color" $vpc
Step 'terraform validate (infra-vpc)' ($vv.Code -eq 0) (Tail $vv.Out 20)
$vp = Sh "$TF plan -no-color" $vpc
Step 'terraform plan (infra-vpc, offline, nothing applied)' ($vp.Code -eq 0 -and $vp.Out -match 'Plan: 15 to add') (Tail $vp.Out 12)

$null = Sh 'docker compose down -v'
Finish
