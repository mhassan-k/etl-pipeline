---
title: "Add .env.example file for configuration guidance"
labels: documentation, developer-experience, configuration, good-first-issue
---

## Description
No `.env.example` file exists to guide users on required environment variables. New users must read code or documentation to discover configuration options.

## Impact
- Poor developer experience
- Configuration errors
- Delayed onboarding for new contributors
- Trial and error to get app running

## Proposed Solution
Create `.env.example` with all configurable variables and safe defaults:

```env
# API Configuration
API_URL=https://jsonplaceholder.typicode.com/posts

# Database Configuration
# For production, use sslmode=require or sslmode=verify-full
DATABASE_URL=postgres://etl_user:etl_password@localhost:5432/etl_db?sslmode=prefer

# ETL Configuration
# Interval in seconds between data fetches
FETCH_INTERVAL=30

# Server Configuration
SERVER_PORT=8080
```

## Additional Improvements
Also update README.md to reference the `.env.example` file:
```markdown
## Configuration

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` with your configuration values
```

## Files to Create
- `.env.example` (root directory)
- Update `README.md` to reference it

## Priority
**High** - Improves onboarding experience

## Acceptance Criteria
- [ ] `.env.example` created with all environment variables
- [ ] Each variable has a comment explaining its purpose
- [ ] Safe default values provided
- [ ] README.md updated to reference the file
- [ ] Add `.env` to `.gitignore` if not already present
