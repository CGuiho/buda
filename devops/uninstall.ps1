[CmdletBinding()]
param(
  [string]$Wiki,
  [switch]$PreserveConfig,
  [switch]$PreserveData,
  [switch]$DryRun,
  [switch]$Yes
)
$ErrorActionPreference='Stop'
$homeDir=[Environment]::GetFolderPath('UserProfile'); $guiho=Join-Path $homeDir '.guiho'; $cliHome=Join-Path $guiho 'buda'; $bin=Join-Path $guiho 'bin'; $launcher=Join-Path $bin 'buda.exe'
$tempRoot=Join-Path $guiho '.temp'
if ((Test-Path -LiteralPath $launcher) -and $env:BUDA_UNINSTALL_FALLBACK -ne '1') {
  $cliArgs=@('uninstall','--yes')
  if ($PreserveConfig) { $cliArgs += '--preserve-config' }
  if ($PreserveData) { $cliArgs += '--preserve-data' }
  if ($DryRun) { $cliArgs += '--dry-run' }
  if ($Wiki) { $cliArgs += @('--wiki',$Wiki) }
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
        $cliArgs=@('uninstall','--yes')
        if ($PreserveConfig) { $cliArgs += '--preserve-config' }
        if ($PreserveData) { $cliArgs += '--preserve-data' }
        if ($DryRun) { $cliArgs += '--dry-run' }
        if ($Wiki) { $cliArgs += @('--wiki',$Wiki) }
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
Write-Host "REMOVE $launcher"
Write-Host "REMOVE $(Join-Path $homeDir '.agents\skills\guiho-s-0002-buda')"
Write-Host "REMOVE $(Join-Path $homeDir '.claude\skills\guiho-s-0002-buda')"
Write-Host "PRESERVE $tempRoot"
if ($PreserveData) { Write-Host "PRESERVE $cliHome" } else { Write-Host "REMOVE $cliHome" }
if ($Wiki) { if ($PreserveConfig) { Write-Host "PRESERVE $(Join-Path $Wiki 'buda.yaml')" } else { Write-Host "REMOVE $(Join-Path $Wiki 'buda.yaml')" }; Write-Host "REMOVE managed block $(Join-Path $Wiki 'AGENTS.md')" }
if ($DryRun) { exit 0 }
if (-not $Yes) { if (-not [Environment]::UserInteractive) { throw '--Yes is required for non-interactive uninstall.' }; $answer=Read-Host 'Remove Buda-owned files? [y/N]'; if ($answer -notin @('y','Y','yes','YES')) { throw 'Uninstall cancelled.' } }
if (Test-Path -LiteralPath $launcher) { Remove-Item -Force -LiteralPath $launcher }
Remove-Item -Force -ErrorAction SilentlyContinue -LiteralPath (Join-Path $homeDir '.agents\skills\guiho-s-0002-buda\SKILL.md'),(Join-Path $homeDir '.claude\skills\guiho-s-0002-buda\SKILL.md')
Remove-Item -Recurse -Force -ErrorAction SilentlyContinue -LiteralPath (Join-Path $cliHome 'versions'),(Join-Path $cliHome 'state'),(Join-Path $cliHome 'current.json'),(Join-Path $cliHome 'installed-artifacts.json'),(Join-Path $cliHome 'cache.json')
if (-not $PreserveData) { Remove-Item -Recurse -Force -ErrorAction SilentlyContinue -LiteralPath (Join-Path $cliHome 'data'),(Join-Path $cliHome 'database') }
if (-not $PreserveConfig) { Remove-Item -Force -ErrorAction SilentlyContinue -LiteralPath (Join-Path $cliHome 'buda.global.yaml') }
if (-not $PreserveConfig -and -not $PreserveData) { Remove-Item -Force -ErrorAction SilentlyContinue -LiteralPath $cliHome }
if ($Wiki) {
  if (-not $PreserveConfig) { Remove-Item -Force -ErrorAction SilentlyContinue -LiteralPath (Join-Path $Wiki 'buda.yaml') }
  foreach ($name in @('AGENTS.md','CLAUDE.md')) {
    $instruction=Join-Path $Wiki $name
    if (Test-Path -LiteralPath $instruction) { $content=Get-Content -Raw -LiteralPath $instruction; $content=[regex]::Replace($content,'(?s)<!-- BEGIN BUDA INSTRUCTIONS -->.*?<!-- END BUDA INSTRUCTIONS -->\r?\n?',''); Set-Content -NoNewline -LiteralPath $instruction -Value $content }
  }
  Remove-Item -Force -ErrorAction SilentlyContinue -LiteralPath (Join-Path $Wiki '.agents\skills\guiho-s-0002-buda\SKILL.md'),(Join-Path $Wiki '.claude\skills\guiho-s-0002-buda\SKILL.md')
}
Write-Host 'Buda uninstall completed synchronously.'
