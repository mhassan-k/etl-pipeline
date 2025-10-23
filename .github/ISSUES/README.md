# Repository Issues

This directory contains detailed issue reports identified during a comprehensive repository review conducted on **2025-10-23**.

## Quick Summary

- **Total Issues:** 21
- **Critical:** 1 (blocking)
- **High Priority:** 8
- **Medium Priority:** 7
- **Low Priority:** 4

## How to Create These Issues on GitHub

Since GitHub CLI (`gh`) is not available in this environment, you have several options:

### Option 1: Manual Creation (Recommended for Quick Start)
1. Go to https://github.com/mhassan-k/etl-pipeline/issues/new
2. Copy the title from the markdown file `---` frontmatter
3. Copy the body content
4. Add the labels specified in the frontmatter
5. Click "Submit new issue"

### Option 2: Bulk Import via GitHub API
Use the provided script to create all issues programmatically:

```bash
# Set your GitHub token
export GITHUB_TOKEN="your_github_personal_access_token"

# Run the import script
./create-issues.sh
```

### Option 3: Use GitHub CLI (if you install it)
```bash
# Install gh
brew install gh  # macOS
# or
sudo apt install gh  # Ubuntu

# Authenticate
gh auth login

# Create issues from files
for file in .github/ISSUES/*.md; do
  [ "$file" = ".github/ISSUES/README.md" ] && continue

  title=$(grep "^title:" "$file" | sed 's/title: //' | tr -d '"')
  labels=$(grep "^labels:" "$file" | sed 's/labels: //')
  body=$(sed '1,/^---$/d; /^---$/,/^$/d' "$file")

  gh issue create --title "$title" --body "$body" --label "$labels"
done
```

## Issue Files

| Priority | File | Title |
|----------|------|-------|
| 🔴 CRITICAL | `01-critical-test-compilation.md` | Tests fail to compile - unused import |
| 🟠 HIGH | `02-unused-dependency.md` | Remove unused cron dependency |
| 🟠 HIGH | `03-env-example-missing.md` | Add .env.example file |
| 🟠 HIGH | `04-hardcoded-credentials.md` | Move credentials to environment variables |
| 🟠 HIGH | `05-ssl-disabled.md` | Enable SSL for database connections |
| 🟠 HIGH | `06-metrics-authentication.md` | Add authentication to /metrics endpoint |
| 🟠 HIGH | `07-deployment-scripts-incomplete.md` | Implement or remove deployment scripts |
| 🟠 HIGH | `08-file-timestamp-collision.md` | Fix file timestamp precision |
| 🟠 HIGH | `09-race-condition-etl-cycles.md` | Prevent overlapping ETL cycles |
| 🟡 MEDIUM | `10-test-coverage-inadequate.md` | Expand test coverage to 80%+ |

Additional medium and low priority issues are documented in `../ISSUES_REPORT.md`.

## Recommended Action Plan

### Immediate (This Week)
1. **Fix Critical:** Issue #1 - Fix test compilation (5 minutes)
2. **Security:** Issues #4, #5, #6 - Credential and SSL fixes (2 hours)

### Short-term (This Sprint)
3. **Dependencies:** Issue #2 - Remove unused dependency (10 minutes)
4. **Developer Experience:** Issue #3 - Add .env.example (30 minutes)
5. **Bug Fixes:** Issues #8, #9 - File timestamp & race condition (3 hours)
6. **Infrastructure:** Issue #7 - Decide on deployment strategy (varies)

### Medium-term (Next Sprint)
7. **Testing:** Issue #10 - Expand test coverage (1-2 weeks)
8. **Documentation:** Additional documentation improvements
9. **Monitoring:** Enhanced metrics and observability

## Full Report

For the complete detailed report with all 21 issues, see:
- `../ISSUES_REPORT.md` - Comprehensive analysis with code examples, fixes, and testing recommendations

## Contributing

When working on these issues:
1. Reference the issue number in your commits: `fix: resolve #1 - remove unused import`
2. Follow the implementation guidance in each issue file
3. Add tests for your changes
4. Update documentation as needed
5. Link your PR to the issue

## Questions?

If you have questions about any issue:
1. Check the detailed `ISSUES_REPORT.md` first
2. Comment on the GitHub issue after creating it
3. Review the proposed solutions in each issue file

---

**Generated:** 2025-10-23
**Review Type:** Automated + Manual Code Analysis
**Coverage:** 100% of repository files
