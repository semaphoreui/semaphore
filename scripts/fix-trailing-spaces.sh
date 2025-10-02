#!/bin/bash

# Script to fix trailing spaces in project files
echo "🔧 Fixing trailing spaces in project files..."

# Find and fix trailing spaces in common file types
find web/src -name "*.vue" -o -name "*.js" -o -name "*.ts" -o -name "*.css" -o -name "*.scss" -o -name "*.html" -o -name "*.md" | while read file; do
    if [ -f "$file" ]; then
        # Check if file has trailing spaces
        if grep -q '[[:space:]]$' "$file"; then
            echo "  Fixing: $file"
            # Remove trailing spaces
            sed -i '' 's/[[:space:]]*$//' "$file"
        fi
    fi
done

echo "✅ Trailing spaces fixed!"
