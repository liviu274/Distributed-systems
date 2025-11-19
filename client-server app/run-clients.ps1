<#
run-clients.ps1

Loads run-config.json (next to this script), builds the client executable,
then starts the configured number of background jobs. Each job runs the
client exe passing `-name` and `-max` flags.
#>

# Determine script directory and config path
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$configPath = Join-Path $scriptDir 'run-config.json'

if (-not (Test-Path $configPath)) {
  Write-Error "Config file not found: $configPath"
  exit 1
}

# Parse JSON config into a PowerShell object
$cfg = Get-Content -Raw $configPath | ConvertFrom-Json

# Ensure numeric values are usable
$clientsCount = [int]$cfg.ClientsCount
$clientExe = $cfg.ClientExe
$maxElements = [int]$cfg.MaxElements

# Build the client exe once for faster startup
Push-Location -Path $scriptDir
try {
  Write-Output "Building client executable: $clientExe"
  & go build -o $clientExe ex2-client.go
  if ($LASTEXITCODE -ne 0) {
    Write-Error "go build failed (exit code $LASTEXITCODE)"
    exit 1
  }
} finally {
  Pop-Location
}

# Compute absolute path for the exe
$exeFullPath = Join-Path $scriptDir $clientExe

Write-Output "Starting $clientsCount client jobs"

1..$clientsCount | ForEach-Object {
  $i = $_
  $name = "Client-$i"
  Start-Job -Name ("client" + $i) -ScriptBlock {
    param($exePath, $name, $max)
    # Ensure working directory is where the exe is located
    $exeDir = Split-Path -Parent $exePath
    if ($exeDir -ne '') { Set-Location -Path $exeDir }
    # Invoke the client exe with flags
    & $exePath -name $name -max $max
  } -ArgumentList $exeFullPath, $name, $maxElements
}

# Wait for all jobs, collect and print outputs, then remove jobs
Get-Job | Wait-Job
Get-Job | Receive-Job
Get-Job | Remove-Job