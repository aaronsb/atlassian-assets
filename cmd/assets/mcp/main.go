package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/modelcontextprotocol/go-sdk/jsonschema"
	"github.com/aaronsb/atlassian-assets/cmd/assets/common"
	"github.com/aaronsb/atlassian-assets/cmd/assets/common/foundation"
	"github.com/aaronsb/atlassian-assets/cmd/assets/common/composite"
	"github.com/aaronsb/atlassian-assets/internal/client"
	"github.com/aaronsb/atlassian-assets/internal/config"
	"github.com/aaronsb/atlassian-assets/internal/hints"
	"github.com/aaronsb/atlassian-assets/internal/logger"
	"github.com/aaronsb/atlassian-assets/internal/version"
)

// Global client for tool handlers
var assetsClient common.ClientInterface

// Initialize MCP server
func NewMCPServer() (*mcp.Server, error) {
	// Load configuration (from .env file or environment variables)
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate critical configuration
	if cfg.Email == "" || cfg.APIToken == "" || cfg.Host == "" {
		return nil, fmt.Errorf("missing required configuration: email, api_token, and host must be set in .env file or environment variables")
	}

	// Create client
	client, err := client.NewAssetsClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	assetsClient = client

	// Create MCP server
	versionInfo := version.GetInfo()
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "atlassian-assets-mcp",
		Version: versionInfo.Version,
	}, nil)

	// Register tools and resources
	registerTools(server)
	registerResources(server)

	return server, nil
}

// Register all available tools
func registerTools(server *mcp.Server) {
	// Foundation tools
	mcp.AddTool(server, &mcp.Tool{
		Name: "assets_search",
		Description: "Search for assets using exact matches or AQL queries",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"simple": {Type: "string", Description: "Simple search term for name/key matching (exact matches only)"},
				"query": {Type: "string", Description: "AQL query for advanced searching"},
				"schema": {Type: "string", Description: "Schema ID or name to limit search scope (required)"},
				"limit": {Type: "integer", Description: "Maximum number of results (1-1000, default: 50)"},
				"offset": {Type: "integer", Description: "Number of results to skip for pagination (default: 0)"},
			},
			Required: []string{"schema"},
		},
	}, handleSearchTool)
	
	mcp.AddTool(server, &mcp.Tool{
		Name: "assets_list",
		Description: "List all assets in a schema with pagination",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"schema": {Type: "string", Description: "Schema ID or name (required)"},
				"limit": {Type: "integer", Description: "Maximum number of results (1-1000, default: 50)"},
				"offset": {Type: "integer", Description: "Number of results to skip for pagination (default: 0)"},
			},
			Required: []string{"schema"},
		},
	}, handleListTool)
	
	mcp.AddTool(server, &mcp.Tool{
		Name: "assets_get",
		Description: "Get complete details of a specific asset object",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"id": {Type: "string", Description: "Unique asset object ID (like OBJ-123)"},
			},
			Required: []string{"id"},
		},
	}, handleGetTool)
	
	mcp.AddTool(server, &mcp.Tool{
		Name: "assets_create_object",
		Description: "Create a new asset object instance",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"object_type_id": {Type: "string", Description: "Object type ID where the asset will be created"},
				"attributes": {Type: "object", Description: "Asset attributes as key-value pairs"},
			},
			Required: []string{"object_type_id", "attributes"},
		},
	}, handleCreateObjectTool)
	
	mcp.AddTool(server, &mcp.Tool{
		Name: "assets_delete",
		Description: "Delete an asset object by ID",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"id": {Type: "string", Description: "Unique asset object ID to delete"},
			},
			Required: []string{"id"},
		},
	}, handleDeleteTool)
	
	mcp.AddTool(server, &mcp.Tool{
		Name: "assets_list_schemas",
		Description: "List all available schemas in the workspace",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{},
		},
	}, handleListSchemasTool)
	
	mcp.AddTool(server, &mcp.Tool{
		Name: "assets_get_schema",
		Description: "Get details of a specific schema",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"schema_id": {Type: "string", Description: "Schema ID or name"},
			},
			Required: []string{"schema_id"},
		},
	}, handleGetSchemaTool)
	
	mcp.AddTool(server, &mcp.Tool{
		Name: "assets_create_object_type",
		Description: "Create a new object type within a schema",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"schema": {Type: "string", Description: "Schema ID or name (required)"},
				"name": {Type: "string", Description: "Object type name (required)"},
				"description": {Type: "string", Description: "Description of the object type"},
				"parent": {Type: "string", Description: "Parent object type ID"},
				"icon": {Type: "string", Description: "Icon ID"},
			},
			Required: []string{"schema", "name"},
		},
	}, handleCreateObjectTypeTool)
	
	mcp.AddTool(server, &mcp.Tool{
		Name: "assets_get_object_type_attributes",
		Description: "Get attributes for a specific object type",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"object_type_id": {Type: "string", Description: "Object type ID"},
			},
			Required: []string{"object_type_id"},
		},
	}, handleGetObjectTypeAttributesTool)
	
	// Composite tools
	mcp.AddTool(server, &mcp.Tool{
		Name: "assets_browse_schema",
		Description: "Explore schema structure, object types, and asset distribution",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"schema_id": {Type: "string", Description: "Schema ID or name to explore"},
			},
			Required: []string{"schema_id"},
		},
	}, handleBrowseSchemaTool)
	
	mcp.AddTool(server, &mcp.Tool{
		Name: "assets_validate",
		Description: "Validate object data against object type requirements",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"object_type_id": {Type: "string", Description: "Object type ID"},
				"data": {Type: "object", Description: "Object data to validate"},
			},
			Required: []string{"object_type_id", "data"},
		},
	}, handleValidateTool)
	
	mcp.AddTool(server, &mcp.Tool{
		Name: "assets_complete_object",
		Description: "Intelligently complete asset creation with validation and defaults",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"object_type_id": {Type: "string", Description: "Object type ID where the asset will be created"},
				"data": {Type: "object", Description: "Partial asset data - will be enhanced with intelligent defaults"},
			},
			Required: []string{"object_type_id", "data"},
		},
	}, handleCompleteObjectTool)
	
	mcp.AddTool(server, &mcp.Tool{
		Name: "assets_trace_relationships",
		Description: "Trace object relationships and dependencies",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"object_id": {Type: "string", Description: "Object ID to trace relationships from"},
				"depth": {Type: "integer", Description: "Relationship depth (default: 1)"},
			},
			Required: []string{"object_id"},
		},
	}, handleTraceRelationshipsTool)
}

// Register resources
func registerResources(server *mcp.Server) {
	// Version resource
	server.AddResource(&mcp.Resource{
		Name:     "Version Information",
		MIMEType: "application/json",
		URI:      "version://current",
	}, handleVersionResource)
	
	// Tool capabilities resource
	server.AddResource(&mcp.Resource{
		Name:     "Tool Capabilities",
		MIMEType: "application/json",
		URI:      "capabilities://tools",
	}, handleCapabilitiesResource)
}


// Resource handlers
func handleVersionResource(ctx context.Context, ss *mcp.ServerSession, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
	versionInfo := version.GetInfo()
	
	content, err := json.Marshal(versionInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal version info: %w", err)
	}
	
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      params.URI,
				MIMEType: "application/json",
				Text:     string(content),
			},
		},
	}, nil
}

func handleCapabilitiesResource(ctx context.Context, ss *mcp.ServerSession, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
	semantics, err := hints.LoadAIToolSemantics()
	if err != nil {
		return nil, fmt.Errorf("failed to load AI tool semantics: %w", err)
	}
	
	content, err := json.Marshal(semantics)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tool semantics: %w", err)
	}
	
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      params.URI,
				MIMEType: "application/json",
				Text:     string(content),
			},
		},
	}, nil
}

// Tool handlers
func handleSearchTool(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]interface{}]) (*mcp.CallToolResult, error) {
	// Use the parsed arguments directly
	args := params.Arguments
	
	searchParams := common.SearchParams{
		Simple: getStringParam(args, "simple", ""),
		Query:  getStringParam(args, "query", ""),
		Schema: getStringParam(args, "schema", ""),
		Limit:  getIntParam(args, "limit", 50),
		Offset: getIntParam(args, "offset", 0),
	}
	
	response, err := foundation.SearchObjects(assetsClient, searchParams)
	if err != nil {
		return nil, err
	}
	
	// Add AI guidance
	context := buildToolContext(args, response)
	responseWithGuidance := addAIGuidance(response, "assets_search", context)
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: formatResponse(responseWithGuidance)},
		},
	}, nil
}

func handleListTool(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]interface{}]) (*mcp.CallToolResult, error) {
	// Use the parsed arguments directly
	args := params.Arguments
	
	listParams := common.ListParams{
		Schema: getStringParam(args, "schema", ""),
		Limit:  getIntParam(args, "limit", 50),
		Offset: getIntParam(args, "offset", 0),
	}
	
	response, err := foundation.ListObjects(assetsClient, listParams)
	if err != nil {
		return nil, err
	}
	
	// Add AI guidance
	context := buildToolContext(args, response)
	responseWithGuidance := addAIGuidance(response, "assets_list", context)
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: formatResponse(responseWithGuidance)},
		},
	}, nil
}

func handleGetTool(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]interface{}]) (*mcp.CallToolResult, error) {
	// Use the parsed arguments directly
	args := params.Arguments
	
	getParams := common.GetParams{
		ID: getStringParam(args, "id", ""),
	}
	
	response, err := foundation.GetObject(assetsClient, getParams)
	if err != nil {
		return nil, err
	}
	
	// Add AI guidance
	context := buildToolContext(args, response)
	responseWithGuidance := addAIGuidance(response, "assets_get", context)
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: formatResponse(responseWithGuidance)},
		},
	}, nil
}

func handleCreateObjectTool(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]interface{}]) (*mcp.CallToolResult, error) {
	// Use the parsed arguments directly
	args := params.Arguments
	
	createParams := common.CreateObjectParams{
		ObjectTypeID: getStringParam(args, "object_type_id", ""),
		Attributes:   getMapParam(args, "attributes"),
	}
	
	response, err := foundation.CreateObject(assetsClient, createParams)
	if err != nil {
		return nil, err
	}
	
	// Add AI guidance
	context := buildToolContext(args, response)
	responseWithGuidance := addAIGuidance(response, "assets_create_object", context)
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: formatResponse(responseWithGuidance)},
		},
	}, nil
}

func handleDeleteTool(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]interface{}]) (*mcp.CallToolResult, error) {
	// Use the parsed arguments directly
	args := params.Arguments
	
	deleteParams := common.DeleteObjectParams{
		ID: getStringParam(args, "id", ""),
	}
	
	response, err := foundation.DeleteObject(assetsClient, deleteParams)
	if err != nil {
		return nil, err
	}
	
	// Add AI guidance
	context := buildToolContext(args, response)
	responseWithGuidance := addAIGuidance(response, "assets_delete", context)
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: formatResponse(responseWithGuidance)},
		},
	}, nil
}

func handleListSchemasTool(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]interface{}]) (*mcp.CallToolResult, error) {
	// Use the parsed arguments directly
	args := params.Arguments
	
	response, err := foundation.ListSchemas(assetsClient)
	if err != nil {
		return nil, err
	}
	
	// Add AI guidance
	context := buildToolContext(args, response)
	responseWithGuidance := addAIGuidance(response, "assets_list_schemas", context)
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: formatResponse(responseWithGuidance)},
		},
	}, nil
}

func handleGetSchemaTool(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]interface{}]) (*mcp.CallToolResult, error) {
	// Use the parsed arguments directly
	args := params.Arguments
	
	schemaParams := common.GetSchemaParams{
		SchemaID: getStringParam(args, "schema_id", ""),
	}
	
	response, err := foundation.GetSchema(assetsClient, schemaParams)
	if err != nil {
		return nil, err
	}
	
	// Add AI guidance
	context := buildToolContext(args, response)
	responseWithGuidance := addAIGuidance(response, "assets_get_schema", context)
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: formatResponse(responseWithGuidance)},
		},
	}, nil
}

func handleCreateObjectTypeTool(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]interface{}]) (*mcp.CallToolResult, error) {
	// Use the parsed arguments directly
	args := params.Arguments
	
	createParams := common.CreateObjectTypeParams{
		Schema:      getStringParam(args, "schema", ""),
		Name:        getStringParam(args, "name", ""),
		Description: getStringParam(args, "description", ""),
		Icon:        getStringParam(args, "icon", ""),
	}
	
	if parent := getStringParam(args, "parent", ""); parent != "" {
		createParams.Parent = &parent
	}
	
	response, err := foundation.CreateObjectType(assetsClient, createParams)
	if err != nil {
		return nil, err
	}
	
	// Add AI guidance
	context := buildToolContext(args, response)
	responseWithGuidance := addAIGuidance(response, "assets_create_object_type", context)
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: formatResponse(responseWithGuidance)},
		},
	}, nil
}

func handleGetObjectTypeAttributesTool(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]interface{}]) (*mcp.CallToolResult, error) {
	// Use the parsed arguments directly
	args := params.Arguments
	
	attrParams := common.GetObjectTypeAttributesParams{
		ObjectTypeID: getStringParam(args, "object_type_id", ""),
	}
	
	response, err := foundation.GetObjectTypeAttributes(assetsClient, attrParams)
	if err != nil {
		return nil, err
	}
	
	// Add AI guidance
	context := buildToolContext(args, response)
	responseWithGuidance := addAIGuidance(response, "assets_get_object_type_attributes", context)
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: formatResponse(responseWithGuidance)},
		},
	}, nil
}

func handleBrowseSchemaTool(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]interface{}]) (*mcp.CallToolResult, error) {
	// Use the parsed arguments directly
	args := params.Arguments
	
	browseParams := common.BrowseSchemaParams{
		SchemaID: getStringParam(args, "schema_id", ""),
	}
	
	response, err := composite.BrowseSchema(assetsClient, browseParams)
	if err != nil {
		return nil, err
	}
	
	// Add AI guidance
	context := buildToolContext(args, response)
	responseWithGuidance := addAIGuidance(response, "assets_browse_schema", context)
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: formatResponse(responseWithGuidance)},
		},
	}, nil
}

func handleValidateTool(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]interface{}]) (*mcp.CallToolResult, error) {
	// Use the parsed arguments directly
	args := params.Arguments
	
	validateParams := common.ValidateObjectParams{
		ObjectTypeID: getStringParam(args, "object_type_id", ""),
		Data:         getMapParam(args, "data"),
	}
	
	response, err := composite.ValidateObject(assetsClient, validateParams)
	if err != nil {
		return nil, err
	}
	
	// Add AI guidance
	context := buildToolContext(args, response)
	responseWithGuidance := addAIGuidance(response, "assets_validate", context)
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: formatResponse(responseWithGuidance)},
		},
	}, nil
}

func handleCompleteObjectTool(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]interface{}]) (*mcp.CallToolResult, error) {
	// Use the parsed arguments directly
	args := params.Arguments
	
	completeParams := common.CompleteObjectParams{
		ObjectTypeID: getStringParam(args, "object_type_id", ""),
		Data:         getMapParam(args, "data"),
	}
	
	response, err := composite.CompleteObject(assetsClient, completeParams)
	if err != nil {
		return nil, err
	}
	
	// Add AI guidance
	context := buildToolContext(args, response)
	responseWithGuidance := addAIGuidance(response, "assets_complete_object", context)
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: formatResponse(responseWithGuidance)},
		},
	}, nil
}

func handleTraceRelationshipsTool(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]interface{}]) (*mcp.CallToolResult, error) {
	// Use the parsed arguments directly
	args := params.Arguments
	
	traceParams := common.TraceRelationshipsParams{
		ObjectID: getStringParam(args, "object_id", ""),
		Depth:    getIntParam(args, "depth", 1),
	}
	
	response, err := composite.TraceRelationships(assetsClient, traceParams)
	if err != nil {
		return nil, err
	}
	
	// Add AI guidance
	context := buildToolContext(args, response)
	responseWithGuidance := addAIGuidance(response, "assets_trace_relationships", context)
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: formatResponse(responseWithGuidance)},
		},
	}, nil
}

// Utility functions
func getStringParam(params map[string]interface{}, key, defaultValue string) string {
	if value, ok := params[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getIntParam(params map[string]interface{}, key string, defaultValue int) int {
	if value, ok := params[key]; ok {
		switch v := value.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if intVal, err := strconv.Atoi(v); err == nil {
				return intVal
			}
		}
	}
	return defaultValue
}

func getMapParam(params map[string]interface{}, key string) map[string]interface{} {
	if value, ok := params[key]; ok {
		if mapVal, ok := value.(map[string]interface{}); ok {
			return mapVal
		}
	}
	return make(map[string]interface{})
}

func buildToolContext(params map[string]interface{}, response *common.Response) map[string]interface{} {
	context := map[string]interface{}{
		"success": response.Success,
	}
	
	// Add parameter context
	for key, value := range params {
		context[key] = value
	}
	
	// Add result count if available
	if response.Success && response.Data != nil {
		if data, ok := response.Data.(map[string]interface{}); ok {
			if total, ok := data["total"].(int); ok {
				context["result_count"] = total
			}
		}
	}
	
	return context
}

func addAIGuidance(response *common.Response, toolName string, context map[string]interface{}) *ResponseWithHints {
	// Add success to context for guidance evaluation
	context["success"] = response.Success
	
	// Get AI guidance
	aiGuidance, err := hints.GetAIGuidance(toolName, context)
	if err != nil {
		// If AI guidance fails, provide a basic response
		aiGuidance = &hints.AIGuidance{
			OperationSummary: fmt.Sprintf("Executed %s operation", toolName),
			NextActions:      []hints.AIAction{},
			WorkflowContext: hints.WorkflowContext{
				CurrentState:         "operation_completed",
				CompletionPercentage: 100,
			},
			SemanticContext: context,
		}
	}
	
	return &ResponseWithHints{
		Response:   response,
		AIGuidance: aiGuidance,
	}
}

type ResponseWithHints struct {
	*common.Response
	AIGuidance *hints.AIGuidance `json:"ai_guidance,omitempty"`
}

func formatResponse(response *ResponseWithHints) string {
	jsonData, _ := json.MarshalIndent(response, "", "  ")
	return string(jsonData)
}

// Main function
func main() {
	// Setup logger for MCP server (stderr only)
	logger.SetupStandardLogger()

	// Initialize server
	mcpServer, err := NewMCPServer()
	if err != nil {
		logger.Fatal("Failed to initialize MCP server: %v", err)
	}
	defer assetsClient.Close()

	// Start the MCP server using the SDK with stdio transport
	transport := mcp.NewStdioTransport()
	if err := mcpServer.Run(context.Background(), transport); err != nil {
		logger.Fatal("MCP server failed: %v", err)
	}
}