# Build script for CanvusNoteMapper
$ErrorActionPreference = "Stop"

# Create bin directory if it doesn't exist
if (-not (Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin"
}

# Set version
$version = "1.0.0"  # You can update this as needed

# Clean previous builds
Remove-Item -Path "bin/*" -Force -ErrorAction SilentlyContinue

# Build for Linux
Write-Host "Building for Linux..."
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
go build -o "bin/notescan" ./cmd/main.go

# Create release directory
$releaseDir = "bin/release"
if (Test-Path $releaseDir) {
    Remove-Item -Path $releaseDir -Recurse -Force
}
New-Item -ItemType Directory -Path $releaseDir

# Copy necessary files to release directory
Copy-Item "bin/notescan" -Destination "$releaseDir/"
Copy-Item "web" -Destination "$releaseDir/" -Recurse
Copy-Item ".env.example" -Destination "$releaseDir/.env" -ErrorAction SilentlyContinue

# Create a simple start script
@"
#!/bin/bash
# Start script for CanvusNoteMapper
export PORT=3030
./notescan
"@ | Out-File -FilePath "$releaseDir/start.sh" -Encoding ASCII

# Make start script executable (for Linux)
(Get-Item "$releaseDir/start.sh").UnixFileMode = 0o755

# Create README for release
@"
# CanvusNoteMapper Release v$version

## Quick Start
1. Configure your environment variables in .env file
2. Run the start script: ./start.sh

## Configuration
- PORT: Server port (default: 3030)
- GOOGLE_GENAI_API_KEY: Your Google Generative AI API key

## Files
- notescan: The main application binary
- web/: Web interface files
- .env: Configuration file
- start.sh: Startup script
"@ | Out-File -FilePath "$releaseDir/README.md" -Encoding UTF8

# Create release archive
$releaseName = "CanvusNoteMapper-v$version-linux-amd64"
Compress-Archive -Path "$releaseDir/*" -DestinationPath "bin/$releaseName.zip" -Force

Write-Host "Build complete! Release package created at bin/$releaseName.zip" 