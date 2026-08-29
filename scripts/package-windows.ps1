$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root
$version = (Get-Content VERSION -Raw).Trim()
$ldflags = "-s -w -H windowsgui -X github.com/vibloteket/eit2/internal/version.Value=$version"
$packages = Join-Path $root 'dist/packages'
$stage = Join-Path $root 'dist/stage-windows'
$folderName = "eit2-windows-x86_64-v$version"
$directory = Join-Path $stage $folderName
$archive = Join-Path $packages "$folderName.zip"

Remove-Item $packages, $stage -Recurse -Force -ErrorAction SilentlyContinue
New-Item $directory -ItemType Directory -Force | Out-Null

$env:GOOS = 'windows'
$env:GOARCH = 'amd64'
$env:CGO_ENABLED = '0'
go build -trimpath -ldflags $ldflags -o (Join-Path $directory 'eit2.exe') ./cmd/eit2
if ($LASTEXITCODE -ne 0) { throw 'Windows build failed' }

Copy-Item packaging/README.txt, LICENSE, NOTICE.md, ASSETS.md -Destination $directory
Copy-Item LICENSES -Destination $directory -Recurse
Compress-Archive -Path $directory -DestinationPath $archive -CompressionLevel Optimal

$hash = (Get-FileHash $archive -Algorithm SHA256).Hash.ToLowerInvariant()
"$hash  $folderName.zip" | Set-Content "$archive.sha256" -Encoding ascii
Write-Host "Created $archive"
