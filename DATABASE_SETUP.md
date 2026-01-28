# Database Setup Guide

## Local Development (pgAdmin)

### Current Configuration
Based on your `.env.local`:
- **Host**: `localhost`
- **Port**: `5432`
- **Username**: `postgres`
- **Password**: `postgres`
- **Database**: `personaforge`

### Connecting via pgAdmin



1. **Open pgAdmin** and connect to your PostgreSQL server:
   - Right-click on "Servers" → "Create" → "Server"
   - **General Tab**:
     - Name: `Local PostgreSQL` (or any name you prefer)
   - **Connection Tab**:
     - Host name/address: `localhost`
     - Port: `5432`
     - Maintenance database: `postgres`
     - Username: `postgres`
     - Password: `postgres` (check "Save password" if you want)
   - Click "Save"

2. **Create the database** (if it doesn't exist):
   - Right-click on "Databases" → "Create" → "Database"
   - **General Tab**:
     - Database: `personaforge`
   - Click "Save"

3. **Verify the connection**:
   - Expand your server → Databases → `personaforge`
   - You should see tables after running migrations

### Testing the Connection

You can test if your connection works by running:
```bash
make run
```

The application will:
1. Load `.env.local` automatically
2. Connect to `localhost:5432/personaforge`
3. Run database migrations automatically
4. Start the server

## Production/Hosting Setup

### Option 1: Using `.env.prod` file

1. **Create/Update `.env.prod`** with your production database:
   ```env
   ENV=production
   PORT=8080
   
   DATABASE_URL=postgres://username:password@your-production-host:5432/personaforge?sslmode=require
   JWT_SECRET=your-production-secret-key
   JWT_EXPIRY_MINUTES=30
   
   GEMINI_API_KEY=your-api-key
   GEMINI_MODEL=gemini-2.5-flash
   
   GOOGLE_CLIENT_ID=your-client-id
   ```

2. **Deploy using Docker**:
   ```bash
   make docker-prod
   ```

### Option 2: Environment Variables (Cloud Platforms)

For platforms like Heroku, Railway, Render, etc.:

1. **Set environment variables** in your platform's dashboard:
   - `DATABASE_URL` - Your production database connection string
   - `JWT_SECRET` - Production secret
   - `GEMINI_API_KEY` - Your API key
   - `GOOGLE_CLIENT_ID` - Your client ID
   - `ENV=production`
   - `PORT` - Usually set automatically by platform

2. **The application will automatically use these** (no `.env` file needed in production)

## Database Connection String Format

```
postgres://[username]:[password]@[host]:[port]/[database]?sslmode=[mode]
```

Examples:
- **Local**: `postgres://postgres:postgres@localhost:5432/personaforge?sslmode=disable`
- **Production**: `postgres://user:pass@db.example.com:5432/personaforge?sslmode=require`

## Troubleshooting

### "Password authentication failed"
- Verify the password in `.env.local` matches your PostgreSQL password
- Check if PostgreSQL is running: `pg_isready -h localhost -p 5432`

### "Database does not exist"
- Create it via pgAdmin (see step 2 above)
- Or via command line: `createdb -U postgres personaforge`

### "Connection refused"
- Ensure PostgreSQL is running
- Check if port 5432 is correct
- Verify firewall settings

## Quick Commands

```bash
# Test database connection
psql -h localhost -U postgres -d personaforge

# Create database (if needed)
createdb -U postgres personaforge

# Run application locally
make run

# Run with Docker (includes database)
make docker-local
```
