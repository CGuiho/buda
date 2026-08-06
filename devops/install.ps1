param(
  [string]$Version = "latest",
  [string]$InstallDir = "$env:LOCALAPPDATA\GUIHO\bin"
)

# Buda canonical-tag installer (PowerShell 5.1 compatible).
#
# Environment variables:
#   BUDA_RELEASE_ASSET_DIR  Load release assets from a local directory instead of
#                           GitHub; requires an exact canonical tag argument.
#   BUDA_SKILL_DIRS         Semicolon-separated list of extra skill destination
#                           directories. Each is created if missing and the
#                           embedded skill directory is copied into it (the skill
#                           id subdirectory is appended automatically). Defaults
#                           to empty; the built-in ~/.agents/skills and
#                           ~/.claude/skills destinations are always installed by
#                           `buda agent skill update`.
#   HERMES_SKILLS_DIR       Hermes agent skills directory. When this directory
#                           exists (default: ~/.hermes/skills) the embedded skill
#                           is additionally registered there by the installer so
#                           Hermes agents can use Buda without a manual copy.

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
  # Probe common off-PATH locations (agent-managed environments) before failing.
  if (-not $Qmd) {
    $ProbeDirs = @()
    $NpmRoot = $null
    try { $NpmRoot = (npm root -g 2>$null) } catch { }
    if ($NpmRoot) { $ProbeDirs += (Join-Path $NpmRoot "bin") }
    $ProbeDirs += @(
      (Join-Path $env:USERPROFILE ".hermes\node\bin"),
      (Join-Path $env:USERPROFILE ".local\bin"),
      (Join-Path $env:USERPROFILE ".npm-global\bin"),
      (Join-Path $env:APPDATA "npm")
    )
    foreach ($ProbeDir in $ProbeDirs) {
      if ([string]::IsNullOrWhiteSpace($ProbeDir)) { continue }
      $ProbePath = Join-Path $ProbeDir "qmd.cmd"
      if (Test-Path -LiteralPath $ProbePath -PathType Leaf) {
        $Qmd = Get-Item -LiteralPath $ProbePath
        Write-Warning "qmd was not on PATH but was found at $ProbePath; add '$ProbeDir' to PATH for reliable operation."
        break
      }
      $ProbePath = Join-Path $ProbeDir "qmd"
      if (Test-Path -LiteralPath $ProbePath -PathType Leaf) {
        $Qmd = Get-Item -LiteralPath $ProbePath
        Write-Warning "qmd was not on PATH but was found at $ProbePath; add '$ProbeDir' to PATH for reliable operation."
        break
      }
    }
  }
  if (-not $Qmd) {
    throw "qmd is required but not installed. Install @tobilu/qmd@2.5.3, then run: buda doctor --wiki <path>"
  }
  $QmdSource = $Qmd.Source
  if (-not $QmdSource) { $QmdSource = $Qmd.FullName }
  $QmdVersionOutput = (& $QmdSource --version | Out-String).Trim()
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
  # The binary is verified and installed; mark the install successful before
  # the optional skill-destination registrations below so an optional
  # registration failure can never roll back a valid binary replacement.
  $InstallSucceeded = $true

  # Register the embedded skill into optional destinations: BUDA_SKILL_DIRS
  # (semicolon-separated) and the Hermes skills directory when it exists. The
  # built-in ~/.agents/skills and ~/.claude/skills destinations are already
  # handled by `buda agent skill update` above; these are additive and
  # non-fatal: a failure only warns and leaves the installation complete.
  function Register-SkillDir([string]$Target) {
    if ([string]::IsNullOrWhiteSpace($Target)) { return }
    $Dest = Join-Path $Target "guiho-s-0002-buda"
    try {
      New-Item -ItemType Directory -Force -Path $Target | Out-Null
      if (Test-Path -LiteralPath $Dest) { Remove-Item -Recurse -Force -LiteralPath $Dest }
      Copy-Item -Recurse -LiteralPath $SourceSkill -Destination $Dest
      Write-Host "[OK] Registered Buda skill: $Dest"
    } catch {
      Write-Warning "Could not register Buda skill in '$Target' (skipped): $($_.Exception.Message)"
    }
  }

  if (-not [string]::IsNullOrWhiteSpace($env:BUDA_SKILL_DIRS)) {
    foreach ($SkillDir in ($env:BUDA_SKILL_DIRS -split ";")) {
      Register-SkillDir $SkillDir
    }
  }

  $HermesSkillsDir = $env:HERMES_SKILLS_DIR
  if ([string]::IsNullOrWhiteSpace($HermesSkillsDir)) {
    $HermesSkillsDir = Join-Path $env:USERPROFILE ".hermes\skills"
  }
  if (Test-Path -LiteralPath $HermesSkillsDir -PathType Container) {
    Register-SkillDir $HermesSkillsDir
  }

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
  # Verify the binary is callable in the current session; warn (not fail) when
  # the install directory is not yet on the session PATH.
  if (-not (Get-Command $CliName -ErrorAction SilentlyContinue)) {
    Write-Warning "buda was installed to $InstallDir which is not on the session PATH. Open a new shell or run: `$env:Path = '$InstallDir;' + `$env:Path. Then run: buda doctor --wiki <path>"
  }
  Write-Host $QmdVersionOutput
  Write-Host "[OK] Buda installation complete. Repository instructions are installed only for an explicit --wiki path."
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
