#!/bin/bash
# validate-docs.sh - Validate documentation structure and links

echo "📚 Validating Documentation Structure..."
echo "======================================"

# Check if all expected directories exist
EXPECTED_DIRS=(
    "use-cases"
    "guides"
    "examples"
    "architecture"
    "api"
    "screenshots"
)

for dir in "${EXPECTED_DIRS[@]}"; do
    if [[ -d "$dir" ]]; then
        echo "✓ Directory exists: $dir"
    else
        echo "✗ Missing directory: $dir"
    fi
done

echo ""
echo "📄 Checking Key Documentation Files..."
echo "======================================"

# Check if key files exist
KEY_FILES=(
    "README.md"
    "use-cases/claude-desktop-schema-creation.md"
    "use-cases/cli-automation-workflows.md"
    "guides/mcp-integration-guide.md"
    "examples/automation-scenarios.md"
    "architecture/mcp-deployment-patterns.md"
    "architecture/sdk-fix-documentation.md"
    "architecture/design.md"
    "architecture/requirements.md"
    "architecture/tasks.md"
    "architecture/tests.md"
    "api/tools-inventory.md"
)

for file in "${KEY_FILES[@]}"; do
    if [[ -f "$file" ]]; then
        echo "✓ File exists: $file"
    else
        echo "✗ Missing file: $file"
    fi
done

echo ""
echo "🔗 Checking for Broken Relative Links..."
echo "======================================"

# Simple check for common broken link patterns
find . -name "*.md" -exec grep -l "\.md" {} \; | while read -r file; do
    echo "Checking: $file"
    
    # Extract markdown links and check if referenced files exist
    grep -o '\[.*\]([^)]*\.md[^)]*)' "$file" | while read -r link; do
        # Extract the file path from the link
        filepath=$(echo "$link" | sed 's/.*(\([^)]*\)).*/\1/')
        
        # Skip external links (http/https)
        if [[ "$filepath" =~ ^https?:// ]]; then
            continue
        fi
        
        # Convert relative path to absolute path for checking
        if [[ "$filepath" =~ ^\.\. ]]; then
            # Path goes up from current directory
            check_path=$(dirname "$file")/$filepath
        elif [[ "$filepath" =~ ^\. ]]; then
            # Path relative to current directory
            check_path=$(dirname "$file")/$filepath
        else
            # Assume relative to current file's directory
            check_path=$(dirname "$file")/$filepath
        fi
        
        # Normalize the path
        check_path=$(realpath -m "$check_path" 2>/dev/null || echo "$check_path")
        
        if [[ ! -f "$check_path" ]] && [[ ! "$check_path" =~ \.\./README\.md$ ]]; then
            echo "  ⚠️  Potential broken link: $link"
            echo "      File: $file"
            echo "      Looking for: $check_path"
        fi
    done
done

echo ""
echo "📊 Documentation Statistics..."
echo "=============================="

# Count files by type
echo "Total .md files: $(find . -name "*.md" | wc -l)"
echo "Use case files: $(find ./use-cases -name "*.md" 2>/dev/null | wc -l)"
echo "Guide files: $(find ./guides -name "*.md" 2>/dev/null | wc -l)"
echo "Example files: $(find ./examples -name "*.md" 2>/dev/null | wc -l)"
echo "Architecture files: $(find ./architecture -name "*.md" 2>/dev/null | wc -l)"
echo "API files: $(find ./api -name "*.md" 2>/dev/null | wc -l)"
echo "Screenshot files: $(find ./screenshots -name "*.jpeg" -o -name "*.png" 2>/dev/null | wc -l)"

echo ""
echo "✅ Documentation validation complete!"