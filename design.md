# Design - Atlassian Assets CLI Toolkit

## Architecture Overview

### Primary Goal: MCP Agent Interface
The ultimate goal is to create an MCP (Model Context Protocol) interface that allows AI agents to manage Atlassian Assets through CRUD operations. This interface will enable agents to integrate asset management into automated workflows.

### Development Strategy: CLI-First Approach
We'll prototype the functionality as a CLI first because:
1. **Parallel Design Patterns**: CLI interfaces and MCP interfaces share similar command structures and parameter handling
2. **Easier Testing**: CLI tools are simpler to test and debug during development
3. **Natural Translation**: Well-designed CLI commands translate directly to MCP tool definitions
4. **Incremental Complexity**: Start simple with CLI, then add MCP layer

### Architecture Flow
```
CLI Prototype → MCP Interface → AI Agent Integration
     ↓              ↓                ↓
  Human User    AI Agent       Automated Workflows
```

## Technology Stack

### Recommended Options
- **SDK Library**: Top candidates identified
- **CLI Framework**: TBD based on selected SDK language
- **Configuration**: TBD
- **Authentication**: OAuth 2.0 or API tokens

### SDK Library Options Evaluated

#### Option 1: go-atlassian (Go)
**Pros:**
- Actively maintained (1,295+ commits)
- Explicit Atlassian Assets API support
- Strong documentation at docs.go-atlassian.io
- OAuth 2.0 authentication
- Good community adoption (167 stars)

**Cons:**
- Smaller community compared to Python options
- Go learning curve if not familiar

#### Option 2: jira.js (JavaScript/Node.js) 
**Pros:**
- Actively maintained (242 commits)
- Modern TypeScript support
- Node.js 20.0.0+ support
- Good documentation (459 stars)
- Tree-shaking support

**Cons:**
- No explicit Assets API support mentioned
- May require custom Assets API implementation

#### Option 3: atlassian-python-api (Python)
**Pros:**
- Most mature library (1,836+ commits, 1.5k stars)
- Large contributor base (325 contributors)
- Latest release: 4.0.4 (May 2025)
- Support for JSM (may include Assets)

**Cons:**
- Assets API support unclear from documentation
- May require verification/custom implementation

## Key Design Decisions

### Decision 1: SDK Selection
**Status**: 🔄 In Progress  
**Context**: Need to evaluate available open-source Atlassian Assets SDK options  
**Options**: 
1. **go-atlassian (Go)**: Explicit Assets API support, actively maintained
2. **jira.js (JavaScript)**: Modern but unclear Assets support  
3. **atlassian-python-api (Python)**: Most mature but Assets support unclear
**Decision**: ✅ **Selected: go-atlassian**  
**Rationale**: Only library with explicit Assets API support and active maintenance

### Decision 2: CLI → MCP Translation Strategy
**Status**: ✅ **Complete**  
**Context**: Need to design CLI commands that translate naturally to MCP tools  
**Approach**: 
- CLI commands become MCP tool names
- CLI flags become MCP tool parameters
- CLI output becomes MCP tool responses
**Decision**: ✅ **1:1 mapping strategy with consistent parameter patterns**

## CLI Command Structure Design

### Core CRUD Operations
```bash
# CREATE
assets create --schema <schema-id> --type <object-type> --data <json-data>
assets create --schema computers --type laptop --data '{"name":"MacBook Pro","owner":"john"}'

# READ
assets list --schema <schema-id> [--type <object-type>] [--filter <aql-query>]
assets get --id <object-id>
assets search --query <aql-query>

# UPDATE  
assets update --id <object-id> --data <json-data>
assets update --id OBJ-123 --data '{"owner":"jane"}'

# DELETE
assets delete --id <object-id>
assets delete --id OBJ-123
```

### Schema Management
```bash
# Schema operations
assets schema list
assets schema get --id <schema-id>
assets schema types --schema <schema-id>
```

### Configuration & Auth
```bash
# Configuration
assets config set --url <instance-url> --token <api-token>
assets config show
assets config test  # Test connection

# Authentication
assets auth login --url <instance-url>
assets auth status
```

## MCP Translation Mapping

### CLI → MCP Tool Translation
| CLI Command | MCP Tool Name | MCP Parameters | MCP Response |
|-------------|---------------|----------------|--------------|
| `assets create` | `assets_create` | `schema_id`, `object_type`, `data` | `{id, created_at, ...}` |
| `assets list` | `assets_list` | `schema_id`, `object_type?`, `filter?` | `{objects: [...]}` |
| `assets get` | `assets_get` | `object_id` | `{object: {...}}` |
| `assets update` | `assets_update` | `object_id`, `data` | `{updated_at, ...}` |
| `assets delete` | `assets_delete` | `object_id` | `{deleted: true}` |
| `assets search` | `assets_search` | `query` | `{results: [...]}` |

### Benefits of This Design
1. **Natural Translation**: Each CLI command maps directly to an MCP tool
2. **Consistent Parameters**: CLI flags become MCP tool parameters
3. **Structured Output**: CLI JSON output becomes MCP tool responses
4. **Composable**: Operations can be chained in both CLI and MCP contexts

## Configuration & Authentication Design

### Configuration Management
```yaml
# ~/.config/atlassian-assets/config.yaml
default_profile: production

profiles:
  production:
    url: "https://company.atlassian.net"
    auth_type: "token"
    token: "ATATT3xFfGF0..."
    workspace_id: "12345678-1234-1234-1234-123456789012"
  
  staging:
    url: "https://company-staging.atlassian.net"
    auth_type: "oauth"
    client_id: "..."
    client_secret: "..."
    workspace_id: "87654321-4321-4321-4321-210987654321"
```

### Authentication Methods
1. **API Tokens** (Recommended for CLI)
   - Simple username + API token
   - Easy to configure and manage
   - Works well for automation

2. **OAuth 2.0** (Future MCP consideration)
   - More secure for user-facing applications
   - Supports token refresh
   - Better for long-running services

### Configuration Commands
```bash
# Profile management
assets config profiles                    # List profiles
assets config set-profile <name>         # Switch active profile
assets config create-profile <name>      # Create new profile

# Profile configuration
assets config set --profile <name> --url <url> --token <token>
assets config show --profile <name>
assets config test --profile <name>      # Test connection

# Workspace discovery
assets config discover-workspace --profile <name>  # Auto-detect workspace ID
```

### Environment Variables
```bash
ATLASSIAN_ASSETS_URL=https://company.atlassian.net
ATLASSIAN_ASSETS_TOKEN=ATATT3xFfGF0...
ATLASSIAN_ASSETS_WORKSPACE_ID=12345678-1234-1234-1234-123456789012
ATLASSIAN_ASSETS_PROFILE=production
```

## Open Questions

1. ✅ What open-source SDK libraries are available for Atlassian Assets?
2. ✅ Which programming languages have the best SDK support?
3. What authentication methods are supported? (OAuth 2.0, API tokens confirmed)
4. What are the rate limits and API constraints?
5. How should we handle configuration management?
6. Should we verify Assets API support in Python library?

## References

- [Atlassian Assets REST API Guide](https://developer.atlassian.com/cloud/assets/assets-rest-api-guide/workflow/)
- [go-atlassian Documentation](https://docs.go-atlassian.io)
- [jira.js GitHub Repository](https://github.com/MrRefactoring/jira.js)
- [atlassian-python-api GitHub Repository](https://github.com/atlassian-api/atlassian-python-api)

---

## Changelog
- 2025-07-17: Initial design document structure created