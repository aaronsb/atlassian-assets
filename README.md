# Atlassian Assets CLI

A command-line tool for managing Atlassian Assets, designed as a prototype for an MCP (Model Context Protocol) interface for AI agents.

## Overview

This project provides CRUD operations for Atlassian Assets through a clean CLI interface. The CLI is designed to translate naturally to MCP tools, enabling AI agents to manage assets programmatically.

## Features

- **Complete CRUD Operations**: Create, read, update, and delete assets
- **Schema Management**: List schemas, get schema details, and explore object types
- **Configuration Management**: Profile-based configuration with environment variable support
- **Multiple Output Formats**: JSON, YAML, and table output
- **MCP-Ready Design**: CLI commands map directly to future MCP tools

## Installation

```bash
git clone https://github.com/aaronsb/atlassian-assets
cd atlassian-assets
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

## Usage

### Basic Commands

```bash
# Show help
./bin/assets --help

# Show current configuration
./bin/assets config show

# Test connection
./bin/assets config test

# List assets in a schema
./bin/assets list --schema computers

# Create a new asset
./bin/assets create --schema computers --type laptop --data '{"name":"MacBook Pro","owner":"john.doe"}'

# Get asset details
./bin/assets get --id OBJ-123

# Update an asset
./bin/assets update --id OBJ-123 --data '{"owner":"jane.doe"}'

# Delete an asset
./bin/assets delete --id OBJ-123

# Search assets
./bin/assets search --query "Name like 'MacBook%'"
```

### Schema Management

```bash
# List all schemas
./bin/assets schema list

# Get schema details
./bin/assets schema get --id computers

# List object types in schema
./bin/assets schema types --schema computers
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

- ✅ CLI Framework (Cobra)
- ✅ Configuration Management
- ✅ go-atlassian SDK Integration
- ✅ Command Structure Design
- 🔄 API Implementation (in progress)
- ⏳ Workspace ID Discovery
- ⏳ MCP Interface Layer

## Contributing

This project follows a structured development approach with requirements tracking:

1. See `requirements.md` for user stories and acceptance criteria
2. See `design.md` for architecture decisions
3. See `tasks.md` for implementation plan

## Author

Aaron Bockelie

## License

MIT License