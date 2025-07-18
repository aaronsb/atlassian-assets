# MCP Deployment Patterns

## Overview

The Atlassian Assets MCP server supports multiple deployment patterns to meet different organizational needs, from local development to enterprise-scale production deployments.

## Deployment Options

### 1. Local stdio (Development)
**Use Case**: Development, testing, single-user scenarios
**Transport**: Direct process communication

```bash
# Direct execution
./mcp-server

# With environment variables
ATLASSIAN_ASSETS_LOG_LEVEL=DEBUG ./mcp-server
```

**Pros**:
- Simple setup and debugging
- No network dependencies
- Full control over environment

**Cons**:
- Single client only
- No remote access
- Process lifecycle management

### 2. HTTP via mcp-remote (Multi-Client)
**Use Case**: Team collaboration, automation platforms, remote access
**Transport**: HTTP with JSON-RPC over WebSocket/HTTP

```bash
# Install mcp-remote
npm install -g mcp-remote

# Start HTTP server
mcp-remote --stdio-command "/path/to/mcp-server" --port 3001

# With authentication
mcp-remote --stdio-command "/path/to/mcp-server" --port 3001 --auth-token "your-token"
```

**Pros**:
- Multiple concurrent clients
- Remote access capability
- Standard HTTP integration
- Authentication support

**Cons**:
- Additional dependency (Node.js)
- Network latency
- HTTP overhead

### 3. Hosted Service (Production)
**Use Case**: Enterprise deployment, high availability, centralized management
**Transport**: HTTP service with load balancing

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o mcp-server ./cmd/assets/mcp

FROM node:18-alpine
RUN npm install -g mcp-remote
WORKDIR /app
COPY --from=builder /app/mcp-server .
EXPOSE 3001
CMD ["mcp-remote", "--stdio-command", "./mcp-server", "--port", "3001"]
```

**Pros**:
- High availability
- Scalable architecture
- Centralized configuration
- Monitoring and logging

**Cons**:
- Complex deployment
- Infrastructure overhead
- Security considerations

## Client Integration Patterns

### AI Clients

#### Claude Desktop
```json
{
  "mcpServers": {
    "atlassian-assets": {
      "command": "/path/to/mcp-server",
      "transportType": "stdio",
      "env": { "ATLASSIAN_EMAIL": "..." }
    }
  }
}
```

#### Claude Code
```bash
# Register MCP server
claude mcp add atlassian-assets /path/to/mcp-server

# Use in development workflow
claude mcp list
```

#### Custom AI Applications
```python
import asyncio
from mcp import MCPClient

async def main():
    client = MCPClient("stdio", ["/path/to/mcp-server"])
    await client.connect()
    
    # Use MCP tools
    result = await client.call_tool("assets_search", {
        "schema": "1",
        "simple": "marketing"
    })
    
    print(result)
    await client.disconnect()

asyncio.run(main())
```

### Automation Platforms

#### n8n Workflow
```json
{
  "nodes": [
    {
      "name": "Assets MCP",
      "type": "n8n-nodes-base.mcp",
      "parameters": {
        "server": "http://localhost:3001/mcp",
        "tool": "assets_search",
        "arguments": {
          "schema": "{{ $json.schema_id }}",
          "simple": "{{ $json.search_term }}"
        }
      }
    }
  ]
}
```

#### Zapier Integration
```javascript
// Zapier custom integration
const response = await fetch('http://your-mcp-server:3001/mcp', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer your-token'
  },
  body: JSON.stringify({
    jsonrpc: '2.0',
    method: 'tools/call',
    params: {
      name: 'assets_search',
      arguments: {
        schema: inputData.schema,
        simple: inputData.searchTerm
      }
    },
    id: 1
  })
});
```

### Custom Script Integration

#### Python Script
```python
import subprocess
import json

def call_mcp_tool(tool_name, arguments):
    """Call MCP tool via subprocess"""
    cmd = ['/path/to/mcp-server']
    process = subprocess.Popen(
        cmd,
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True
    )
    
    request = {
        "jsonrpc": "2.0",
        "method": "tools/call",
        "params": {
            "name": tool_name,
            "arguments": arguments
        },
        "id": 1
    }
    
    stdout, stderr = process.communicate(json.dumps(request))
    return json.loads(stdout)

# Usage
result = call_mcp_tool("assets_search", {
    "schema": "1",
    "simple": "marketing"
})
```

#### Shell Script
```bash
#!/bin/bash
# Simple MCP tool wrapper

MCP_SERVER="/path/to/mcp-server"
TOOL_NAME="$1"
ARGUMENTS="$2"

REQUEST=$(cat <<EOF
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "$TOOL_NAME",
    "arguments": $ARGUMENTS
  },
  "id": 1
}
EOF
)

echo "$REQUEST" | $MCP_SERVER
```

## Security Considerations

### Authentication
- API token management
- Environment variable security
- Network access controls
- SSL/TLS encryption for HTTP deployments

### Authorization
- Role-based access control
- Tool-level permissions
- Resource-level restrictions
- Audit logging

### Network Security
- Firewall configuration
- VPN access for remote deployments
- Network segmentation
- DDoS protection

## Monitoring and Observability

### Logging
```bash
# Enable detailed logging
export ATLASSIAN_ASSETS_LOG_LEVEL=DEBUG

# Structured logging output
export ATLASSIAN_ASSETS_LOG_FORMAT=json
```

### Metrics
- Request/response times
- Error rates
- Tool usage statistics
- Resource utilization

### Health Checks
```bash
# Simple health check
curl -X POST http://localhost:3001/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"initialize","params":{},"id":1}'
```

## Scaling Patterns

### Horizontal Scaling
- Multiple MCP server instances
- Load balancing
- Session affinity considerations
- Shared configuration

### Vertical Scaling
- Resource optimization
- Memory usage patterns
- CPU utilization
- I/O performance

### Caching Strategies
- Response caching
- Asset metadata caching
- Connection pooling
- Rate limiting

## Best Practices

### Development
1. Start with local stdio for development
2. Use environment variables for configuration
3. Implement proper error handling
4. Test with multiple clients

### Production
1. Use hosted service deployment
2. Implement comprehensive monitoring
3. Set up automated backups
4. Plan for disaster recovery

### Security
1. Secure API token storage
2. Implement network security
3. Regular security audits
4. Access control reviews

### Performance
1. Monitor resource usage
2. Optimize query patterns
3. Implement caching where appropriate
4. Plan for peak load scenarios

## Migration Strategies

### Development to Production
1. Environment parity
2. Configuration management
3. Deployment automation
4. Rollback procedures

### Platform Migration
1. Client compatibility testing
2. Feature parity verification
3. Performance comparison
4. User training and documentation

## Future Considerations

### Emerging Patterns
- Container orchestration (Kubernetes)
- Serverless deployments
- Edge computing
- Multi-cloud strategies

### Protocol Evolution
- MCP specification updates
- Transport layer improvements
- Security enhancements
- Performance optimizations

## Conclusion

The flexibility of MCP deployment patterns enables organizations to choose the approach that best fits their needs, from simple development scenarios to complex enterprise deployments. The key is to start with the simplest pattern that meets your requirements and evolve as your needs grow.