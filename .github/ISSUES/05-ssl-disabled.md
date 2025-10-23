---
title: "Enable SSL/TLS for database connections by default"
labels: security, production-readiness, database
---

## Description
Default database connection strings use `sslmode=disable`, which transmits credentials and data in plaintext. This is insecure for production environments.

## File Locations
- `internal/config/config.go:25`
- `docker-compose.yml:30`

## Current Code
```go
// internal/config/config.go
DatabaseURL: "postgres://etl_user:etl_password@localhost:5432/etl_db?sslmode=disable"
```

## Security Risks
- Database credentials transmitted in plaintext over network
- Vulnerable to man-in-the-middle attacks
- Data can be intercepted and read
- Compliance violations (PCI-DSS, HIPAA, SOC 2, GDPR)

## Proposed Fix

### For Development (Recommended)
Use `sslmode=prefer` which attempts SSL but falls back to unencrypted:
```go
DatabaseURL: "postgres://etl_user:etl_password@localhost:5432/etl_db?sslmode=prefer"
```

### For Production (Required)
Use `sslmode=require` or `sslmode=verify-full`:
```go
// Production
DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=require

// Production with certificate verification (most secure)
DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=verify-full&sslrootcert=/path/to/ca.crt
```

## Implementation Plan

1. **Update default in config.go:**
   ```go
   func LoadConfig() *Config {
       config := &Config{
           APIUrl:        getEnv("API_URL", "https://jsonplaceholder.typicode.com/posts"),
           DatabaseURL:   getEnv("DATABASE_URL", "postgres://etl_user:etl_password@localhost:5432/etl_db?sslmode=prefer"),
           FetchInterval: getEnvAsInt("FETCH_INTERVAL", 30),
           ServerPort:    getEnv("SERVER_PORT", "8080"),
       }
       return config
   }
   ```

2. **Update docker-compose.yml:**
   ```yaml
   environment:
     DATABASE_URL: ${DATABASE_URL:-postgres://etl_user:etl_password@postgres:5432/etl_db?sslmode=prefer}
   ```

3. **Update .env.example:**
   ```env
   # Development (attempts SSL, falls back if unavailable)
   DATABASE_URL=postgres://etl_user:etl_password@localhost:5432/etl_db?sslmode=prefer

   # Production (requires SSL)
   # DATABASE_URL=postgres://user:password@host:5432/db?sslmode=require
   ```

4. **Update README with SSL configuration guidance**

## PostgreSQL SSL Mode Options
- `disable` - ❌ No SSL (current, insecure)
- `allow` - ⚠️ Try non-SSL first, use SSL if server requires
- `prefer` - ✅ Try SSL first, fallback to non-SSL (good for dev)
- `require` - ✅ Require SSL, don't verify certificate (good for prod)
- `verify-ca` - ✅ Require SSL, verify certificate against CA
- `verify-full` - ✅ Require SSL, verify certificate and hostname (best for prod)

## Priority
**High** - Security vulnerability in production

## Acceptance Criteria
- [ ] Default connection string uses `sslmode=prefer` minimum
- [ ] Documentation includes SSL configuration options
- [ ] `.env.example` shows both dev and prod SSL configurations
- [ ] README has section on secure database connections
- [ ] Tested with both SSL and non-SSL PostgreSQL instances
