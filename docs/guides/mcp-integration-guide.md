# MCP Integration Guide

## Overview

This guide explains how to integrate the Atlassian Assets MCP server with various AI clients and automation tools, providing conversational asset management capabilities. The MCP server supports multiple deployment patterns from local stdio to hosted HTTP endpoints, making it accessible to a wide range of applications.

## Supported Integrations

### AI Clients
- **Claude Desktop** - Direct conversational asset management
- **Claude Code** - Development workflow integration
- **Custom AI Applications** - Any MCP-compatible client

### Automation Platforms
- **n8n** - Workflow automation and asset discovery
- **Zapier** - Integration with other business tools
- **Custom Scripts** - Direct MCP tool access

### Deployment Options
- **Local stdio** - Direct process communication
- **HTTP via mcp-remote** - Remote access through HTTP endpoints
- **Hosted Services** - Centralized MCP server deployment

## Quick Start

### 1. Local stdio Configuration (Claude Desktop)

Add the MCP server to your Claude Desktop configuration:

```json
{
  "mcpServers": {
    "atlassian-assets": {
      "command": "/path/to/mcp-server",
      "args": [],
      "env": {
        "ATLASSIAN_EMAIL": "your-email@example.com",
        "ATLASSIAN_HOST": "your-instance.atlassian.net",
        "ATLASSIAN_API_TOKEN": "your-api-token",
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

### 2. HTTP Remote Configuration (n8n, Zapier, etc.)

Deploy the MCP server as an HTTP endpoint using `mcp-remote`:

```bash
# Install mcp-remote globally
npm install -g mcp-remote

# Start HTTP server on port 3001
mcp-remote --stdio-command "/path/to/mcp-server" --port 3001
```

Then connect from automation platforms:

```json
{
  "mcpServers": {
    "atlassian-assets": {
      "command": "npx",
      "args": [
        "mcp-remote", 
        "http://your-server:3001/mcp",
        "--header", "Authorization: Bearer your-auth-token"
      ],
      "disabled": false,
      "transportType": "stdio"
    }
  }
}
```

### 3. Hosted Service Configuration

For production deployments, host the MCP server as a service:

```dockerfile
# Dockerfile for hosted MCP server
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o mcp-server ./cmd/assets/mcp

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/mcp-server .
EXPOSE 3001
CMD ["sh", "-c", "mcp-remote --stdio-command ./mcp-server --port 3001"]
```

### 2. Environment Variables

The MCP server supports flexible authentication:

```bash
# Primary configuration
export ATLASSIAN_EMAIL="your-email@example.com"
export ATLASSIAN_HOST="your-instance.atlassian.net"
export ATLASSIAN_API_TOKEN="your-api-token"
export ATLASSIAN_ASSETS_WORKSPACE_ID="your-workspace-id"

# Optional: Logging control
export ATLASSIAN_ASSETS_LOG_LEVEL="WARNING"  # DEBUG, INFO, WARNING, ERROR, SILENT
```

### 3. Authentication Setup

#### Generate API Token
1. Go to [Atlassian Account Settings](https://id.atlassian.com/manage-profile/security/api-tokens)
2. Create a new API token
3. Copy the token for use in configuration

#### Find Workspace ID
Use the CLI tool to discover your workspace ID:
```bash
./assets --help  # Will show workspace discovery commands
```

## Available MCP Tools

### Foundation Tools
- `assets_search` - Search for assets using exact matches or AQL queries
- `assets_list` - List all assets in a schema with pagination
- `assets_get` - Get complete details of a specific asset object
- `assets_create_object` - Create a new asset object instance
- `assets_delete` - Delete an asset object by ID
- `assets_list_schemas` - List all available schemas
- `assets_get_schema` - Get details of a specific schema
- `assets_get_object_type_attributes` - Get attributes for a specific object type
- `assets_create_object_type` - Create a new object type within a schema

### Composite Tools
- `assets_browse_schema` - Explore schema structure, object types, and asset distribution
- `assets_complete_object` - Intelligently complete asset creation with validation and defaults
- `assets_validate` - Validate object data against object type requirements
- `assets_trace_relationships` - Trace object relationships and dependencies

## AI-Specific Features

### Intelligent Guidance
The MCP server provides AI-specific enhancements:

- **Contextual Hints**: Next-step suggestions for each operation
- **Workflow Context**: Progress tracking and completion percentages
- **Error Recovery**: Meaningful error messages with suggested fixes
- **Validation Support**: Pre-creation validation with intelligent defaults

### Conversation Patterns

#### Schema Exploration
```
AI: "Let me explore your asset schemas first"
→ Uses assets_browse_schema to understand structure
→ Provides summary of available object types
→ Suggests next steps based on findings
```

#### Object Creation
```
AI: "I'll create a new marketing asset for you"
→ Uses assets_get_object_type_attributes to understand requirements
→ Uses assets_complete_object to suggest defaults
→ Uses assets_create_object to create the asset
→ Provides confirmation with object details
```

#### Relationship Mapping
```
AI: "Let me trace the relationships for this asset"
→ Uses assets_trace_relationships to map dependencies
→ Visualizes the relationship network
→ Suggests related objects or actions
```

## Configuration Options

### Logging Levels
Control verbosity for debugging:

```bash
# Detailed debugging (development)
export ATLASSIAN_ASSETS_LOG_LEVEL="DEBUG"

# Production (warnings and errors only)
export ATLASSIAN_ASSETS_LOG_LEVEL="WARNING"

# Silent (no logging)
export ATLASSIAN_ASSETS_LOG_LEVEL="SILENT"
```

### Auto-Approval Settings
Configure which tools can run without user confirmation:

```json
"autoApprove": [
  "assets_list_schemas",      // Safe: Read-only schema listing
  "assets_search",            // Safe: Search operations
  "assets_get",               // Safe: Object retrieval
  "assets_browse_schema",     // Safe: Schema exploration
  "assets_get_schema",        // Safe: Schema details
  "assets_get_object_type_attributes"  // Safe: Type inspection
]
```

**Note**: Creation and deletion operations are intentionally excluded from auto-approval for safety.

## Troubleshooting

### Common Issues

#### 1. Authentication Errors
```
Error: ATLASSIAN_EMAIL is required
```
**Solution**: Ensure all required environment variables are set correctly.

#### 2. Workspace Not Found
```
Error: Workspace ID not found
```
**Solution**: Verify your workspace ID is correct and accessible with your API token.

#### 3. Permission Denied
```
Error: 403 Forbidden
```
**Solution**: Check that your API token has Assets permissions in Atlassian.

#### 4. JSON-RPC Errors
```
Error: Unexpected token 'D', "DEBUG: ..."
```
**Solution**: Set `ATLASSIAN_ASSETS_LOG_LEVEL="WARNING"` to reduce debug output.

### Debugging Steps

1. **Test CLI First**: Verify authentication works with the CLI tool
2. **Check Logs**: Enable DEBUG logging to see detailed operations
3. **Verify Permissions**: Ensure API token has proper Assets access
4. **Test Individual Tools**: Use `mcptools` to test specific MCP tools

## Best Practices

### 1. Security
- Store API tokens securely (not in version control)
- Use environment variables or secure configuration files
- Regularly rotate API tokens

### 2. Performance
- Use pagination for large result sets
- Enable auto-approval for read-only operations
- Monitor API rate limits

### 3. Workflow Design
- Start with schema exploration before creating objects
- Use validation tools before creation
- Implement proper error handling

### 4. Integration Patterns
- Design conversational workflows that feel natural
- Provide clear feedback at each step
- Use the AI guidance features for better UX

## Advanced Configuration

### Custom Client Integration
For custom MCP clients, implement these patterns:

```python
# Example Python MCP client integration
async def create_asset_with_validation(client, object_type_id, data):
    # Step 1: Validate data
    validation_result = await client.call_tool("assets_validate", {
        "object_type_id": object_type_id,
        "data": data
    })
    
    # Step 2: Complete missing fields
    if not validation_result["validation_result"]["is_valid"]:
        completed_data = await client.call_tool("assets_complete_object", {
            "object_type_id": object_type_id,
            "data": data
        })
        data = completed_data["completion_result"]["completed_data"]
    
    # Step 3: Create object
    result = await client.call_tool("assets_create_object", {
        "object_type_id": object_type_id,
        "attributes": data
    })
    
    return result
```

### Multi-Environment Setup
Configure different environments:

```bash
# Development
export ATLASSIAN_HOST="dev-instance.atlassian.net"
export ATLASSIAN_ASSETS_WORKSPACE_ID="dev-workspace-id"

# Production
export ATLASSIAN_HOST="prod-instance.atlassian.net"
export ATLASSIAN_ASSETS_WORKSPACE_ID="prod-workspace-id"
```

## Next Steps

1. **Set up authentication** using the configuration above
2. **Test with Claude Desktop** using simple commands
3. **Explore the use cases** in the documentation
4. **Build custom workflows** using the MCP tools
5. **Contribute feedback** to improve the integration