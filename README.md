# Atlassian Assets CLI

A command-line tool for managing Atlassian Assets, designed as a prototype for an MCP (Model Context Protocol) interface for AI agents.

## Overview

This project provides CRUD operations for Atlassian Assets through a clean CLI interface. The CLI is designed to translate naturally to MCP tools, enabling AI agents to manage assets programmatically.

## Features

- **✅ Complete CRUD Operations**: Create, read, update, and delete assets with intelligent workflows
- **✅ Dual Search System**: Simple exact-match search and advanced AQL query search with full pagination  
- **✅ Schema Management**: List schemas, get schema details, and explore object types
- **✅ Advanced Workflows**: 24+ specialized commands for asset management, discovery, and automation
- **✅ Pagination Support**: Handle large datasets with `--limit` and `--offset` on search and list operations
- **✅ SDK Bug Fixes**: Direct HTTP implementation bypassing broken go-atlassian SDK methods
- **✅ Contextual Hints**: Intelligent guidance system for streamlined workflows
- **✅ Version Management**: Proper semantic versioning with build-time injection
- **✅ Multiple Output Formats**: JSON, YAML, and table output
- **✅ MCP-Ready Design**: CLI commands map directly to future MCP tools

## Quick Start

### 🚀 Super Simple Build (No Go Experience Required)

```bash
# Clone and build in one go
git clone https://github.com/aaronsb/atlassian-assets
cd atlassian-assets
./build.sh
```

The build script will:
- ✅ Check if Go is installed (and guide you if not)
- ✅ Download dependencies 
- ✅ Run tests
- ✅ Build the binary
- ✅ Offer to install to your system

### 🛠️ Advanced Build (For Developers)

```bash
# Using Makefile (more options)
make help           # See all build options
make build          # Interactive build with options
make test           # Run all tests
make build-release VERSION=v1.0.0  # Release build

# Manual build
go build -o bin/assets ./cmd/assets
```

## Configuration

Create a `.env` file with your Atlassian credentials:

```bash
ATLASSIAN_EMAIL=your.email@company.com
ATLASSIAN_HOST=https://yourcompany.atlassian.net
ATLASSIAN_API_TOKEN=your-api-token
ATLASSIAN_ASSETS_WORKSPACE_ID=your-workspace-id
```

## Complete Command Reference

### 🔧 Core CRUD Operations

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
./bin/assets remove --attribute "old-field" --from OBJ-123
```

### 🔍 Advanced Search & Discovery

```bash
# Simple search with pagination
./bin/assets search --simple "MacBook Pro M3" --schema 8
./bin/assets search --simple "*" --schema 8 --limit 50 --offset 50

# Advanced AQL search
./bin/assets search --query "Name = \"MacBook Pro\" AND Status = \"Active\""

# Browse and explore  
./bin/assets browse --schema computers
./bin/assets catalog --global-objects
./bin/assets trace --object OBJ-123 --depth 2
```

### 📊 Schema & Metadata Management

```bash
# Schema operations
./bin/assets schema list
./bin/assets schema get --id computers  
./bin/assets schema types --schema computers

# Attribute management
./bin/assets attributes --schema computers
./bin/assets copy-attributes --from laptop --to workstation
./bin/assets extract --schema computers --format csv
```

### 🤖 Intelligent Workflows

```bash
# Smart completion and validation
./bin/assets complete --schema computers --type laptop
./bin/assets validate --id OBJ-123 --strict

# Bulk operations and automation
./bin/assets apply --attribute "Status:Active" --to-schema computers
./bin/assets workflows --list
./bin/assets summary --schema computers --analytics
```

### ⚙️ System & Configuration

```bash
# Configuration management
./bin/assets config show
./bin/assets config test
./bin/assets version --output json

# Environment testing
./bin/assets test --connection
./bin/assets resolve --name "MacBook Pro" --to-id
```

## Architecture

### CLI → MCP Translation

The CLI is designed for natural translation to MCP tools:

| CLI Command | MCP Tool Name | Purpose |
|-------------|---------------|---------|
| `assets create` | `assets_create` | Create new assets |
| `assets list` | `assets_list` | List assets |
| `assets get` | `assets_get` | Get asset details |
| `assets update` | `assets_update` | Update assets |
| `assets delete` | `assets_delete` | Delete assets |
| `assets search` | `assets_search` | Search assets |

### Project Structure

```
├── cmd/assets/          # CLI entry point and commands
├── internal/
│   ├── client/          # Atlassian API client wrapper
│   └── config/          # Configuration management
├── pkg/assets/          # Public API (future MCP interface)
└── requirements.md      # Requirements tracking
```

## Development Status

### ✅ Production Ready Features (v0.1.0)

- **Complete CLI Framework**: 24+ specialized commands with Cobra
- **Advanced Search & Pagination**: Dual search modes with full pagination support
- **Intelligent Workflows**: Contextual hints and guided asset management
- **SDK Bug Fixes**: Direct HTTP implementation bypassing broken go-atlassian methods
- **Comprehensive CRUD**: Create, Read, Update, Delete with safety controls
- **Schema Management**: Full schema and object type operations
- **Version Management**: Semantic versioning with build-time injection
- **Live Testing**: Validated against real Atlassian Assets environment
- **Build Automation**: Simple and advanced build systems for all skill levels

### 🚀 Next Phase (v0.2.0) 

- **MCP Server Mode**: `--mcp-server` flag for stdio MCP protocol
- **AI Agent Interface**: All CLI commands as MCP tools 
- **Automated Workflows**: AI-driven asset management and discovery

## Important Notes

### SDK Issue and Fix

**⚠️ Critical Bug Fixed**: The go-atlassian SDK v2.6.1 has a broken `Object.Filter()` method that makes AQL searches non-functional. This project includes a **direct HTTP implementation** that bypasses the broken SDK method.

**Impact**: Without this fix, `assets search` and `assets list` commands would return empty results despite objects existing.

**Solution**: See `SDK_FIX_DOCUMENTATION.md` for complete technical details, including:
- Root cause analysis proving SDK is broken
- Working direct HTTP replacement implementation  
- Validation results showing fix success
- GitHub issue #387 filed with upstream maintainers

**Upstream Issue**: https://github.com/ctreminiom/go-atlassian/issues/387

**Status**: Both search and list operations now work perfectly and return complete object data.

## Contributing

This project follows a structured development approach with requirements tracking:

1. See `requirements.md` for user stories and acceptance criteria
2. See `design.md` for architecture decisions
3. See `tasks.md` for implementation plan

## Author

Aaron Bockelie

## License

MIT License