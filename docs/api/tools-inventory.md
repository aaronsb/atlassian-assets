# Atlassian Assets CLI Tools Inventory

## Fundamental Tools (Core Operations)

### Schema Management
- `assets schema list` - List all schemas
- `assets schema get --id <schema>` - Get schema details  
- `assets schema types --schema <id>` - List object types in schema
- `assets schema create-type` - Create new object types

### Object Management
- `assets objects list --schema <id>` - List objects in schema
- `assets objects get --id <object>` - Get specific object
- `assets objects search --query <aql>` - Search with AQL
- `assets complete --type <id> --data <json>` - Intelligent object creation

### Attribute Management
- `assets attributes copy` - Copy attributes between object types
- `assets attributes trace-reference` - Trace cross-schema references

## Compositional Tools (Glue Commands)

### Browse & Discovery
- `assets browse hierarchy` - Hierarchical object type view
- `assets browse children` - Child object types
- `assets browse attributes` - Attribute details

### Summary & Analysis  
- `assets summary completion` - Object completion summary
- `assets summary schema` - Schema overview

### Universal Attribute Marketplace
- `assets extract` - Extract attributes from any source
- `assets apply` - Apply attributes to any target
- `assets catalog` - Global attribute catalog with search

### Reference Resolution
- `assets trace` - Trace references and dependencies
- `assets resolver` - Resolve names to IDs

## Status: What Works vs What's Broken

### ✅ Working Perfectly
- Object instance creation (`assets complete`)
- Object listing and search
- Attribute copying and marketplace
- Reference tracing and resolution
- Schema browsing and discovery

### ❌ Currently Broken
- **Object type creation** - "client: atlassian invalid payload"
  - Proven external issue (same code that worked before now fails)
  - Both direct API calls and SDK calls fail
  - Likely Atlassian API change or permission issue

### 🔄 Investigation Status
- Confirmed issue is NOT in our code (tested exact working commit)
- Object instances work, object types don't → API-specific issue
- Need to investigate Atlassian API documentation for changes

## Key Achievements

1. **Universal Attribute Marketplace** - Extract attributes from any source, apply to any target
2. **Cross-Schema Reference Resolution** - Maintain referential integrity across schemas  
3. **Intelligent Object Completion** - AI-friendly creation with meaningful defaults
4. **Compositional CLI Design** - Fundamental tools + glue commands eliminate complex jq patterns

## Next Steps

1. Investigate Atlassian API changes for object type creation
2. Create alternative object type creation method if needed
3. Complete regional data center experiment with Fields object type
4. Package tools for MCP translation