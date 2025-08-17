#!/bin/bash

# Script to check for potential secrets before commit

set -e

echo "🔍 Checking for potential secrets..."

# Common secret patterns
SECRET_PATTERNS=(
    "password\s*=\s*['\"][^'\"]{4,}['\"]"
    "secret\s*=\s*['\"][^'\"]{4,}['\"]"
    "key\s*=\s*['\"][^'\"]{4,}['\"]"
    "token\s*=\s*['\"][^'\"]{4,}['\"]"
    "api[_-]?key\s*=\s*['\"][^'\"]{4,}['\"]"
    "auth[_-]?token\s*=\s*['\"][^'\"]{4,}['\"]"
    "access[_-]?key\s*=\s*['\"][^'\"]{4,}['\"]"
    "private[_-]?key"
    "-----BEGIN\s+(RSA\s+)?PRIVATE\s+KEY-----"
    "jwt[_-]?secret\s*=\s*['\"][^'\"]{4,}['\"]"
    "database[_-]?url\s*=\s*['\"][^'\"]*://[^'\"]*['\"]"
)

# Files to check (staged for commit)
FILES=$(git diff --cached --name-only --diff-filter=ACM)

if [ -z "$FILES" ]; then
    echo "✅ No files to check"
    exit 0
fi

FOUND_SECRETS=false

for file in $FILES; do
    if [ -f "$file" ]; then
        # Skip certain file types and examples
        if [[ "$file" =~ \.(example|template|md|go\.sum)$ ]] || [[ "$file" =~ example ]]; then
            continue
        fi
        
        # Check each pattern
        for pattern in "${SECRET_PATTERNS[@]}"; do
            if grep -iE "$pattern" "$file" > /dev/null 2>&1; then
                echo "❌ Potential secret found in $file:"
                grep -iE --color=always "$pattern" "$file" || true
                FOUND_SECRETS=true
            fi
        done
        
        # Check for hardcoded IPs (except localhost/examples)
        if grep -E "([0-9]{1,3}\.){3}[0-9]{1,3}" "$file" | grep -v -E "(127\.0\.0\.1|localhost|0\.0\.0\.0|example)" > /dev/null 2>&1; then
            echo "⚠️  Hardcoded IP address found in $file:"
            grep -E --color=always "([0-9]{1,3}\.){3}[0-9]{1,3}" "$file" | grep -v -E "(127\.0\.0\.1|localhost|0\.0\.0\.0|example)" || true
        fi
        
        # Check for base64 encoded secrets (common pattern)
        if grep -E "[A-Za-z0-9+/]{20,}={0,2}" "$file" | grep -v -E "(example|test|demo)" > /dev/null 2>&1; then
            echo "⚠️  Potential base64 encoded data in $file (review manually):"
            grep -E --color=always "[A-Za-z0-9+/]{20,}={0,2}" "$file" | head -5 || true
        fi
    fi
done

if [ "$FOUND_SECRETS" = true ]; then
    echo ""
    echo "❌ Potential secrets detected! Please review the above findings."
    echo "   - Move secrets to environment variables"
    echo "   - Use .env files (which are gitignored)"
    echo "   - Use Kubernetes secrets for deployment"
    echo "   - Use .example files for templates"
    echo ""
    echo "If these are false positives, update this script to exclude them."
    exit 1
fi

echo "✅ No secrets detected"
exit 0