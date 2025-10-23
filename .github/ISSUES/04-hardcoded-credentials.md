---
title: "Move hardcoded credentials in docker-compose.yml to environment variables"
labels: security, configuration, docker
---

## Description
Database credentials are hardcoded in `docker-compose.yml` instead of using environment variables or `.env` file, violating 12-factor app principles and creating security risks.

## File Location
`docker-compose.yml:8-10, 30`

## Current Code
```yaml
services:
  postgres:
    environment:
      POSTGRES_USER: etl_user        # ❌ Hardcoded
      POSTGRES_PASSWORD: etl_password # ❌ Hardcoded
      POSTGRES_DB: etl_db

  etl-pipeline:
    environment:
      DATABASE_URL: postgres://etl_user:etl_password@postgres:5432/etl_db?sslmode=disable
```

## Security Risks
- Production credentials could be accidentally committed
- Cannot easily change credentials per environment
- Credentials visible in plain text in repository
- Violates security best practices

## Proposed Fix
Update `docker-compose.yml` to use environment variables:

```yaml
services:
  postgres:
    environment:
      POSTGRES_USER: ${POSTGRES_USER:-etl_user}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-etl_password}
      POSTGRES_DB: ${POSTGRES_DB:-etl_db}

  etl-pipeline:
    environment:
      DATABASE_URL: ${DATABASE_URL:-postgres://etl_user:etl_password@postgres:5432/etl_db?sslmode=prefer}
```

Then create `.env` file (gitignored):
```env
POSTGRES_USER=etl_user
POSTGRES_PASSWORD=secure_password_here
POSTGRES_DB=etl_db
DATABASE_URL=postgres://etl_user:secure_password_here@postgres:5432/etl_db?sslmode=prefer
```

## Additional Changes Needed
1. Ensure `.env` is in `.gitignore`
2. Create `.env.example` with placeholder values
3. Update README with setup instructions

## Priority
**High** - Security improvement

## Acceptance Criteria
- [ ] `docker-compose.yml` uses environment variables
- [ ] `.env.example` created with safe defaults
- [ ] `.env` in `.gitignore`
- [ ] README updated with configuration instructions
- [ ] Tested with both default and custom values
