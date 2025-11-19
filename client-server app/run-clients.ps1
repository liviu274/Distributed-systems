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

## Build and run multiple client sources (ex2-client.go, ex5-client.go)
$clientSources = @("ex2-client.go", "ex5-client.go", "ex7-client.go", "ex9-client.go", "ex14-client.go")

foreach ($src in $clientSources) {
  $srcPath = Join-Path $scriptDir $src
  if (-not (Test-Path $srcPath)) {
    Write-Warning "Client source not found, skipping: $src"
    continue
  }

  $exeName = [IO.Path]::ChangeExtension($src, '.exe')

  Push-Location -Path $scriptDir
  try {
    Write-Output "Building $src -> $exeName"
    & go build -o $exeName $src
    if ($LASTEXITCODE -ne 0) {
      Write-Error "go build failed for $src (exit code $LASTEXITCODE)"
      exit 1
    }
  } finally {
    Pop-Location
  }

  $exeFullPath = Join-Path $scriptDir $exeName
  Write-Output "Starting $clientsCount jobs for $exeName"

  1..$clientsCount | ForEach-Object {
    $i = $_
    $name = "${exeName.Replace('.exe','')}-${i}"
    Start-Job -Name ("client-" + $exeName + "-" + $i) -ScriptBlock {
      param($exePath, $name, $max)
      $exeDir = Split-Path -Parent $exePath
      if ($exeDir -ne '') { Set-Location -Path $exeDir }
      & $exePath -name $name -max $max
    } -ArgumentList $exeFullPath, $name, $maxElements
  }

}

Write-Output "All jobs started. Waiting for completion..."

# Wait for all jobs, collect and print outputs, then remove jobs
Get-Job | Wait-Job
Get-Job | Receive-Job
Get-Job | Remove-Job