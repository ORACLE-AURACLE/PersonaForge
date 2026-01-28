#!/bin/bash
# Database Setup Script for PersonaForge
# This script helps you create the database and user in PostgreSQL

echo "PersonaForge Database Setup"
echo "============================"
echo ""

# Read database configuration from .env.local
if [ -f .env.local ]; then
    source .env.local
    DB_HOST=${DB_HOST:-localhost}
    DB_PORT=${DB_PORT:-5432}
    DB_NAME=${DB_NAME:-personaforge}
    DB_USER=${DB_USER:-PersonaForge}
    DB_PASSWORD=${DB_PASSWORD:-password}
else
    echo "Warning: .env.local not found. Using defaults."
    DB_HOST=localhost
    DB_PORT=5432
    DB_NAME=personaforge
    DB_USER=PersonaForge
    DB_PASSWORD=password
fi

echo "Configuration:"
echo "  Host: $DB_HOST"
echo "  Port: $DB_PORT"
echo "  Database: $DB_NAME"
echo "  User: $DB_USER"
echo ""

# Check if PostgreSQL is running
if ! pg_isready -h "$DB_HOST" -p "$DB_PORT" > /dev/null 2>&1; then
    echo "Error: PostgreSQL is not running or not accessible at $DB_HOST:$DB_PORT"
    echo "Please start PostgreSQL and try again."
    exit 1
fi

echo "PostgreSQL is running. Proceeding with setup..."
echo ""

# Connect as postgres superuser to create user and database
echo "Step 1: Creating user '$DB_USER' (if it doesn't exist)..."
psql -h "$DB_HOST" -p "$DB_PORT" -U postgres -c "CREATE USER \"$DB_USER\" WITH PASSWORD '$DB_PASSWORD';" 2>/dev/null
if [ $? -eq 0 ]; then
    echo "✓ User '$DB_USER' created successfully"
else
    echo "  User '$DB_USER' may already exist (this is okay)"
fi

echo ""
echo "Step 2: Creating database '$DB_NAME' (if it doesn't exist)..."
psql -h "$DB_HOST" -p "$DB_PORT" -U postgres -c "CREATE DATABASE \"$DB_NAME\" OWNER \"$DB_USER\";" 2>/dev/null
if [ $? -eq 0 ]; then
    echo "✓ Database '$DB_NAME' created successfully"
else
    echo "  Database '$DB_NAME' may already exist (this is okay)"
fi

echo ""
echo "Step 3: Granting privileges..."
psql -h "$DB_HOST" -p "$DB_PORT" -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE \"$DB_NAME\" TO \"$DB_USER\";" 2>/dev/null
echo "✓ Privileges granted"

echo ""
echo "============================"
echo "Setup complete!"
echo ""
echo "You can now connect to the database using:"
echo "  Host: $DB_HOST"
echo "  Port: $DB_PORT"
echo "  Database: $DB_NAME"
echo "  User: $DB_USER"
echo "  Password: $DB_PASSWORD"
echo ""
echo "In pgAdmin, use these credentials to create a new server connection."
