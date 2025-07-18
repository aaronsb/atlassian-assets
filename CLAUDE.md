# Atlassian Assets CLI - Project Configuration

This is the project-specific CLAUDE.md for the Atlassian Assets CLI tool.

## Project Overview

A command-line tool for managing Atlassian Assets, designed as a prototype for an MCP (Model Context Protocol) interface for AI agents. The tool provides CRUD operations, dual search modes, and intelligent workflows for asset management.

## Version Management & Release Process

### Version System
- **Version Package**: `internal/version/version.go` handles version information
- **Build-time Injection**: Version, commit, and date are injected via ldflags during build
- **Version Display**: 
  - `./assets --version` - Quick version info
  - `./assets version` - Detailed version info (supports JSON output)
  - Main help shows version in header

### Release Workflow

**CRITICAL**: Always update version before tagging releases!

1. **Before Release**:
   ```bash
   # Update version in internal/version/version.go
   # Change Version = "dev" to Version = "v1.2.3"
   ```

2. **Build with Version Info**:
   ```bash
   # Development build (current default)
   go build -o assets ./cmd/assets
   
   # Release build with version injection
   VERSION="v1.2.3"
   COMMIT=$(git rev-parse HEAD)
   DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
   
   go build -ldflags "\
     -X github.com/aaronsb/atlassian-assets/internal/version.Version=$VERSION \
     -X github.com/aaronsb/atlassian-assets/internal/version.Commit=$COMMIT \
     -X github.com/aaronsb/atlassian-assets/internal/version.Date=$DATE" \
     -o assets ./cmd/assets
   ```

3. **Git Tagging**:
   ```bash
   git add .
   git commit -m "Release v1.2.3"
   git tag -a v1.2.3 -m "Release v1.2.3"
   git push origin main
   git push origin v1.2.3
   ```

4. **Verification**:
   ```bash
   ./assets --version
   ./assets version --output json
   ```

### MCP Server Version

When implementing the MCP server mode (`--mcp-server` flag), the version information should be exposed as an MCP resource:

```json
{
  "uri": "version://info",
  "name": "Version Information", 
  "description": "Current version, build, and runtime information",
  "mimeType": "application/json"
}
```

This allows MCP clients to query the server version programmatically.

## Key Features Completed

### Dual Search System
- **Simple Search**: Exact-match patterns with regex-inspired syntax
  - `=exact` - Exact match
  - `^exact$` - Exact match with anchors  
  - `*` - Wildcard (all objects)
- **AQL Search**: Advanced queries using Assets Query Language
- **Limitation**: AQL LIKE queries are non-functional, only exact matches work

### SDK Fix Implementation
- **Issue**: go-atlassian SDK v2.6.1 Filter method is broken
- **Solution**: Direct HTTP implementation bypassing SDK
- **Status**: Reported to upstream (GitHub issue #387)

### Intelligent Workflows
- **Contextual Hints**: Next-step suggestions for all commands
- **Guided Creation**: Intelligent object completion and validation
- **Error Handling**: Meaningful error messages with recovery suggestions

## Development Guidelines

### Code Structure
- Use the existing hint system for all new commands
- Follow the pattern: `cmd/assets/*.go` for CLI commands
- Place shared logic in `internal/` packages
- Add contextual hints to enhance user experience

### Testing Strategy
- Integration tests with live environment
- Help text validation for all commands
- Version information verification
- MCP compatibility testing (future)

### Documentation Standards
- Update README.md for user-facing changes
- Update SDK_FIX_DOCUMENTATION.md for technical issues
- Maintain this CLAUDE.md for project-specific workflows
- Document limitations and workarounds clearly

## Future MCP Integration

The CLI is designed to translate naturally to MCP tools:

| CLI Command | MCP Tool Name | Purpose |
|-------------|---------------|---------|
| `assets create` | `assets_create` | Create new assets |
| `assets search` | `assets_search` | Dual-mode search |
| `assets list` | `assets_list` | List assets |
| `assets get` | `assets_get` | Get asset details |
| `assets update` | `assets_update` | Update assets |
| `assets delete` | `assets_delete` | Delete assets |

### MCP Version Resource

When implementing MCP mode, ensure version information is available as a resource:

```typescript
// MCP resource for version info
{
  uri: "version://current",
  data: {
    version: "v1.2.3",
    commit: "abc123def",
    built: "2025-01-20T10:30:00Z",
    capabilities: ["search", "crud", "schema_management"]
  }
}
```

## MCP Server Integration

### Current Status
- **MCP Server**: Implemented at `/cmd/assets/mcp/main.go` using official MCP Go SDK v0.2.0
- **Registration**: Added to Claude Code environment as `atlassian-assets` server
- **Tools Available**: 13 tools (foundation + composite functions) with AI-specific guidance
- **Resources**: Version information and tool capabilities exposed

### Environment Restart Requirement
**IMPORTANT**: When MCP server changes are made, Claude Code environment must be restarted to pick up the new server configuration. 

**Process**: 
1. Make changes to MCP server code
2. Rebuild the server: `go build -o mcp-server ./cmd/assets/mcp`
3. Ask user to restart Claude Code environment
4. Test MCP server functionality after restart

### Authentication Configuration
The MCP server uses the same authentication configuration as the CLI tool:

**Option 1: .env file (recommended)**
```bash
# Create .env file in the directory where you run the MCP server
ATLASSIAN_ASSETS_EMAIL=your-email@example.com
ATLASSIAN_ASSETS_API_TOKEN=your-api-token
ATLASSIAN_ASSETS_HOST=your-instance.atlassian.net
ATLASSIAN_ASSETS_WORKSPACE_ID=your-workspace-id
```

**Option 2: Environment variables**
```bash
# Set environment variables in the shell or AI client configuration
export ATLASSIAN_ASSETS_EMAIL=your-email@example.com
export ATLASSIAN_ASSETS_API_TOKEN=your-api-token
export ATLASSIAN_ASSETS_HOST=your-instance.atlassian.net
export ATLASSIAN_ASSETS_WORKSPACE_ID=your-workspace-id
```

**For AI clients using MCP:**
Configure the MCP server with environment variables in your AI client settings:
```json
{
  "atlassian-assets": {
    "command": "/path/to/mcp-server",
    "env": {
      "ATLASSIAN_ASSETS_EMAIL": "your-email@example.com",
      "ATLASSIAN_ASSETS_API_TOKEN": "your-api-token",
      "ATLASSIAN_ASSETS_HOST": "your-instance.atlassian.net",
      "ATLASSIAN_ASSETS_WORKSPACE_ID": "your-workspace-id"
    }
  }
}
```

### MCP Server Commands
```bash
# Add server to environment
claude mcp add atlassian-assets /home/aaron/Projects/ai/mcp/jira-insights/mcp-server

# List registered servers
claude mcp list

# Remove server (if needed)
claude mcp remove atlassian-assets
```

### MCP Development Tools

**mcptools** - Go-based MCP server inspection tool (https://github.com/f/mcptools)

```bash
# Install mcptools
go install github.com/f/mcptools/cmd/mcptools@latest

# Inspect MCP server tools without restarting Claude
mcptools tools ./mcp-server
mcptools tools ./mcp-server -f json  # Get detailed JSON schemas

# List resources
mcptools resources ./mcp-server

# Test tool calls
mcptools call ./mcp-server -t assets_list_schemas

# Interactive shell
mcptools shell ./mcp-server
```

**Development Workflow**:
1. Make changes to MCP server code
2. Rebuild: `go build -o mcp-server ./cmd/assets/mcp`
3. Test with mcptools: `mcptools tools ./mcp-server`
4. For full integration testing: Ask user to restart Claude environment
5. Compare CLI vs MCP workflows

## Remember

1. **Always update version before releases**
2. **Test version display after builds**  
3. **Document breaking changes in releases**
4. **Maintain MCP compatibility in architecture**
5. **Keep SDK workarounds documented**
6. **Restart Claude Code environment after MCP server changes**