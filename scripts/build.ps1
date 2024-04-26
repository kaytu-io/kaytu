# this script updates powershell and nuspec files found in the the templates folder to the latest version of Infracost.

# $zip = "kaytu_$($env:CHOCO_API_KEY)_windows_amd64.zip"
$version = (Invoke-Webrequest https://api.github.com/repos/kaytu-io/kaytu/releases/latest | convertfrom-json).name

Write-Host "$(get-date) - downloading release $version"
# Invoke-WebRequest -uri "https://github.com/kaytu-io/kaytu/releases/download/$($version)/$($zip)" -OutFile $zip

# $sha = (Get-FileHash $zip).Hash
# $contents = (Get-Content $shafile)
# if ("$($sha)  $($zip)" -ne $contents) {
#   Write-Host "sha of $($sha) mismatched for downloaded artefact contents: $($contents)"
#   exit 1
# }

if (Test-Path -Path ".\tools") {
  Remove-Item .\tools -Recurse
}
New-Item .\tools -ItemType "directory"
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
# removing the first v as chocolatey doesnt like this version
$chocoVersion = $version.Substring(1, ($version.Length-1));
choco new -h
function Get-ScriptDirectory { Split-Path $MyInvocation.ScriptName }
$templatePath = Join-Path (Get-ScriptDirectory) ".\templates"

Get-Content "$($templatePath)\chocolateyinstall.ps1" | %{$_ -replace "{PLACEHOLDER_VERSION}",$version} | Out-File .\tools\chocolateyinstall.ps1
Get-Content "$($templatePath)\kaytu.nuspec" | %{$_ -replace "{PLACEHOLDER_VERSION}",$chocoVersion} | Out-File .\kaytu.nuspec

Write-Host "$(get-date) - Building choco pkg"
choco pack --version $chocoVersion

Write-Host "$(get-date) - Testing choco pkg is valid"
choco install kaytu --source .

# $out = (kaytu --version)
# if ("kaytu $($version)" -ne $out) {
#   Write-Host "kaytu output: $($out) from choco dry run install did not match expected: 'kaytu $($version)'"
#   exit 1
# }

# Write-Host "$(get-date) - Test install of kaytu passed --version check: $($out)"

Get-ChildItem *.nupkg
Write-Host "$(get-date) - Pushing to Chocolatey"
choco push -s https://push.chocolatey.org/ --api-key=$env:CHOCO_API_KEY