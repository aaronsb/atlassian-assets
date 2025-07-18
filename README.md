# Atlassian Assets CLI & MCP Server

A dual-purpose tool for managing Atlassian Assets through both CLI and MCP (Model Context Protocol) interfaces, enabling both human operators and AI agents to manage assets programmatically.

## Overview

This project provides CRUD operations for Atlassian Assets through two interfaces:
1. **CLI Interface**: Direct command-line tool for human operators
2. **MCP Server**: AI agent interface via Model Context Protocol

Both interfaces share the same underlying codebase, ensuring consistency and reducing maintenance overhead.

## Features

- **✅ Dual Interface**: CLI commands and MCP tools share common functionality
- **✅ 13 MCP Tools**: Complete AI agent interface with intelligent guidance
- **✅ 24+ CLI Commands**: Specialized commands for asset management and discovery
- **✅ Complete CRUD Operations**: Create, read, update, and delete assets
- **✅ Dual Search System**: Simple exact-match search and advanced AQL query search
- **✅ Schema Management**: List schemas, get schema details, and explore object types
- **✅ Pagination Support**: Handle large datasets with configurable limits
- **✅ SDK Bug Fixes**: Direct HTTP implementation bypassing broken go-atlassian SDK methods
- **✅ Contextual Hints**: Intelligent guidance system for streamlined workflows
- **✅ AI-Specific Guidance**: Context-aware suggestions for AI agents
- **✅ Version Management**: Proper semantic versioning with build-time injection
- **✅ Multiple Output Formats**: JSON, YAML, and table output

## Quick Start

### 🚀 Build Both CLI and MCP Server

```bash
# Clone and build
git clone https://github.com/aaronsb/atlassian-assets
cd atlassian-assets
./build.sh

# Or manual build
go build -o bin/assets ./cmd/assets          # CLI tool
go build -o mcp-server ./cmd/assets/mcp      # MCP server
```

### 🔧 Configuration

Create a `.env` file with your Atlassian credentials:

```bash
ATLASSIAN_ASSETS_EMAIL=your.email@company.com
ATLASSIAN_ASSETS_HOST=yourcompany.atlassian.net
ATLASSIAN_ASSETS_API_TOKEN=your-api-token
ATLASSIAN_ASSETS_WORKSPACE_ID=your-workspace-id
```

## CLI Interface

### Core CRUD Operations

```bash
# Create assets with intelligent workflow
./bin/assets create --schema computers --type laptop --guided

# List assets with pagination
./bin/assets list --schema computers --limit 100 --offset 0

# Get specific asset details  
./bin/assets get --id OBJ-123

# Update asset properties
./bin/assets update --id OBJ-123 --data '{"owner":"jane.doe"}'

# Delete assets (with safety controls)
./bin/assets delete --id OBJ-123
```

### Advanced Search & Discovery

```bash
# Simple search with pagination
./bin/assets search --simple "MacBook Pro M3" --schema 8
./bin/assets search --simple "*" --schema 8 --limit 50 --offset 50

# Advanced AQL search
./bin/assets search --query "Name = \"MacBook Pro\" AND Status = \"Active\""

# Browse and explore  
./bin/assets browse hierarchy --schema computers
./bin/assets catalog --global-objects
./bin/assets trace dependencies --object-type 65 --schema 8
```

### Schema & Metadata Management

```bash
# Schema operations
./bin/assets schema list
./bin/assets schema get --id computers  
./bin/assets schema types --schema computers

# Attribute management
./bin/assets attributes --schema computers
./bin/assets extract --schema computers --format csv
```

## MCP Server Interface

### Available MCP Tools

The MCP server provides 13 tools with AI-specific guidance:

| MCP Tool | Purpose | CLI Equivalent |
|----------|---------|----------------|
| `assets_list_schemas` | List all available schemas | `schema list` |
| `assets_search` | Search for assets with dual modes | `search` |
| `assets_list` | List objects with pagination | `list` |
| `assets_get` | Get detailed object information | `get` |
| `assets_create_object` | Create new asset instances | `create` |
| `assets_delete` | Delete objects with validation | `delete` |
| `assets_get_schema` | Get schema details | `schema get` |
| `assets_create_object_type` | Create new object types | `schema create-type` |
| `assets_get_object_type_attributes` | Get object type structure | `attributes` |
| `assets_browse_schema` | Intelligent schema exploration | `browse hierarchy` |
| `assets_validate` | Object validation against requirements | `validate` |
| `assets_complete_object` | Intelligent object completion | `complete` |
| `assets_trace_relationships` | Trace object dependencies | `trace dependencies` |

### Claude Desktop Configuration

Add to your `~/.config/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "atlassian-assets": {
      "command": "/path/to/mcp-server",
      "args": [],
      "env": {
        "ATLASSIAN_ASSETS_EMAIL": "your.email@company.com",
        "ATLASSIAN_ASSETS_HOST": "yourcompany.atlassian.net",
        "ATLASSIAN_ASSETS_API_TOKEN": "your-api-token",
        "ATLASSIAN_ASSETS_WORKSPACE_ID": "your-workspace-id"
      },
      "disabled": false,
      "transportType": "stdio",
      "autoApprove": [
        "assets_list_schemas",
        "assets_search", 
        "assets_get",
        "assets_browse_schema",
        "assets_get_schema",
        "assets_get_object_type_attributes"
      ]
    }
  }
}
```

### Claude Code Configuration

Add the MCP server to your Claude Code environment:

```bash
# Add server to Claude Code
claude mcp add atlassian-assets /path/to/mcp-server

# List registered servers
claude mcp list

# Remove server (if needed)
claude mcp remove atlassian-assets
```

### Other MCP Clients

For any JSON-based MCP client, use this configuration pattern:

```json
{
  "atlassian-assets": {
    "command": "/path/to/mcp-server",
    "env": {
      "ATLASSIAN_ASSETS_EMAIL": "your.email@company.com",
      "ATLASSIAN_ASSETS_HOST": "yourcompany.atlassian.net", 
      "ATLASSIAN_ASSETS_API_TOKEN": "your-api-token",
      "ATLASSIAN_ASSETS_WORKSPACE_ID": "your-workspace-id"
    }
  }
}
```

## Architecture

### Dual Interface Design

```
┌─────────────────┐    ┌─────────────────┐
│   CLI Commands  │    │   MCP Tools     │
│   (Human UI)    │    │   (AI Interface)│
└─────────┬───────┘    └─────────┬───────┘
          │                      │
          └──────────┬───────────┘
                     │
         ┌───────────▼───────────┐
         │   Common Package      │
         │  ┌─────────────────┐  │
         │  │   Foundation    │  │  # Core CRUD operations
         │  │   (objects.go)  │  │
         │  └─────────────────┘  │
         │  ┌─────────────────┐  │
         │  │   Composite     │  │  # Intelligent workflows
         │  │   (browse.go)   │  │
         │  └─────────────────┘  │
         └───────────────────────┘
                     │
         ┌───────────▼───────────┐
         │   Atlassian Assets    │
         │      API Client       │
         └───────────────────────┘
```

### Project Structure

```
├── cmd/assets/                  # CLI entry point
│   ├── main.go                  # CLI commands
│   ├── common/                  # Shared functionality
│   │   ├── foundation/          # Core CRUD operations
│   │   ├── composite/           # Intelligent workflows
│   │   └── types.go             # Common interfaces
│   └── mcp/                     # MCP server
│       └── main.go              # MCP tool handlers
├── internal/
│   ├── client/                  # Atlassian API client
│   ├── config/                  # Configuration management
│   └── hints/                   # AI & CLI guidance systems
├── reference/claude/            # Claude development guidelines
└── .claude-github/              # GitHub integration files
```

## Development Status

### ✅ Production Features (v1.0.0)

- **Dual Interface**: CLI and MCP server with shared codebase
- **Complete CRUD**: All asset management operations
- **Advanced Search**: Dual search modes with full pagination
- **Schema Management**: Full schema and object type operations
- **AI Integration**: 13 MCP tools with context-aware guidance
- **SDK Bug Fixes**: Direct HTTP implementation bypassing broken SDK methods
- **Intelligent Workflows**: Contextual hints and guided operations
- **Version Management**: Semantic versioning with build-time injection
- **Authentication**: Environment variable and .env file support
- **Live Testing**: Validated against real Atlassian Assets environment

### 🚀 Future Enhancements

See our [GitHub Issues](https://github.com/aaronsb/atlassian-assets/issues) for planned improvements:

- **Response Size Management**: Pagination and filtering recommendations (#6)
- **Named Object Resolution**: Human-readable object references (#7)
- **Enhanced Error Handling**: Structured error recovery (#8)
- **Automated Testing**: Unit and integration test suites (#11)
- **Logging & Monitoring**: Production-grade observability (#10)

## Important Notes

### SDK Issue and Fix

**⚠️ Critical Bug Fixed**: The go-atlassian SDK v2.6.1 has a broken `Object.Filter()` method that makes AQL searches non-functional. This project includes a **direct HTTP implementation** that bypasses the broken SDK method.

**Impact**: Without this fix, search and list operations would return empty results.

**Solution**: See `SDK_FIX_DOCUMENTATION.md` for complete technical details.

**Upstream Issue**: https://github.com/ctreminiom/go-atlassian/issues/387

### MCP vs CLI Differences

| Feature | CLI Interface | MCP Interface |
|---------|---------------|---------------|
| **User Type** | Human operators | AI agents |
| **Output** | Human-readable with next steps | Structured data with AI guidance |
| **Error Handling** | User-friendly messages | Structured error responses |
| **Workflow Hints** | CLI-specific suggestions | AI-specific context and recommendations |
| **Response Size** | Full details | Optimized for context management |
| **Authentication** | .env file or flags | Environment variables only |

## Contributing

This project follows structured development with GitHub integration:

1. **Requirements**: GitHub Issues with `requirement` label
2. **Tasks**: GitHub Milestones for major features
3. **Sub-tasks**: GitHub Issues with `task` label
4. **Guidelines**: See `reference/claude/USER_SCOPE_CLAUDE.md`

## License

MIT License

## Author

Aaron Bockelie