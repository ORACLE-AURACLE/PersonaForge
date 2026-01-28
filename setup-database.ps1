# Database Setup Script for PersonaForge (PowerShell)
# This script helps you create the database and user in PostgreSQL

Write-Host "PersonaForge Database Setup" -ForegroundColor Cyan
Write-Host "============================" -ForegroundColor Cyan
Write-Host ""

# Read database configuration from .env.local
$envFile = ".env.local"
if (Test-Path $envFile) {
    Get-Content $envFile | ForEach-Object {
        if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
            $key = $matches[1].Trim()
            $value = $matches[2].Trim()
            [Environment]::SetEnvironmentVariable($key, $value, "Process")
        }
    }
}

$DB_HOST = if ($env:DB_HOST) { $env:DB_HOST } else { "localhost" }
$DB_PORT = if ($env:DB_PORT) { $env:DB_PORT } else { "5432" }
$DB_NAME = if ($env:DB_NAME) { $env:DB_NAME } else { "personaforge" }
$DB_USER = if ($env:DB_USER) { $env:DB_USER } else { "PersonaForge" }
$DB_PASSWORD = if ($env:DB_PASSWORD) { $env:DB_PASSWORD } else { "password" }

Write-Host "Configuration:" -ForegroundColor Yellow
Write-Host "  Host: $DB_HOST"
Write-Host "  Port: $DB_PORT"
Write-Host "  Database: $DB_NAME"
Write-Host "  User: $DB_USER"
Write-Host ""

# Check if PostgreSQL is running
$pgReady = & pg_isready -h $DB_HOST -p $DB_PORT 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: PostgreSQL is not running or not accessible at ${DB_HOST}:${DB_PORT}" -ForegroundColor Red
    Write-Host "Please start PostgreSQL and try again."
    exit 1
}

Write-Host "PostgreSQL is running. Proceeding with setup..." -ForegroundColor Green
Write-Host ""

# Set PGPASSWORD for psql commands
$env:PGPASSWORD = "postgres"  # Assuming postgres user password, adjust if needed

Write-Host "Step 1: Creating user '$DB_USER' (if it doesn't exist)..." -ForegroundColor Yellow
$createUser = "CREATE USER `"$DB_USER`" WITH PASSWORD '$DB_PASSWORD';"
& psql -h $DB_HOST -p $DB_PORT -U postgres -c $createUser 2>&1 | Out-Null
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ User '$DB_USER' created successfully" -ForegroundColor Green
} else {
    Write-Host "  User '$DB_USER' may already exist (this is okay)" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Step 2: Creating database '$DB_NAME' (if it doesn't exist)..." -ForegroundColor Yellow
$createDb = "CREATE DATABASE `"$DB_NAME`" OWNER `"$DB_USER`";"
& psql -h $DB_HOST -p $DB_PORT -U postgres -c $createDb 2>&1 | Out-Null
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Database '$DB_NAME' created successfully" -ForegroundColor Green
} else {
    Write-Host "  Database '$DB_NAME' may already exist (this is okay)" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Step 3: Granting privileges..." -ForegroundColor Yellow
$grantPrivs = "GRANT ALL PRIVILEGES ON DATABASE `"$DB_NAME`" TO `"$DB_USER`";"
& psql -h $DB_HOST -p $DB_PORT -U postgres -c $grantPrivs 2>&1 | Out-Null
Write-Host "✓ Privileges granted" -ForegroundColor Green

Write-Host ""
Write-Host "============================" -ForegroundColor Cyan
Write-Host "Setup complete!" -ForegroundColor Green
Write-Host ""
Write-Host "You can now connect to the database using:" -ForegroundColor Yellow
Write-Host "  Host: $DB_HOST"
Write-Host "  Port: $DB_PORT"
Write-Host "  Database: $DB_NAME"
Write-Host "  User: $DB_USER"
Write-Host "  Password: $DB_PASSWORD"
Write-Host ""
Write-Host "In pgAdmin, use these credentials to create a new server connection." -ForegroundColor Cyan
