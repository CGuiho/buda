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

  $ChecksumLine = Get-Content -LiteralPath $ChecksumsPath | Where-Object { $_ -match "\s+$([regex]::Escape($Asset))$" } | Select-Object -First 1
  if (-not $ChecksumLine) { throw "Checksum entry missing for $Asset." }
  $ExpectedHash = ($ChecksumLine -split "\s+")[0].ToUpperInvariant()
  $ActualHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $BinaryPath).Hash.ToUpperInvariant()
  if ($ExpectedHash -ne $ActualHash) { throw "Checksum verification failed for $Asset." }
  Write-Host "[OK] SHA-256 verification complete."

  New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
  $Destination = Join-Path $InstallDir "buda.exe"
  Copy-Item -Force -LiteralPath $BinaryPath -Destination $Destination
  Write-Host "[OK] Installed binary: $Destination"

  $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
  $PathEntries = @($UserPath -split ";" | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
  if ($PathEntries -notcontains $InstallDir) {
    [Environment]::SetEnvironmentVariable("Path", (($PathEntries + $InstallDir) -join ";"), "User")
    Write-Host "[OK] Added installation directory to the user PATH."
  }

  $ExpandedSkill = Join-Path $TempDir "skill"
  Expand-Archive -LiteralPath $SkillPath -DestinationPath $ExpandedSkill
  $SourceSkill = Join-Path $ExpandedSkill "guiho-s-0002-buda"
  if (-not (Test-Path -LiteralPath (Join-Path $SourceSkill "SKILL.md"))) {
    throw "Skill archive does not contain guiho-s-0002-buda/SKILL.md."
  }
  foreach ($Root in @("$HOME\.agents\skills", "$HOME\.claude\skills")) {
    New-Item -ItemType Directory -Force -Path $Root | Out-Null
    $Target = Join-Path $Root "guiho-s-0002-buda"
    if (Test-Path -LiteralPath $Target) { Remove-Item -Recurse -Force -LiteralPath $Target }
    Copy-Item -Recurse -LiteralPath $SourceSkill -Destination $Target
    Write-Host "[OK] Installed global Buda skill: $Target"
  }

  & $Destination --version
  $Qmd = Get-Command qmd -ErrorAction SilentlyContinue
  if (-not $Qmd) {
    throw "qmd is required but not installed. Install @tobilu/qmd, then run: buda doctor --wiki <path>"
  }
  & qmd --version
  Write-Host "[OK] Buda installation complete. Repository instructions are installed only for an explicit --wiki path."
} finally {
  if (Test-Path -LiteralPath $TempDir) { Remove-Item -Recurse -Force -LiteralPath $TempDir }
}
