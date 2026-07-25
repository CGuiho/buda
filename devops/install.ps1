param(
  [string]$Version = "latest",
  [string]$InstallDir = "$env:LOCALAPPDATA\GUIHO\bin"
)

$ErrorActionPreference = "Stop"
$Owner = "CGuiho"
$Repository = "buda"
$CliName = "buda"

switch ($env:PROCESSOR_ARCHITECTURE.ToUpperInvariant()) {
  "AMD64" { $Asset = "buda-windows-amd64.exe" }
  "ARM64" { $Asset = "buda-windows-arm64.exe" }
  default { throw "Unsupported Windows architecture: $env:PROCESSOR_ARCHITECTURE" }
}

if ($Version -eq "latest") {
  $Release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Owner/$Repository/releases/latest"
  $Tag = $Release.tag_name
  if ([string]::IsNullOrWhiteSpace($Tag)) { throw "Could not resolve the latest Buda release tag." }
} else {
  # Non-latest values are exact full release tags; Buda owns no implicit prefix.
  $Tag = $Version
}

$BaseUrl = "https://github.com/$Owner/$Repository/releases/download/$Tag"
$SkillAsset = "guiho-s-0002-buda.zip"
$TempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("buda-" + [guid]::NewGuid().ToString("N"))
$Destination = $null
$BackupPath = $null
$BinaryReplaced = $false
$InstallSucceeded = $false
New-Item -ItemType Directory -Path $TempDir | Out-Null

try {
  Write-Host "Initiating Buda installation sequence..."
  Write-Host "Target version: $Tag"
  Write-Host "Target asset: $Asset"
  Write-Host "Source URL: $BaseUrl/$Asset"

  $BinaryPath = Join-Path $TempDir $Asset
  $ChecksumsPath = Join-Path $TempDir "checksums.txt"
  $SkillPath = Join-Path $TempDir $SkillAsset
  Invoke-WebRequest -Uri "$BaseUrl/$Asset" -OutFile $BinaryPath
  Invoke-WebRequest -Uri "$BaseUrl/checksums.txt" -OutFile $ChecksumsPath
  Invoke-WebRequest -Uri "$BaseUrl/$SkillAsset" -OutFile $SkillPath

  foreach ($Verification in @(@($Asset, $BinaryPath), @($SkillAsset, $SkillPath))) {
    $VerificationName = $Verification[0]
    $VerificationPath = $Verification[1]
    $ChecksumLine = Get-Content -LiteralPath $ChecksumsPath | Where-Object { $_ -match "\s+$([regex]::Escape($VerificationName))$" } | Select-Object -First 1
    if (-not $ChecksumLine) { throw "Checksum entry missing for $VerificationName." }
    $ExpectedHash = ($ChecksumLine -split "\s+")[0].ToUpperInvariant()
    $ActualHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $VerificationPath).Hash.ToUpperInvariant()
    if ($ExpectedHash -ne $ActualHash) { throw "Checksum verification failed for $VerificationName." }
  }
  Write-Host "[OK] SHA-256 verification complete for binary and skill archive."

  New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
  $Destination = Join-Path $InstallDir "buda.exe"

  $ExpandedSkill = Join-Path $TempDir "skill"
  Expand-Archive -LiteralPath $SkillPath -DestinationPath $ExpandedSkill
  $SourceSkill = Join-Path $ExpandedSkill "guiho-s-0002-buda"
  if (-not (Test-Path -LiteralPath (Join-Path $SourceSkill "SKILL.md"))) {
    throw "Skill archive does not contain guiho-s-0002-buda/SKILL.md."
  }

  $Qmd = Get-Command qmd -ErrorAction SilentlyContinue
  if (-not $Qmd) {
    throw "qmd is required but not installed. Install @tobilu/qmd, then run: buda doctor --wiki <path>"
  }
  $QmdVersionOutput = (& qmd --version | Out-String).Trim()
  if ($QmdVersionOutput -notmatch '(?i)(?:qmd\s+)?v?(\d+)\.(\d+)\.(\d+)') {
    throw "Could not parse qmd version: $QmdVersionOutput"
  }
  $QmdMajor = [int]$Matches[1]
  $QmdMinor = [int]$Matches[2]
  if ($QmdMajor -ne 2 -or $QmdMinor -lt 5) {
    throw "Unsupported qmd version '$QmdVersionOutput'; Buda requires >=2.5.0,<3.0.0."
  }

  $StagedBinary = Join-Path $InstallDir (".buda-new-" + [guid]::NewGuid().ToString("N") + ".exe")
  Copy-Item -LiteralPath $BinaryPath -Destination $StagedBinary
  if (Test-Path -LiteralPath $Destination) {
    $BackupPath = Join-Path $InstallDir (".buda-backup-" + [guid]::NewGuid().ToString("N") + ".exe")
    Move-Item -LiteralPath $Destination -Destination $BackupPath
  }
  $BinaryReplaced = $true
  Move-Item -LiteralPath $StagedBinary -Destination $Destination
  Write-Host "[OK] Installed binary: $Destination"

  & $Destination agent skill update
  Write-Host "[OK] Installed both global Buda skill destinations transactionally from embedded resources."

  $InstalledVersion = (& $Destination --version | Out-String).Trim()
  $ExpectedVersion = $Tag.TrimStart("v")
  if ($InstalledVersion -ne "buda v$ExpectedVersion") { throw "Installed version '$InstalledVersion' does not match requested tag '$Tag'." }
  Write-Host $InstalledVersion
  $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
  $PathEntries = @($UserPath -split ";" | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
  if ($PathEntries -notcontains $InstallDir) {
    [Environment]::SetEnvironmentVariable("Path", (($PathEntries + $InstallDir) -join ";"), "User")
    Write-Host "[OK] Added installation directory to the user PATH."
  }
  Write-Host $QmdVersionOutput
  Write-Host "[OK] Buda installation complete. Repository instructions are installed only for an explicit --wiki path."
  $InstallSucceeded = $true
} catch {
  if ($BinaryReplaced -and -not $InstallSucceeded) {
    if ($Destination -and (Test-Path -LiteralPath $Destination)) { Remove-Item -Force -LiteralPath $Destination }
    if ($BackupPath -and (Test-Path -LiteralPath $BackupPath)) { Move-Item -LiteralPath $BackupPath -Destination $Destination }
  }
  throw
} finally {
  if ($InstallSucceeded -and $BackupPath -and (Test-Path -LiteralPath $BackupPath)) { Remove-Item -Force -LiteralPath $BackupPath }
  if (Test-Path -LiteralPath $TempDir) { Remove-Item -Recurse -Force -LiteralPath $TempDir }
}
