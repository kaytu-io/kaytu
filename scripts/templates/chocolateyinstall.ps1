$ErrorActionPreference = 'Stop'

$url64 = "https://github.com/kaytu-io/kaytu/releases/download/{PLACEHOLDER_VERSION}/kaytu_{PLACEHOLDER_VERSION}_windows_amd64.zip"
$unzipLocation = Split-Path -Parent $MyInvocation.MyCommand.Definition

$packageParams = @{
  PackageName   = 'kaytu'
  UnzipLocation = $unzipLocation
  Url64         = $url64
}

Install-ChocolateyZipPackage @packageParams