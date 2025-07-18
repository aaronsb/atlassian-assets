# CLI Automation Workflows

## Overview

The Atlassian Assets CLI provides powerful command-line access to your asset management system, enabling both human operators and traditional automation systems to efficiently manage assets. This document explores practical use cases and workflow patterns for the CLI tool.

## Human-Driven Workflows

### Asset Discovery and Exploration

#### Quick Schema Overview
```bash
# List all available schemas
./assets list-schemas

# Get detailed information about a specific schema
./assets get-schema --schema-id 1

# Browse schema structure with object types and sample data
./assets browse --schema 1
```

**Use Case**: New team member needs to understand the asset structure
- Quickly explore available schemas and object types
- Understand data organization and relationships
- Identify relevant assets for their project

#### Asset Search and Filtering
```bash
# Simple search for marketing assets
./assets search --schema 1 --simple "marketing" --output json

# Advanced AQL query for specific criteria
./assets search --schema 1 --query "Status = 'Active' AND Owner = 'John Doe'"

# List all assets in a schema with pagination
./assets list --schema 1 --limit 20 --offset 0
```

**Use Case**: Project manager needs to find all assets related to a campaign
- Search across multiple criteria
- Export results for reporting
- Identify asset gaps or duplicates

### Asset Management Operations

#### Creating Asset Hierarchies
```bash
# Create a new object type for project assets
./assets create-object-type --schema 1 --name "Project Asset" --description "Assets specific to project management"

# Create individual assets
./assets create-object --object-type-id 142 --attributes '{"name": "Q4 Campaign Plan", "status": "Active"}'
```

**Use Case**: Setting up a new project's asset structure
- Create project-specific object types
- Initialize template assets
- Establish naming conventions

#### Bulk Asset Operations
```bash
# Export all assets for backup
./assets list --schema 1 --output json > backup_$(date +%Y%m%d).json

# Search and export specific asset types
./assets search --schema 1 --query "objectType = 'Marketing Material'" --output json > marketing_assets.json
```

**Use Case**: Regular asset maintenance and reporting
- Create backups of asset data
- Generate compliance reports
- Audit asset usage patterns

### Quality Assurance Workflows

#### Asset Validation and Cleanup
```bash
# Get detailed information about an asset
./assets get --id OBJ-123 --output json

# Validate asset structure (using composite commands)
./assets validate --object-type-id 142 --data '{"name": "Test Asset"}'

# Check for orphaned or incomplete assets
./assets search --schema 1 --query "Name = ''" --output json
```

**Use Case**: Data quality manager performs regular audits
- Identify assets with missing required fields
- Validate data consistency
- Clean up orphaned or duplicate assets

## Traditional Automation Scenarios

### Schema Migration and Replication

#### Cross-Environment Migration
```bash
#!/bin/bash
# migrate_schema.sh - Migrate schema from dev to prod

SOURCE_WORKSPACE="dev-workspace-id"
TARGET_WORKSPACE="prod-workspace-id"
SCHEMA_ID="$1"

if [[ -z "$SCHEMA_ID" ]]; then
    echo "Usage: $0 <schema_id>"
    exit 1
fi

# Export source schema structure
echo "Exporting schema structure from development..."
./assets get-schema --schema-id $SCHEMA_ID --output json > source_schema.json

# Export object types
./assets browse --schema $SCHEMA_ID --output json > source_types.json

# Parse and recreate in target environment
export ATLASSIAN_ASSETS_WORKSPACE_ID="$TARGET_WORKSPACE"

# Create object types in dependency order
jq -r '.data.object_types[] | @json' source_types.json | while read -r type; do
    name=$(echo "$type" | jq -r '.name')
    description=$(echo "$type" | jq -r '.description // ""')
    parent=$(echo "$type" | jq -r '.parent // ""')
    
    echo "Creating object type: $name"
    
    if [[ "$parent" != "" ]]; then
        ./assets create-object-type --schema $SCHEMA_ID --name "$name" --description "$description" --parent "$parent"
    else
        ./assets create-object-type --schema $SCHEMA_ID --name "$name" --description "$description"
    fi
done

echo "Schema migration completed"
```

#### Schema Cloning for Testing
```bash
#!/bin/bash
# clone_schema.sh - Clone production schema for testing

PROD_SCHEMA="$1"
TEST_SCHEMA="$2"

echo "Cloning schema $PROD_SCHEMA to $TEST_SCHEMA..."

# Export production schema structure
./assets browse --schema $PROD_SCHEMA --output json > prod_schema.json

# Create object types in test schema
jq -r '.data.object_types[] | select(.parent == null) | @json' prod_schema.json | while read -r type; do
    name=$(echo "$type" | jq -r '.name')
    description="TEST: $(echo "$type" | jq -r '.description // ""')"
    
    echo "Creating root type: $name"
    ./assets create-object-type --schema $TEST_SCHEMA --name "$name" --description "$description"
done

# Create child types (second pass)
jq -r '.data.object_types[] | select(.parent != null) | @json' prod_schema.json | while read -r type; do
    name=$(echo "$type" | jq -r '.name')
    description="TEST: $(echo "$type" | jq -r '.description // ""')"
    parent=$(echo "$type" | jq -r '.parent')
    
    echo "Creating child type: $name"
    ./assets create-object-type --schema $TEST_SCHEMA --name "$name" --description "$description" --parent "$parent"
done

# Create sample data (10% of production)
./assets list --schema $PROD_SCHEMA --limit 50 --output json | jq -r '.data.objects[] | @json' | while read -r obj; do
    name="TEST: $(echo "$obj" | jq -r '.name')"
    type_id=$(echo "$obj" | jq -r '.objectType.id')
    
    # Create test version of object
    ./assets create-object --object-type-id $type_id --attributes "{\"name\": \"$name\", \"status\": \"Test\"}"
done

echo "Schema cloning completed"
```

### Consultancy Template Management

#### Template Library System
```bash
#!/bin/bash
# template_manager.sh - Manage reusable schema templates

TEMPLATE_DIR="/opt/atlassian-assets/templates"
COMMAND="$1"
TEMPLATE_NAME="$2"

case "$COMMAND" in
    "create")
        # Create template from existing schema
        SCHEMA_ID="$3"
        
        echo "Creating template '$TEMPLATE_NAME' from schema $SCHEMA_ID..."
        
        # Export complete schema structure
        ./assets browse --schema $SCHEMA_ID --output json > "$TEMPLATE_DIR/$TEMPLATE_NAME.json"
        
        # Create template metadata
        cat > "$TEMPLATE_DIR/$TEMPLATE_NAME.meta.json" << EOF
{
    "name": "$TEMPLATE_NAME",
    "description": "Template created from schema $SCHEMA_ID",
    "created": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "version": "1.0.0",
    "industry": "general",
    "use_cases": ["asset_management", "project_tracking"]
}
EOF
        
        echo "Template created: $TEMPLATE_DIR/$TEMPLATE_NAME.json"
        ;;
        
    "list")
        echo "Available templates:"
        for template in "$TEMPLATE_DIR"/*.json; do
            if [[ -f "$template" ]] && [[ ! "$template" =~ \.meta\.json$ ]]; then
                base=$(basename "$template" .json)
                description=$(jq -r '.description // "No description"' "$TEMPLATE_DIR/$base.meta.json" 2>/dev/null)
                echo "  - $base: $description"
            fi
        done
        ;;
        
    "apply")
        # Apply template to new schema
        TARGET_SCHEMA="$3"
        
        if [[ ! -f "$TEMPLATE_DIR/$TEMPLATE_NAME.json" ]]; then
            echo "Template '$TEMPLATE_NAME' not found"
            exit 1
        fi
        
        echo "Applying template '$TEMPLATE_NAME' to schema $TARGET_SCHEMA..."
        
        # Read template and create object types
        jq -r '.data.object_types[] | @json' "$TEMPLATE_DIR/$TEMPLATE_NAME.json" | while read -r type; do
            name=$(echo "$type" | jq -r '.name')
            description=$(echo "$type" | jq -r '.description // ""')
            parent=$(echo "$type" | jq -r '.parent // ""')
            
            echo "Creating object type: $name"
            
            if [[ "$parent" != "" ]]; then
                ./assets create-object-type --schema $TARGET_SCHEMA --name "$name" --description "$description" --parent "$parent"
            else
                ./assets create-object-type --schema $TARGET_SCHEMA --name "$name" --description "$description"
            fi
        done
        
        echo "Template applied successfully"
        ;;
        
    "validate")
        # Validate template structure
        if [[ ! -f "$TEMPLATE_DIR/$TEMPLATE_NAME.json" ]]; then
            echo "Template '$TEMPLATE_NAME' not found"
            exit 1
        fi
        
        echo "Validating template '$TEMPLATE_NAME'..."
        
        # Check JSON structure
        if ! jq empty "$TEMPLATE_DIR/$TEMPLATE_NAME.json" 2>/dev/null; then
            echo "ERROR: Invalid JSON structure"
            exit 1
        fi
        
        # Check required fields
        required_fields=(".data.object_types" ".data.schema")
        for field in "${required_fields[@]}"; do
            if [[ $(jq -r "$field" "$TEMPLATE_DIR/$TEMPLATE_NAME.json") == "null" ]]; then
                echo "ERROR: Missing required field: $field"
                exit 1
            fi
        done
        
        echo "Template validation passed"
        ;;
        
    *)
        echo "Usage: $0 {create|list|apply|validate} [template_name] [schema_id]"
        echo "  create <name> <schema_id>  - Create template from existing schema"
        echo "  list                       - List available templates"
        echo "  apply <name> <schema_id>   - Apply template to schema"
        echo "  validate <name>            - Validate template structure"
        exit 1
        ;;
esac
```

#### Industry-Specific Template Deployment
```bash
#!/bin/bash
# deploy_industry_template.sh - Deploy pre-configured industry templates

INDUSTRY="$1"
CLIENT_NAME="$2"
SCHEMA_ID="$3"

case "$INDUSTRY" in
    "manufacturing")
        echo "Deploying manufacturing template for $CLIENT_NAME..."
        
        # Apply manufacturing-specific object types
        ./template_manager.sh apply "manufacturing_base" $SCHEMA_ID
        
        # Create client-specific customizations
        ./assets create-object-type --schema $SCHEMA_ID --name "Production Line" --description "Manufacturing production line for $CLIENT_NAME"
        ./assets create-object-type --schema $SCHEMA_ID --name "Quality Control Point" --description "QC checkpoint for $CLIENT_NAME"
        
        # Create sample assets
        ./assets create-object --object-type-id $(./assets browse --schema $SCHEMA_ID | jq -r '.data.object_types[] | select(.name=="Production Line") | .id') \
            --attributes '{"name": "Line A - Primary Production", "status": "Active", "capacity": "1000 units/day"}'
        ;;
        
    "healthcare")
        echo "Deploying healthcare template for $CLIENT_NAME..."
        
        ./template_manager.sh apply "healthcare_base" $SCHEMA_ID
        
        # Healthcare-specific object types
        ./assets create-object-type --schema $SCHEMA_ID --name "Medical Device" --description "Medical equipment for $CLIENT_NAME"
        ./assets create-object-type --schema $SCHEMA_ID --name "Treatment Protocol" --description "Clinical protocols for $CLIENT_NAME"
        ;;
        
    "retail")
        echo "Deploying retail template for $CLIENT_NAME..."
        
        ./template_manager.sh apply "retail_base" $SCHEMA_ID
        
        # Retail-specific object types
        ./assets create-object-type --schema $SCHEMA_ID --name "Product SKU" --description "Retail products for $CLIENT_NAME"
        ./assets create-object-type --schema $SCHEMA_ID --name "Store Location" --description "Physical locations for $CLIENT_NAME"
        ;;
        
    *)
        echo "Available industries: manufacturing, healthcare, retail"
        exit 1
        ;;
esac

echo "Industry template deployment completed for $CLIENT_NAME"
```

### Schema Introspection and Analysis

#### Schema Complexity Analysis
```bash
#!/bin/bash
# analyze_schema.sh - Analyze schema complexity and structure

SCHEMA_ID="$1"
ANALYSIS_FILE="schema_analysis_$(date +%Y%m%d).json"

echo "Analyzing schema $SCHEMA_ID..."

# Get complete schema structure
./assets browse --schema $SCHEMA_ID --output json > temp_schema.json

# Extract metrics
TOTAL_TYPES=$(jq '.data.object_type_count' temp_schema.json)
TOTAL_OBJECTS=$(jq '.data.total_objects' temp_schema.json)
SAMPLE_SIZE=$(jq '.data.sample_size' temp_schema.json)

# Calculate hierarchy depth
MAX_DEPTH=0
jq -r '.data.object_types[] | @json' temp_schema.json | while read -r type; do
    name=$(echo "$type" | jq -r '.name')
    parent=$(echo "$type" | jq -r '.parent // ""')
    
    # Calculate depth for this type
    depth=0
    current_parent="$parent"
    while [[ "$current_parent" != "" ]]; do
        depth=$((depth + 1))
        current_parent=$(jq -r ".data.object_types[] | select(.name==\"$current_parent\") | .parent // \"\"" temp_schema.json)
    done
    
    if [[ $depth -gt $MAX_DEPTH ]]; then
        MAX_DEPTH=$depth
    fi
done

# Create analysis report
cat > $ANALYSIS_FILE << EOF
{
    "schema_id": "$SCHEMA_ID",
    "analysis_date": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "metrics": {
        "total_object_types": $TOTAL_TYPES,
        "total_objects": $TOTAL_OBJECTS,
        "max_hierarchy_depth": $MAX_DEPTH,
        "average_objects_per_type": $(echo "scale=2; $TOTAL_OBJECTS / $TOTAL_TYPES" | bc),
        "sample_coverage": $(echo "scale=2; $SAMPLE_SIZE / $TOTAL_OBJECTS * 100" | bc)
    },
    "complexity_rating": "$(if [[ $TOTAL_TYPES -gt 20 ]]; then echo "high"; elif [[ $TOTAL_TYPES -gt 10 ]]; then echo "medium"; else echo "low"; fi)",
    "recommendations": []
}
EOF

# Add recommendations based on analysis
if [[ $MAX_DEPTH -gt 5 ]]; then
    jq '.recommendations += ["Consider flattening deep hierarchies (depth > 5)"]' $ANALYSIS_FILE > temp.json && mv temp.json $ANALYSIS_FILE
fi

if [[ $TOTAL_TYPES -gt 50 ]]; then
    jq '.recommendations += ["Large number of object types may indicate over-categorization"]' $ANALYSIS_FILE > temp.json && mv temp.json $ANALYSIS_FILE
fi

echo "Analysis completed: $ANALYSIS_FILE"
rm temp_schema.json
```

#### Schema Comparison Tool
```bash
#!/bin/bash
# compare_schemas.sh - Compare two schemas for differences

SCHEMA_A="$1"
SCHEMA_B="$2"
COMPARISON_FILE="schema_comparison_$(date +%Y%m%d).json"

echo "Comparing schema $SCHEMA_A with schema $SCHEMA_B..."

# Export both schemas
./assets browse --schema $SCHEMA_A --output json > schema_a.json
./assets browse --schema $SCHEMA_B --output json > schema_b.json

# Extract object type names
jq -r '.data.object_types[].name' schema_a.json | sort > types_a.txt
jq -r '.data.object_types[].name' schema_b.json | sort > types_b.txt

# Find differences
ONLY_IN_A=$(comm -23 types_a.txt types_b.txt)
ONLY_IN_B=$(comm -13 types_a.txt types_b.txt)
COMMON=$(comm -12 types_a.txt types_b.txt)

# Create comparison report
cat > $COMPARISON_FILE << EOF
{
    "comparison_date": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "schema_a": "$SCHEMA_A",
    "schema_b": "$SCHEMA_B",
    "summary": {
        "total_types_a": $(wc -l < types_a.txt),
        "total_types_b": $(wc -l < types_b.txt),
        "common_types": $(echo "$COMMON" | wc -l),
        "unique_to_a": $(echo "$ONLY_IN_A" | wc -l),
        "unique_to_b": $(echo "$ONLY_IN_B" | wc -l)
    },
    "differences": {
        "only_in_schema_a": [$(echo "$ONLY_IN_A" | jq -R . | tr '\n' ',' | sed 's/,$//')],
        "only_in_schema_b": [$(echo "$ONLY_IN_B" | jq -R . | tr '\n' ',' | sed 's/,$//')],
        "common_types": [$(echo "$COMMON" | jq -R . | tr '\n' ',' | sed 's/,$//')],
        "similarity_score": $(echo "scale=2; $(echo "$COMMON" | wc -l) / $(cat types_a.txt types_b.txt | sort -u | wc -l) * 100" | bc)
    }
}
EOF

echo "Comparison completed: $COMPARISON_FILE"
rm schema_a.json schema_b.json types_a.txt types_b.txt
```

### Scheduled Maintenance Scripts

#### Daily Asset Backup
```bash
#!/bin/bash
# daily_backup.sh

DATE=$(date +%Y%m%d)
BACKUP_DIR="/backups/assets"
SCHEMAS="1 2 3"

mkdir -p "$BACKUP_DIR"

for schema in $SCHEMAS; do
    echo "Backing up schema $schema..."
    ./assets list --schema $schema --output json > "$BACKUP_DIR/schema_${schema}_${DATE}.json"
    
    # Validate backup
    if [[ -s "$BACKUP_DIR/schema_${schema}_${DATE}.json" ]]; then
        echo "Schema $schema backup successful"
    else
        echo "ERROR: Schema $schema backup failed"
        exit 1
    fi
done

echo "All backups completed successfully"
```

#### Weekly Asset Audit
```bash
#!/bin/bash
# weekly_audit.sh

REPORT_FILE="asset_audit_$(date +%Y%m%d).txt"

echo "Asset Audit Report - $(date)" > $REPORT_FILE
echo "=================================" >> $REPORT_FILE

# Check for assets missing names
echo "Assets missing names:" >> $REPORT_FILE
./assets search --schema 1 --query "Name = ''" --output json | jq -r '.data.objects[] | .key' >> $REPORT_FILE

# Check for recent assets
echo -e "\nRecently created assets:" >> $REPORT_FILE
./assets search --schema 1 --query "Created > '$(date -d '7 days ago' +%Y-%m-%d)'" --output json | jq -r '.data.objects[] | "\(.key): \(.name)"' >> $REPORT_FILE

# Email report
mail -s "Weekly Asset Audit Report" admin@company.com < $REPORT_FILE
```

### Integration with CI/CD Pipelines

#### Asset Validation in Build Pipeline
```yaml
# .github/workflows/asset-validation.yml
name: Asset Validation
on: [push, pull_request]

jobs:
  validate-assets:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Setup Assets CLI
        run: |
          curl -L https://github.com/your-org/atlassian-assets/releases/latest/download/assets-linux -o assets
          chmod +x assets
          
      - name: Validate Project Assets
        env:
          ATLASSIAN_EMAIL: ${{ secrets.ATLASSIAN_EMAIL }}
          ATLASSIAN_API_TOKEN: ${{ secrets.ATLASSIAN_API_TOKEN }}
          ATLASSIAN_HOST: ${{ secrets.ATLASSIAN_HOST }}
          ATLASSIAN_ASSETS_WORKSPACE_ID: ${{ secrets.WORKSPACE_ID }}
        run: |
          # Check if project assets exist
          ./assets search --schema 1 --simple "${{ github.event.repository.name }}" --output json > project_assets.json
          
          # Validate asset count
          asset_count=$(jq '.data.total' project_assets.json)
          if [[ $asset_count -eq 0 ]]; then
            echo "Warning: No assets found for project ${{ github.event.repository.name }}"
          else
            echo "Found $asset_count assets for project"
          fi
```

#### Automated Asset Creation
```bash
#!/bin/bash
# create_release_assets.sh

PROJECT_NAME="$1"
VERSION="$2"
RELEASE_DATE="$(date +%Y-%m-%d)"

if [[ -z "$PROJECT_NAME" || -z "$VERSION" ]]; then
    echo "Usage: $0 <project_name> <version>"
    exit 1
fi

# Create release documentation asset
./assets create-object --object-type-id 145 --attributes "{
    \"name\": \"Release Documentation - $PROJECT_NAME v$VERSION\",
    \"status\": \"Active\",
    \"version\": \"$VERSION\",
    \"release_date\": \"$RELEASE_DATE\"
}" --output json > release_doc.json

# Create release notes asset
./assets create-object --object-type-id 146 --attributes "{
    \"name\": \"Release Notes - $PROJECT_NAME v$VERSION\",
    \"status\": \"Active\",
    \"version\": \"$VERSION\",
    \"content\": \"See CHANGELOG.md for details\"
}" --output json > release_notes.json

echo "Release assets created successfully"
```

### System Integration Workflows

#### LDAP User Synchronization
```bash
#!/bin/bash
# sync_users_to_assets.sh

# Get users from LDAP
ldapsearch -x -H ldap://ldap.company.com -D "cn=admin,dc=company,dc=com" -W \
    -b "ou=users,dc=company,dc=com" "(objectClass=person)" mail displayName > users.ldif

# Process users and create assets
while IFS= read -r line; do
    if [[ $line =~ ^mail:\ (.+) ]]; then
        email="${BASH_REMATCH[1]}"
    elif [[ $line =~ ^displayName:\ (.+) ]]; then
        name="${BASH_REMATCH[1]}"
        
        # Create user asset
        ./assets create-object --object-type-id 150 --attributes "{
            \"name\": \"$name\",
            \"email\": \"$email\",
            \"status\": \"Active\",
            \"department\": \"Unknown\"
        }" --output json
        
        echo "Created asset for user: $name"
    fi
done < users.ldif
```

#### Database Synchronization
```bash
#!/bin/bash
# sync_database_assets.sh

# Export database table to CSV
mysql -u user -p database -e "SELECT id, name, description, created_at FROM products" --batch --raw > products.csv

# Skip header and process each row
tail -n +2 products.csv | while IFS=$'\t' read -r id name description created_at; do
    # Search for existing asset
    existing=$(./assets search --schema 1 --query "Name = '$name'" --output json | jq -r '.data.total')
    
    if [[ $existing -eq 0 ]]; then
        # Create new asset
        ./assets create-object --object-type-id 151 --attributes "{
            \"name\": \"$name\",
            \"description\": \"$description\",
            \"external_id\": \"$id\",
            \"status\": \"Active\"
        }" --output json
        
        echo "Created asset: $name"
    else
        echo "Asset already exists: $name"
    fi
done
```

## Monitoring and Alerting

### Asset Health Monitoring
```bash
#!/bin/bash
# asset_health_check.sh

CRITICAL_ASSETS=(
    "OBJ-001"
    "OBJ-002"
    "OBJ-003"
)

for asset in "${CRITICAL_ASSETS[@]}"; do
    result=$(./assets get --id "$asset" --output json)
    
    if [[ $? -eq 0 ]]; then
        name=$(echo "$result" | jq -r '.data.object.name')
        status=$(echo "$result" | jq -r '.data.object.status')
        
        if [[ "$status" != "Active" ]]; then
            echo "ALERT: Critical asset $asset ($name) is not active: $status"
            # Send alert to monitoring system
            curl -X POST https://monitoring.company.com/alert \
                -H "Content-Type: application/json" \
                -d "{\"message\": \"Critical asset $asset is not active\", \"severity\": \"high\"}"
        fi
    else
        echo "ERROR: Could not retrieve asset $asset"
    fi
done
```

### Performance Monitoring
```bash
#!/bin/bash
# performance_monitor.sh

SCHEMAS="1 2 3"
METRICS_FILE="/tmp/asset_metrics.json"

echo "{" > $METRICS_FILE
echo "  \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"," >> $METRICS_FILE
echo "  \"schemas\": [" >> $METRICS_FILE

first=true
for schema in $SCHEMAS; do
    if [[ $first == true ]]; then
        first=false
    else
        echo "    ," >> $METRICS_FILE
    fi
    
    # Time the list operation
    start_time=$(date +%s.%N)
    result=$(./assets list --schema $schema --limit 1 --output json)
    end_time=$(date +%s.%N)
    
    response_time=$(echo "$end_time - $start_time" | bc)
    total_assets=$(echo "$result" | jq -r '.data.total')
    
    echo "    {" >> $METRICS_FILE
    echo "      \"schema_id\": \"$schema\"," >> $METRICS_FILE
    echo "      \"total_assets\": $total_assets," >> $METRICS_FILE
    echo "      \"response_time\": $response_time" >> $METRICS_FILE
    echo "    }" >> $METRICS_FILE
done

echo "  ]" >> $METRICS_FILE
echo "}" >> $METRICS_FILE

# Send metrics to monitoring system
curl -X POST https://metrics.company.com/assets \
    -H "Content-Type: application/json" \
    -d @$METRICS_FILE
```

## Reporting and Analytics

### Asset Usage Reports
```bash
#!/bin/bash
# generate_usage_report.sh

REPORT_DATE=$(date +%Y-%m-%d)
REPORT_FILE="asset_usage_report_$REPORT_DATE.html"

cat > $REPORT_FILE << EOF
<!DOCTYPE html>
<html>
<head>
    <title>Asset Usage Report - $REPORT_DATE</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <h1>Asset Usage Report - $REPORT_DATE</h1>
EOF

# Generate schema statistics
echo "    <h2>Schema Overview</h2>" >> $REPORT_FILE
echo "    <table>" >> $REPORT_FILE
echo "        <tr><th>Schema</th><th>Total Assets</th><th>Active Assets</th></tr>" >> $REPORT_FILE

for schema in 1 2 3; do
    total=$(./assets list --schema $schema --output json | jq -r '.data.total')
    active=$(./assets search --schema $schema --query "Status = 'Active'" --output json | jq -r '.data.total')
    
    echo "        <tr><td>Schema $schema</td><td>$total</td><td>$active</td></tr>" >> $REPORT_FILE
done

echo "    </table>" >> $REPORT_FILE
echo "</body>" >> $REPORT_FILE
echo "</html>" >> $REPORT_FILE

echo "Report generated: $REPORT_FILE"
```

### Compliance Reporting
```bash
#!/bin/bash
# compliance_report.sh

COMPLIANCE_CHECKS=(
    "Name != ''"
    "Status IN ('Active', 'Inactive')"
    "Created IS NOT NULL"
)

echo "Compliance Report - $(date)"
echo "================================="

for check in "${COMPLIANCE_CHECKS[@]}"; do
    echo "Checking: $check"
    
    # Find non-compliant assets
    violations=$(./assets search --schema 1 --query "NOT ($check)" --output json | jq -r '.data.total')
    
    if [[ $violations -eq 0 ]]; then
        echo "  ✓ PASS: No violations found"
    else
        echo "  ✗ FAIL: $violations violations found"
        
        # List violating assets
        ./assets search --schema 1 --query "NOT ($check)" --output json | \
            jq -r '.data.objects[] | "    - \(.key): \(.name)"'
    fi
    
    echo
done
```

## Best Practices

### Script Organization
1. **Modular Design**: Break complex operations into smaller, reusable scripts
2. **Error Handling**: Always check return codes and handle failures gracefully
3. **Logging**: Implement comprehensive logging for audit trails
4. **Configuration**: Use environment variables for sensitive data

### Performance Optimization
1. **Pagination**: Use appropriate limits and offsets for large datasets
2. **Caching**: Cache frequently accessed data to reduce API calls
3. **Parallel Processing**: Use background jobs for independent operations
4. **Rate Limiting**: Implement delays to avoid overwhelming the API

### Security Considerations
1. **Credential Management**: Use secure methods for storing API tokens
2. **Input Validation**: Validate all inputs to prevent injection attacks
3. **Access Control**: Implement proper permissions for script execution
4. **Audit Logging**: Log all operations for security monitoring

### Maintenance and Documentation
1. **Version Control**: Store all scripts in version control
2. **Documentation**: Document script purpose, parameters, and usage
3. **Testing**: Test scripts in non-production environments
4. **Monitoring**: Implement monitoring for critical automation

## Conclusion

The Atlassian Assets CLI provides a powerful foundation for both human-driven workflows and traditional automation. By combining the flexibility of command-line tools with the structure of asset management, organizations can create efficient, reliable, and maintainable asset operations that scale with their needs.

The key to success is starting with simple use cases and gradually building more sophisticated automation as requirements evolve.