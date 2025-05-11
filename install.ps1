# PowerShell script for installing CDX on Windows
# This script has been AI-generated and has not been tested

# Check if Go is installed
if (!(Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "Error: Go is not installed. Please install Go before continuing." -ForegroundColor Red
    exit 1
}

# Create installation directory
$InstallDir = "$env:USERPROFILE\.cdx"
$BinDir = "$InstallDir\bin"

# Create directories if they don't exist
if (!(Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir | Out-Null
}
if (!(Test-Path $BinDir)) {
    New-Item -ItemType Directory -Path $BinDir | Out-Null
}

# Clone or update repository
$SrcDir = "$InstallDir\src"
if (Test-Path "$SrcDir\.git") {
    Write-Host "Updating CDX..." -ForegroundColor Blue
    Set-Location $SrcDir
    git pull
}
else {
    Write-Host "Downloading CDX..." -ForegroundColor Blue
    if (Test-Path $SrcDir) {
        Remove-Item -Recurse -Force $SrcDir
    }
    git clone https://github.com/RodPaDev/cdx.git $SrcDir
    Set-Location $SrcDir
}

# Build the application
Write-Host "Building CDX..." -ForegroundColor Blue
go build -o "$BinDir\cdx.exe"

# Add to PATH
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (!($UserPath -like "*$BinDir*")) {
    [Environment]::SetEnvironmentVariable("Path", "$UserPath;$BinDir", "User")
    Write-Host "Added CDX to PATH" -ForegroundColor Green
    Write-Host "Note: You may need to restart your terminal or computer for the PATH changes to take effect." -ForegroundColor Blue
}
else {
    Write-Host "CDX is already in PATH" -ForegroundColor Green
}

Write-Host "CDX installed successfully!" -ForegroundColor Green
Write-Host "Run 'cdx' to start the application" -ForegroundColor Blue