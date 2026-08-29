$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root
$version = (Get-Content VERSION -Raw).Trim()
$name = "eit2-windows-x86_64-v$version"
$packages = Join-Path $root 'dist/packages'
$archive = Join-Path $packages "$name.zip"
$expected = ((Get-Content "$archive.sha256" -Raw).Trim() -split '\s+')[0]
$actual = (Get-FileHash $archive -Algorithm SHA256).Hash.ToLowerInvariant()
if ($actual -ne $expected) { throw "Checksum mismatch: $actual != $expected" }

$destination = Join-Path $root 'dist/verify-windows'
Remove-Item $destination -Recurse -Force -ErrorAction SilentlyContinue
Expand-Archive $archive -DestinationPath $destination
$directory = Join-Path $destination $name
foreach ($file in @('eit2.exe', 'VERSION.txt', 'LICENSE', 'NOTICE.md', 'ASSETS.md', 'LICENSES/Apache-2.0.txt')) {
  if (-not (Test-Path (Join-Path $directory $file))) { throw "Missing package file: $file" }
}
$packagedVersion = (Get-Content (Join-Path $directory 'VERSION.txt') -Raw).Trim()
if ($packagedVersion -ne $version) { throw "Unexpected packaged version: $packagedVersion" }
$header = [System.IO.File]::ReadAllBytes((Join-Path $directory 'eit2.exe'))
if ($header[0] -ne 0x4D -or $header[1] -ne 0x5A) { throw 'eit2.exe is not a PE executable' }
Write-Host "Windows package verification passed for v$version"
