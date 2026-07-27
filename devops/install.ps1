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
  if (-not [string]::IsNullOrWhiteSpace($env:BUDA_RELEASE_ASSET_DIR)) {
    throw "BUDA_RELEASE_ASSET_DIR requires an exact canonical tag such as buda/v0.0.2."
  }
  $Release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Owner/$Repository/releases/latest"
  $Tag = $Release.tag_name
  if ([string]::IsNullOrWhiteSpace($Tag)) { throw "Could not resolve the latest Buda release tag." }
} else {
  # Non-latest values are exact full release tags; Buda owns no implicit prefix.
  $Tag = $Version
}

if ($Tag -notmatch '^buda/v(?<Version>(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?)$') {
  throw "Invalid Buda release tag '$Tag'; expected buda/v<semver>."
}
$ExpectedVersion = $Matches.Version

$BaseUrl = "https://github.com/$Owner/$Repository/releases/download/$Tag"
$SkillAsset = "guiho-s-0002-buda.zip"
$TempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("buda-" + [guid]::NewGuid().ToString("N"))
$Destination = $null
$BackupPath = $null
$BinaryReplaced = $false
$InstallSucceeded = $false
New-Item -ItemType Directory -Path $TempDir | Out-Null

function Copy-ReleaseAsset {
  param(
    [Parameter(Mandatory = $true)][string]$Name,
    [Parameter(Mandatory = $true)][string]$Destination
  )

  if (-not [string]::IsNullOrWhiteSpace($env:BUDA_RELEASE_ASSET_DIR)) {
    $Source = Join-Path $env:BUDA_RELEASE_ASSET_DIR $Name
    if (-not (Test-Path -LiteralPath $Source -PathType Leaf)) {
      throw "Local release asset not found: $Source"
    }
    Copy-Item -LiteralPath $Source -Destination $Destination
    return
  }
  Invoke-WebRequest -Uri "$BaseUrl/$Name" -OutFile $Destination
}

try {
  Write-Host "Initiating Buda installation sequence..."
  Write-Host "Target version: $Tag"
  Write-Host "Target asset: $Asset"
  if (-not [string]::IsNullOrWhiteSpace($env:BUDA_RELEASE_ASSET_DIR)) {
    Write-Host "Source directory: $env:BUDA_RELEASE_ASSET_DIR"
  } else {
    Write-Host "Source URL: $BaseUrl/$Asset"
  }

  $BinaryPath = Join-Path $TempDir $Asset
  $ChecksumsPath = Join-Path $TempDir "checksums.txt"
  $SkillPath = Join-Path $TempDir $SkillAsset
  Copy-ReleaseAsset -Name $Asset -Destination $BinaryPath
  Copy-ReleaseAsset -Name "checksums.txt" -Destination $ChecksumsPath
  Copy-ReleaseAsset -Name $SkillAsset -Destination $SkillPath

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

  # npm installs both qmd.cmd and qmd.ps1 on Windows. Prefer the application
  # launcher so the installer works under PowerShell's default script policy.
  $Qmd = Get-Command qmd.cmd -CommandType Application -ErrorAction SilentlyContinue
  if (-not $Qmd) {
    $Qmd = Get-Command qmd -CommandType Application -ErrorAction SilentlyContinue
  }
  if (-not $Qmd) {
    throw "qmd is required but not installed. Install @tobilu/qmd@2.5.3, then run: buda doctor --wiki <path>"
  }
  $QmdVersionOutput = (& $Qmd.Source --version | Out-String).Trim()
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
  if ($InstalledVersion -ne "buda v$ExpectedVersion") { throw "Installed version '$InstalledVersion' does not match requested tag '$Tag'." }
  Write-Host $InstalledVersion
  $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
  $PathEntries = @($UserPath -split ";" | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
  if ($PathEntries -notcontains $InstallDir) {
    try {
      [Environment]::SetEnvironmentVariable("Path", (($PathEntries + $InstallDir) -join ";"), "User")
      Write-Host "[OK] Added installation directory to the user PATH."
    } catch {
      Write-Warning "Buda is installed, but the user PATH could not be updated. Add '$InstallDir' to PATH manually."
    }
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
