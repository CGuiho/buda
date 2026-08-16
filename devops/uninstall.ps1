[CmdletBinding()]
param(
  [Parameter(Mandatory=$true)][string]$Wiki,
  [switch]$PreserveConfig,
  [switch]$PreserveData,
  [switch]$DryRun,
  [switch]$Yes
)
$ErrorActionPreference='Stop'
$homeDir = if ($env:USERPROFILE) { $env:USERPROFILE } elseif ($env:HOME) { $env:HOME } else { [Environment]::GetFolderPath('UserProfile') }
$guiho=Join-Path $homeDir '.guiho'; $cliHome=Join-Path $guiho 'buda'; $bin=Join-Path $guiho 'bin'; $launcher=Join-Path $bin 'buda.exe'
$tempRoot=Join-Path $guiho '.temp'
if ((Test-Path -LiteralPath $launcher) -and $env:BUDA_UNINSTALL_FALLBACK -ne '1') {
  $cliArgs=@('uninstall','--yes','--wiki',$Wiki)
  if ($PreserveConfig) { $cliArgs += '--preserve-config' }
  if ($PreserveData) { $cliArgs += '--preserve-data' }
  if ($DryRun) { $cliArgs += '--dry-run' }
  & $launcher @cliArgs
  exit $LASTEXITCODE
}
if (-not (Test-Path -LiteralPath $launcher) -and $env:BUDA_UNINSTALL_FALLBACK -ne '1') {
  $currentPath = Join-Path $cliHome 'current.json'
  if (Test-Path -LiteralPath $currentPath) {
    try { $active = [string](Get-Content -Raw -LiteralPath $currentPath | ConvertFrom-Json).active } catch { $active = '' }
    if ($active -match '^[^\\/]+/buda(?:\.exe)?$') {
      $payload = Join-Path (Join-Path $cliHome 'versions') ($active -replace '/', [IO.Path]::DirectorySeparatorChar)
      if (Test-Path -LiteralPath $payload -PathType Leaf) {
        $cliArgs=@('uninstall','--yes','--wiki',$Wiki)
        if ($PreserveConfig) { $cliArgs += '--preserve-config' }
        if ($PreserveData) { $cliArgs += '--preserve-data' }
        if ($DryRun) { $cliArgs += '--dry-run' }
        & $payload @cliArgs
        exit $LASTEXITCODE
      }
    }
  }
}
if (Test-Path -LiteralPath $cliHome) {
  $manifestPath=Join-Path $cliHome 'installed-artifacts.json'
  if (-not (Test-Path -LiteralPath $manifestPath)) { throw 'Refusing fallback uninstall without Buda ownership manifest.' }
  $manifest=Get-Content -Raw -LiteralPath $manifestPath | ConvertFrom-Json
  if ($manifest.schema -ne 1 -or $manifest.cli -ne 'buda') { throw 'Refusing fallback uninstall for an invalid or foreign manifest.' }
  throw 'Refusing fallback uninstall: the installed Buda payload is unavailable; no ownership-safe executor remains.'
}
Write-Host "Buda Uninstall Plan:"
Write-Host "REMOVE:"
Write-Host "  - $launcher (buda stable launcher)"
Write-Host "  - $(Join-Path $homeDir '.agents\skills\guiho-s-0002-buda') (buda global agent skill)"
Write-Host "  - $(Join-Path $homeDir '.claude\skills\guiho-s-0002-buda') (buda global agent skill)"
if (-not $PreserveData) { Write-Host "  - $cliHome (buda persistent data)" }
if (-not $PreserveConfig) {
  Write-Host "  - $(Join-Path $cliHome 'buda.global.yaml') (global configuration)"
  Write-Host "  - $(Join-Path $Wiki 'buda.yaml') (selected wiki configuration)"
}
Write-Host "  - $(Join-Path $Wiki 'AGENTS.md') (managed AGENTS instruction block)"
Write-Host "PRESERVE:"
Write-Host "  - $tempRoot (shared temp directory)"
if ($PreserveData) { Write-Host "  - $cliHome (buda persistent data)" }
if ($PreserveConfig) {
  Write-Host "  - $(Join-Path $cliHome 'buda.global.yaml') (global configuration)"
  Write-Host "  - $(Join-Path $Wiki 'buda.yaml') (selected wiki configuration)"
}
if ($DryRun) { exit 0 }
if (-not $Yes) { if (-not [Environment]::UserInteractive) { throw '--Yes is required for non-interactive uninstall.' }; $answer=Read-Host 'Remove Buda-owned files? [y/N]'; if ($answer -notin @('y','Y','yes','YES')) { throw 'Uninstall cancelled.' } }
if (Test-Path -LiteralPath $launcher) { Remove-Item -Force -LiteralPath $launcher }
Remove-Item -Force -ErrorAction SilentlyContinue -LiteralPath (Join-Path $homeDir '.agents\skills\guiho-s-0002-buda\SKILL.md'),(Join-Path $homeDir '.claude\skills\guiho-s-0002-buda\SKILL.md')
if ($PreserveData) {
  Remove-Item -Recurse -Force -ErrorAction SilentlyContinue -LiteralPath (Join-Path $cliHome 'versions'),(Join-Path $cliHome 'state'),(Join-Path $cliHome 'current.json'),(Join-Path $cliHome 'installed-artifacts.json'),(Join-Path $cliHome 'cache.json')
} else {
  Remove-Item -Recurse -Force -ErrorAction SilentlyContinue -LiteralPath (Join-Path $cliHome 'versions'),(Join-Path $cliHome 'state'),(Join-Path $cliHome 'data'),(Join-Path $cliHome 'database'),(Join-Path $cliHome 'current.json'),(Join-Path $cliHome 'installed-artifacts.json'),(Join-Path $cliHome 'cache.json')
}
if (-not $PreserveConfig) { Remove-Item -Force -ErrorAction SilentlyContinue -LiteralPath (Join-Path $cliHome 'buda.global.yaml') }
if (-not $PreserveConfig -and -not $PreserveData) { Remove-Item -Force -ErrorAction SilentlyContinue -LiteralPath $cliHome }
if (-not $PreserveConfig) { Remove-Item -Force -ErrorAction SilentlyContinue -LiteralPath (Join-Path $Wiki 'buda.yaml') }
foreach ($name in @('AGENTS.md','CLAUDE.md')) {
  $instruction=Join-Path $Wiki $name
  if (Test-Path -LiteralPath $instruction) {
    $content=Get-Content -Raw -LiteralPath $instruction
    $content=[regex]::Replace($content,'(?s)<!-- BEGIN BUDA INSTRUCTIONS -->.*?<!-- END BUDA INSTRUCTIONS -->\r?\n?','')
    $tempFile=[IO.Path]::GetTempFileName()
    $encoding = New-Object System.Text.UTF8Encoding($false)
    [IO.File]::WriteAllText($tempFile, $content, $encoding)
    Move-Item -Force -LiteralPath $tempFile -Destination $instruction
  }
}
Remove-Item -Force -ErrorAction SilentlyContinue -LiteralPath (Join-Path $Wiki '.agents\skills\guiho-s-0002-buda\SKILL.md'),(Join-Path $Wiki '.claude\skills\guiho-s-0002-buda\SKILL.md')
Write-Host 'Buda uninstall completed synchronously.'
