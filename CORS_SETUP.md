# CORS Configuration Guide

## Overview

The PersonaForge backend includes a robust, configurable CORS (Cross-Origin Resource Sharing) middleware that allows you to control which origins can access your API.

## Configuration

CORS is configured via the `CORS_ORIGINS` environment variable.

### Environment Variable

```bash
CORS_ORIGINS="https://example.com,https://app.example.com,http://localhost:3000"
```

## Configuration Options

### 1. Allow All Origins (Development Only)

**⚠️ Not recommended for production**

```bash
CORS_ORIGINS="*"
```

This allows requests from any origin. Use only for development/testing.

### 2. Allow Specific Origins (Recommended for Production)

```bash
CORS_ORIGINS="https://yourdomain.com,https://www.yourdomain.com"
```

Only the listed origins will be allowed to make requests to your API.

### 3. Allow Multiple Origins (Multiple Environments)

```bash
CORS_ORIGINS="https://production.com,https://staging.example.com,http://localhost:3000"
```

Comma-separated list supporting production, staging, and local development.

### 4. Wildcard Subdomains

```bash
CORS_ORIGINS="*.example.com"
```

Allows all subdomains of example.com (e.g., app.example.com, api.example.com).

## Examples

### Local Development

```bash
# .env.local
CORS_ORIGINS="http://localhost:3000,http://localhost:5173,http://127.0.0.1:3000"
```

### Production

```bash
# .env.prod
CORS_ORIGINS="https://persona-forge-ffce.onrender.com,https://www.personaforge.com"
```

### Mixed Environment

```bash
CORS_ORIGINS="https://personaforge.com,https://staging.personaforge.com,http://localhost:3000"
```

## Default Behavior

If `CORS_ORIGINS` is not set, the middleware defaults to `"*"` (allow all origins).

## CORS Headers Set

The middleware automatically sets the following headers:

- `Access-Control-Allow-Origin`: The allowed origin(s)
- `Access-Control-Allow-Methods`: GET, POST, PUT, PATCH, DELETE, OPTIONS
- `Access-Control-Allow-Headers`: Origin, Content-Type, Accept, Authorization, X-Requested-With
- `Access-Control-Allow-Credentials`: true
- `Access-Control-Max-Age`: 86400 (24 hours)
- `Access-Control-Expose-Headers`: Content-Length

## Preflight Requests

The middleware automatically handles OPTIONS preflight requests and returns a 204 No Content response.

## Security Best Practices

1. **Never use `*` in production** - Always specify exact origins
2. **Use HTTPS in production** - Only allow `https://` origins in production
3. **Be specific** - Only allow origins you control
4. **Regularly review** - Audit your allowed origins periodically

## Troubleshooting

### CORS Error in Browser

If you see CORS errors in the browser console:

1. Check that your frontend origin is in the `CORS_ORIGINS` list
2. Ensure the origin matches exactly (including protocol and port)
3. Verify the backend server is running
4. Check browser console for the exact error message

### Common Issues

**Issue**: "No 'Access-Control-Allow-Origin' header is present"
- **Solution**: Add your frontend URL to `CORS_ORIGINS`

**Issue**: Works in normal browser but fails in incognito
- **Solution**: Ensure CORS is properly configured (not relying on cached preflight)

**Issue**: Works on localhost but fails on deployed site
- **Solution**: Update `CORS_ORIGINS` to include your deployed frontend URL

## Advanced Customization

To customize CORS behavior, modify `internal/middleware/cors.go`:

```go
// Example: Custom CORS configuration
config := &CORSConfig{
    AllowedOrigins:   []string{"https://example.com"},
    AllowedMethods:   []string{"GET", "POST"},
    AllowedHeaders:   []string{"Content-Type", "Authorization"},
    ExposedHeaders:   []string{"Content-Length"},
    AllowCredentials: true,
    MaxAge:           3600,
}
```

Then update `internal/server/server.go` to use your custom config:

```go
s.router.Use(middleware.CORS(config))
```
