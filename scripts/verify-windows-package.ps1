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
foreach ($file in @('eit2.exe', 'LICENSE', 'NOTICE.md', 'ASSETS.md', 'LICENSES/Apache-2.0.txt')) {
  if (-not (Test-Path (Join-Path $directory $file))) { throw "Missing package file: $file" }
}
$output = & (Join-Path $directory 'eit2.exe') --version
if ($output -ne "Eit 2 v$version") { throw "Unexpected version output: $output" }
Write-Host "Windows package verification passed for v$version"
