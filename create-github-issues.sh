#!/bin/bash
#
# Script to bulk-create GitHub issues from markdown files
# Usage: GITHUB_TOKEN=your_token ./create-github-issues.sh
#

set -e

# Configuration
REPO_OWNER="mhassan-k"
REPO_NAME="etl-pipeline"
ISSUES_DIR=".github/ISSUES"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check for GitHub token
if [ -z "$GITHUB_TOKEN" ]; then
    echo -e "${RED}Error: GITHUB_TOKEN environment variable is required${NC}"
    echo "Usage: GITHUB_TOKEN=your_token ./create-github-issues.sh"
    echo ""
    echo "Create a token at: https://github.com/settings/tokens"
    echo "Required scope: 'repo' (Full control of private repositories)"
    exit 1
fi

# Check if issues directory exists
if [ ! -d "$ISSUES_DIR" ]; then
    echo -e "${RED}Error: Issues directory not found: $ISSUES_DIR${NC}"
    exit 1
fi

echo -e "${GREEN}Creating GitHub issues for $REPO_OWNER/$REPO_NAME${NC}"
echo ""

# Counter for created issues
created=0
skipped=0
failed=0

# Process each markdown file
for file in "$ISSUES_DIR"/*.md; do
    # Skip README
    if [ "$(basename "$file")" = "README.md" ]; then
        continue
    fi

    echo -e "${YELLOW}Processing: $(basename "$file")${NC}"

    # Extract title from frontmatter
    title=$(grep "^title:" "$file" | sed 's/title: *//' | tr -d '"[]')

    # Extract labels from frontmatter
    labels=$(grep "^labels:" "$file" | sed 's/labels: *//')

    # Extract body (everything after the second --- marker)
    body=$(sed '1,/^---$/d; /^---$/,/^$/d' "$file")

    if [ -z "$title" ]; then
        echo -e "  ${RED}✗ Skipped - no title found${NC}"
        ((skipped++))
        continue
    fi

    # Convert labels to JSON array
    labels_json=$(echo "$labels" | awk -F', *' '
        BEGIN { printf "[" }
        {
            for(i=1; i<=NF; i++) {
                if(i>1) printf ","
                printf "\"" $i "\""
            }
        }
        END { printf "]" }
    ')

    # Escape body for JSON
    body_escaped=$(echo "$body" | jq -Rs .)

    # Create JSON payload
    json_payload=$(cat <<EOF
{
  "title": $(echo "$title" | jq -Rs .),
  "body": $body_escaped,
  "labels": $labels_json
}
EOF
)

    # Create the issue via GitHub API
    response=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Authorization: token $GITHUB_TOKEN" \
        -H "Accept: application/vnd.github.v3+json" \
        "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/issues" \
        -d "$json_payload")

    # Extract HTTP status code
    http_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | sed '$d')

    if [ "$http_code" -eq 201 ]; then
        issue_number=$(echo "$response_body" | jq -r '.number')
        issue_url=$(echo "$response_body" | jq -r '.html_url')
        echo -e "  ${GREEN}✓ Created issue #$issue_number${NC}"
        echo -e "    $issue_url"
        ((created++))
    else
        echo -e "  ${RED}✗ Failed (HTTP $http_code)${NC}"
        error_message=$(echo "$response_body" | jq -r '.message // "Unknown error"')
        echo -e "    Error: $error_message"
        ((failed++))
    fi

    echo ""

    # Be nice to GitHub API - rate limiting
    sleep 1
done

# Summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}Summary${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "  Created: ${GREEN}$created${NC}"
echo -e "  Skipped: ${YELLOW}$skipped${NC}"
echo -e "  Failed:  ${RED}$failed${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ $created -gt 0 ]; then
    echo ""
    echo -e "${GREEN}View all issues at:${NC}"
    echo "https://github.com/$REPO_OWNER/$REPO_NAME/issues"
fi

exit 0
