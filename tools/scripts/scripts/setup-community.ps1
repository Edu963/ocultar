<#
.SYNOPSIS
OCULTAR Community (SMB Edition) - Windows Setup Script
#>

Write-Host "=================================================" -ForegroundColor Cyan
Write-Host " OCULTAR COMMUNITY (SMB Edition) - Setup" -ForegroundColor Cyan
Write-Host "=================================================" -ForegroundColor Cyan

# 1. Check Docker
if (!(Get-Command "docker" -ErrorAction SilentlyContinue)) {
    Write-Host "[!] Docker is not installed. Please install Docker Desktop from https://www.docker.com/products/docker-desktop/" -ForegroundColor Red
    Pause
    Exit
}

# 2. Check/Setup .env
if (!(Test-Path ".env")) {
    if (!(Test-Path ".env.example")) {
        Write-Host "[!] .env.example not found in the current directory." -ForegroundColor Red
        Pause
        Exit
    }
    Write-Host "[i] Initialising secure environment (.env)..." -ForegroundColor Yellow
    Copy-Item ".env.example" ".env"
    
    # Generate pseudo-random keys (Windows native fallback since openssl might not be installed)
    $masterBytes = New-Object byte[] 32
    $saltBytes = New-Object byte[] 16
    $rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
    $rng.GetBytes($masterBytes)
    $rng.GetBytes($saltBytes)
    $masterKey = [System.BitConverter]::ToString($masterBytes).Replace("-", "").ToLower()
    $salt = [System.BitConverter]::ToString($saltBytes).Replace("-", "").ToLower()

    $envContent = Get-Content ".env"
    $envContent = $envContent -replace "replace-with-a-secure-32-byte-key", $masterKey
    $envContent = $envContent -replace "ocultar-v112-kdf-salt-fixed-16", $salt
    Set-Content -Path ".env" -Value $envContent

    Write-Host "[`u{2713}] Secure keys generated and saved to .env" -ForegroundColor Green
} else {
    Write-Host "[`u{2713}] Existing .env file found." -ForegroundColor Green
}

# 3. Pull and Start
Write-Host "[i] Pulling AI models & starting the OCULTAR vault... (This may take a few minutes on first run)" -ForegroundColor Yellow
docker compose pull
docker compose up -d

# 4. Open dashboard
Write-Host "[`u{2713}] OCULTAR Engine is running!" -ForegroundColor Green
Write-Host "[i] Launching dashboard..." -ForegroundColor Yellow

Start-Process "http://localhost:9090"

Write-Host "=================================================" -ForegroundColor Cyan
Write-Host " SETUP COMPLETE. You may close this window." -ForegroundColor Cyan
Write-Host "=================================================" -ForegroundColor Cyan
Pause
