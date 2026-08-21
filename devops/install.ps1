[CmdletBinding()]
param(
  [string]$Version,
  [string]$Channel,
  [Parameter(Mandatory=$true)][string]$Wiki,
  [string]$WikiId
)
$ErrorActionPreference = 'Stop'
if (-not [string]::IsNullOrWhiteSpace($Version) -and -not [string]::IsNullOrWhiteSpace($Channel)) { throw '--Version and --Channel are mutually exclusive.' }
if ($Version -match '^buda/') { $Version = $Version.Substring(5) }
if ($Version -match '^v') { $Version = $Version.Substring(1) }
if ($Version -and $Version -notmatch '^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$') { throw "Invalid -Version '$Version'." }
if (-not $Version) { $Version = $env:BUDA_RELEASE_VERSION }
$arch = $env:PROCESSOR_ARCHITECTURE.ToUpperInvariant()
if ($arch -eq 'AMD64') { $binary = 'buda-windows-amd64.exe'; $launcher = 'buda-launcher-windows-amd64.exe' }
elseif ($arch -eq 'ARM64') { $binary = 'buda-windows-arm64.exe'; $launcher = 'buda-launcher-windows-arm64.exe' }
else { throw "Unsupported Windows architecture: $arch" }
if (-not $Version -and -not $env:BUDA_RELEASE_ASSET_DIR) {
  # Windows PowerShell 5.1 emits a top-level JSON array from Invoke-RestMethod
  # as a single non-enumerated pipeline object, so wrapping the call directly
  # in @(...) would nest the releases inside one element and break every filter
  # below. Capture the response first, then flatten it explicitly.
  $wanted = if ($Channel) { $Channel } else { 'stable' }
  $all = @(); $page = 1
  do {
    $releasePage = Invoke-RestMethod "https://api.github.com/repos/CGuiho/buda/releases?per_page=100&page=$page"
    $batch = @($releasePage)
    $all += $batch; $page++
  } while ($batch.Count -eq 100)
  $candidates = @($all | Where-Object { -not $_.draft -and $_.tag_name -match '^buda/v' } | ForEach-Object {
    $candidate = $_.tag_name.Substring(6); $candidateChannel = if ($candidate.Contains('-')) { $candidate.Split('-')[1].Split('.')[0] } else { 'stable' }
    $requiredNames = @($binary, $launcher, 'checksums.txt', 'artifacts.json', 'guiho-s-0002-buda.zip', 'guiho-i-buda.md', 'guiho-p-buda.md', 'buda.schema.json', 'buda.global.schema.json', 'buda.example.yaml', 'buda.global.example.yaml',
      'buda-linux-amd64', 'buda-linux-arm64', 'buda-linux-armv7', 'buda-linux-armv6', 'buda-darwin-amd64', 'buda-darwin-arm64', 'buda-windows-amd64.exe', 'buda-windows-arm64.exe',
      'buda-launcher-linux-amd64', 'buda-launcher-linux-arm64', 'buda-launcher-linux-armv7', 'buda-launcher-linux-armv6', 'buda-launcher-darwin-amd64', 'buda-launcher-darwin-arm64', 'buda-launcher-windows-amd64.exe', 'buda-launcher-windows-arm64.exe')
    $assetNames = @($_.assets | ForEach-Object { [string]$_.name })
    $uniqueAssetNames = @($assetNames | Select-Object -Unique)
    if ($assetNames.Count -ne $uniqueAssetNames.Count) { return }
    if ($candidateChannel -eq $wanted -and $candidate -match '^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$' -and @($requiredNames | Where-Object { $assetNames -notcontains $_ }).Count -eq 0) { [pscustomobject]@{ Release=$_; Version=$candidate } }
  } | Sort-Object Version -Descending)
  if ($candidates.Count -eq 0) { throw "No release found for channel $wanted." }
  function Compare-SemVer([string]$left, [string]$right) {
    $leftParts = $left.Split('+')[0].Split('-',2); $rightParts = $right.Split('+')[0].Split('-',2)
    $lCore = $leftParts[0].Split('.'); $rCore = $rightParts[0].Split('.')
    for ($i=0; $i -lt 3; $i++) { $comparison = [int]$lCore[$i] - [int]$rCore[$i]; if ($comparison -ne 0) { return [Math]::Sign($comparison) } }
    if ($leftParts.Count -eq 1 -and $rightParts.Count -eq 1) { return 0 }
    if ($leftParts.Count -eq 1) { return 1 }
    if ($rightParts.Count -eq 1) { return -1 }
    $lIds = $leftParts[1].Split('.'); $rIds = $rightParts[1].Split('.')
    for ($i=0; $i -lt [Math]::Max($lIds.Count,$rIds.Count); $i++) {
      if ($i -ge $lIds.Count) { return -1 }; if ($i -ge $rIds.Count) { return 1 }
      $ln = $lIds[$i] -match '^[0-9]+$'; $rn = $rIds[$i] -match '^[0-9]+$'
      if ($ln -and $rn) { $comparison = [int64]$lIds[$i] - [int64]$rIds[$i] } elseif ($ln) { return -1 } elseif ($rn) { return 1 } else { $comparison = [string]::CompareOrdinal($lIds[$i],$rIds[$i]) }
      if ($comparison -ne 0) { return [Math]::Sign($comparison) }
    }
    return 0
  }
  $best = $candidates[0]
  foreach ($candidate in $candidates | Select-Object -Skip 1) { if ((Compare-SemVer $candidate.Version $best.Version) -gt 0) { $best = $candidate } }
  $Version = $best.Version
}
if (-not $Version) { throw 'Version selection did not resolve a stable release.' }
$tag = "buda/v$Version"
$assetDir = $env:BUDA_RELEASE_ASSET_DIR
$userHome = if ($env:USERPROFILE) { $env:USERPROFILE } elseif ($env:HOME) { $env:HOME } else { [Environment]::GetFolderPath('UserProfile') }
$guihoHome = Join-Path $userHome '.guiho'
$cliHome = Join-Path $guihoHome 'buda'
$globalConfigName = 'buda.global.yaml'
$binDir = Join-Path $guihoHome 'bin'
$tempRoot = Join-Path $guihoHome '.temp'
New-Item -ItemType Directory -Force -Path $tempRoot | Out-Null
$stage = Join-Path $tempRoot ('buda-install-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $stage | Out-Null
$mutated = $false
$committed = $false
$hadCurrent = $false; $hadManifest = $false; $hadLauncher = $false; $hadVersion = $false
$hadLegacy = $false
$versionDir = Join-Path $cliHome (Join-Path 'versions' $Version)
$backupDir = Join-Path $stage 'backup'
# The historical 0.1.x direct-binary installation wrote the payload to
# %LOCALAPPDATA%\GUIHO\bin\buda.exe. Migration removes it only from that exact
# historical path, only after the new launcher transaction has been verified,
# and never in place.
$legacyPath = if ($env:LOCALAPPDATA) { Join-Path (Join-Path $env:LOCALAPPDATA 'GUIHO\bin') 'buda.exe' } else { '' }

function Get-Asset([string]$name) {
  $destination = Join-Path $stage $name
  if ($assetDir) { Copy-Item -LiteralPath (Join-Path $assetDir $name) -Destination $destination }
  else { Invoke-WebRequest -Uri "https://github.com/CGuiho/buda/releases/download/$tag/$name" -OutFile $destination }
  return $destination
}
function Write-Atomic([string]$path, [string]$content) {
  $temporary = "$path.new"
  $encoding = New-Object System.Text.UTF8Encoding($false)
  [System.IO.File]::WriteAllText($temporary, $content, $encoding)
  Move-Item -Force -LiteralPath $temporary -Destination $path
}
function Get-Sha256([string]$path) {
  $stream = [System.IO.File]::OpenRead($path)
  try {
    $hash = [System.Security.Cryptography.SHA256]::Create()
    try { return ([System.BitConverter]::ToString($hash.ComputeHash($stream)) -replace '-', '').ToUpperInvariant() }
    finally { $hash.Dispose() }
  } finally { $stream.Dispose() }
}
function Verify-Checksum([string]$name, $checksumLines) {
  $checksumMatches = @($checksumLines | Where-Object { $_ -match "^\s*([0-9a-fA-F]{64})\s+\*?$([regex]::Escape($name))\s*$" })
  if ($checksumMatches.Count -ne 1) { throw "Missing or duplicate checksum for $name." }
  $expected = $checksumMatches[0].Substring(0,64).ToUpperInvariant()
  $actual = Get-Sha256 (Join-Path $stage $name)
  if ($expected -ne $actual) { throw "Checksum mismatch for $name." }
}
function Restore-Previous {
  if (-not $mutated) { return }
  if (Test-Path -LiteralPath $versionDir) { Remove-Item -Recurse -Force -LiteralPath $versionDir -ErrorAction SilentlyContinue }
  if ($hadVersion) { Move-Item -Force -LiteralPath (Join-Path $backupDir 'version') -Destination $versionDir }
  $currentPath = Join-Path $cliHome 'current.json'; $manifestPath = Join-Path $cliHome 'installed-artifacts.json'; $launcherPath = Join-Path $binDir 'buda.exe'
  if ($hadCurrent) { Copy-Item -Force -LiteralPath (Join-Path $backupDir 'current.json') -Destination $currentPath } elseif (Test-Path -LiteralPath $currentPath) { Remove-Item -Force -LiteralPath $currentPath }
  if ($hadManifest) { Copy-Item -Force -LiteralPath (Join-Path $backupDir 'installed-artifacts.json') -Destination $manifestPath } elseif (Test-Path -LiteralPath $manifestPath) { Remove-Item -Force -LiteralPath $manifestPath }
  if ($hadLauncher) { Copy-Item -Force -LiteralPath (Join-Path $backupDir 'buda.exe') -Destination $launcherPath } elseif (Test-Path -LiteralPath $launcherPath) { Remove-Item -Force -LiteralPath $launcherPath }
  if ($hadLegacy) { Copy-Item -Force -LiteralPath (Join-Path $backupDir 'legacy-buda.exe') -Destination $legacyPath }
}

try {
  Write-Host "Resolved Buda $Version for windows/$arch"
  Write-Host "CLI home: $cliHome"
  [void](Get-Asset 'artifacts.json'); [void](Get-Asset 'checksums.txt')
  $manifest = Get-Content -Raw -LiteralPath (Join-Path $stage 'artifacts.json') | ConvertFrom-Json
  if ($manifest.schema -ne 1 -or $manifest.cli -ne 'buda' -or $manifest.version -ne $Version) { throw 'Release artifacts.json is not a valid Buda manifest for the selected version.' }
  $paths = @($manifest.artifacts | ForEach-Object { [string]$_.path })
  if (@($paths | Sort-Object -Unique).Count -ne $paths.Count) { throw 'Manifest contains duplicate asset paths.' }
  $required = @($binary,$launcher,'guiho-s-0002-buda.zip','guiho-i-buda.md','guiho-p-buda.md','buda.schema.json','buda.global.schema.json','buda.example.yaml','buda.global.example.yaml','artifacts.json')
  foreach ($name in $paths) {
    if ([string]::IsNullOrWhiteSpace($name) -or [IO.Path]::GetFileName($name) -ne $name -or $name.Contains('/') -or $name.Contains('\') -or $name -eq '..') { throw "Unsafe manifest asset path '$name'." }
    if (-not (Test-Path -LiteralPath (Join-Path $stage $name))) { [void](Get-Asset $name) }
  }
  foreach ($name in $required) { if ($paths -notcontains $name) { throw "Manifest does not declare required asset $name." } }
  $checksumLines = @(Get-Content -LiteralPath (Join-Path $stage 'checksums.txt'))
  foreach ($name in $paths) { Verify-Checksum $name $checksumLines }
  foreach ($line in $checksumLines) {
    if ($line -notmatch '^\s*[0-9a-fA-F]{64}\s+\*?([^\s]+)\s*$') { throw "Malformed checksum entry '$line'." }
    if ($paths -notcontains $matches[1]) { throw "Checksum names undeclared asset $($matches[1])." }
  }
  $observed = (& (Join-Path $stage $binary) --version | Out-String).Trim()
  if ($observed -ne $Version) { throw "Candidate version '$observed' does not match '$Version'." }
  $selfTest = (& (Join-Path $stage $binary) __self-test | Out-String).Trim()
  if ($selfTest -ne 'ok') { throw 'Candidate self-test failed.' }

  New-Item -ItemType Directory -Force -Path $backupDir,(Join-Path $cliHome 'versions'),$binDir | Out-Null
  $currentPath = Join-Path $cliHome 'current.json'; $manifestPath = Join-Path $cliHome 'installed-artifacts.json'; $launcherPath = Join-Path $binDir 'buda.exe'
  if (Test-Path -LiteralPath $currentPath) { Copy-Item -Force -LiteralPath $currentPath -Destination (Join-Path $backupDir 'current.json'); $hadCurrent = $true }
  if (Test-Path -LiteralPath $manifestPath) { Copy-Item -Force -LiteralPath $manifestPath -Destination (Join-Path $backupDir 'installed-artifacts.json'); $hadManifest = $true }
  if (Test-Path -LiteralPath $launcherPath) { Copy-Item -Force -LiteralPath $launcherPath -Destination (Join-Path $backupDir 'buda.exe'); $hadLauncher = $true }
  if ($legacyPath -and (Test-Path -LiteralPath $legacyPath -PathType Leaf)) {
    $legacyObserved = (& $legacyPath --version | Out-String).Trim()
    if ($legacyObserved -match '^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$') {
      Copy-Item -Force -LiteralPath $legacyPath -Destination (Join-Path $backupDir 'legacy-buda.exe'); $hadLegacy = $true
    }
  }
  $mutated = $true
  if (Test-Path -LiteralPath $versionDir) { Move-Item -Force -LiteralPath $versionDir -Destination (Join-Path $backupDir 'version'); $hadVersion = $true }
  $previous = ''
  if ($hadCurrent) { try { $previous = (Get-Content -Raw -LiteralPath $currentPath | ConvertFrom-Json).active } catch { $previous = '' } }
  $previousVersion = if ($previous) { $previous.Split('/')[0] } else { '' }
  New-Item -ItemType Directory -Force -Path (Join-Path $versionDir 'artifacts') | Out-Null
  Copy-Item -Force -LiteralPath (Join-Path $stage $binary) -Destination (Join-Path $versionDir 'buda.exe')
  foreach ($name in $paths) { Copy-Item -Force -LiteralPath (Join-Path $stage $name) -Destination (Join-Path $versionDir (Join-Path 'artifacts' $name)) }
  Copy-Item -Force -LiteralPath (Join-Path $stage 'checksums.txt') -Destination (Join-Path $versionDir 'artifacts/checksums.txt')
  $newLauncher = Join-Path $binDir '.buda-launcher-new.exe'
  Copy-Item -Force -LiteralPath (Join-Path $stage $launcher) -Destination $newLauncher
  Move-Item -Force -LiteralPath $newLauncher -Destination $launcherPath
  $pointer = @{ schema = 1; active = "$Version/buda.exe"; previous = $previous; active_version = $Version; previous_version = $previousVersion } | ConvertTo-Json -Compress
  Write-Atomic $currentPath $pointer
  Copy-Item -Force -LiteralPath (Join-Path $stage 'artifacts.json') -Destination (Join-Path $cliHome '.installed-artifacts.json.new')
  Move-Item -Force -LiteralPath (Join-Path $cliHome '.installed-artifacts.json.new') -Destination $manifestPath
  $committed = $true
  $observed = (& $launcherPath --version | Out-String).Trim()
  if ($observed -ne $Version) { throw "Stable launcher reported '$observed'; expected '$Version'." }
  $selfTest = (& $launcherPath __self-test | Out-String).Trim()
  if ($selfTest -ne 'ok') { throw 'Stable launcher self-test failed.' }
  $userPath = [Environment]::GetEnvironmentVariable('Path','User'); $entries = @($userPath -split ';' | Where-Object { $_ }); if ($entries -notcontains $binDir) { [Environment]::SetEnvironmentVariable('Path', (($entries + $binDir) -join ';'), 'User') }
  if ($WikiId) { & $launcherPath init --wiki $Wiki --wiki-id $WikiId } else { & $launcherPath init --wiki $Wiki }; if ($LASTEXITCODE -ne 0) { throw 'Buda installed, but explicit-wiki init failed.' }
  if ($hadLegacy -and (Test-Path -LiteralPath $legacyPath)) { Remove-Item -Force -LiteralPath $legacyPath }
  Write-Host "Installed Buda $Version"
  Write-Host "Launcher: $launcherPath"
  Write-Host "Payload: $(Join-Path $versionDir 'buda.exe')"
  Write-Host "CLI home: $cliHome"
}
catch {
  Restore-Previous
  throw
}
finally { if (Test-Path -LiteralPath $stage) { Remove-Item -Recurse -Force -LiteralPath $stage } }
