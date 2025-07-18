package common

import (
	"context"
	"github.com/aaronsb/atlassian-assets/internal/client"
)

// Parameter structures for foundation functions
type SearchParams struct {
	Query    string // AQL query
	Simple   string // Simple search term
	Schema   string // Schema ID or name
	Type     string // Object type filter
	Status   string // Status filter
	Owner    string // Owner filter
	Limit    int    // Maximum results
	Offset   int    // Pagination offset
}

type ListParams struct {
	Schema string // Schema ID or name
	Type   string // Object type filter
	Filter string // AQL filter
	Limit  int    // Maximum results
	Offset int    // Pagination offset
}

type GetParams struct {
	ID string // Object ID
}

type CreateObjectTypeParams struct {
	Schema      string  // Schema ID or name
	Name        string  // Object type name
	Description string  // Description
	Parent      *string // Parent object type ID
	Icon        string  // Icon ID
}

type CreateObjectParams struct {
	ObjectTypeID string                 // Object type ID
	Attributes   map[string]interface{} // Attribute values
}

type UpdateObjectParams struct {
	ID   string                 // Object ID
	Data map[string]interface{} // Updated attributes
}

type DeleteObjectParams struct {
	ID string // Object ID
}

type GetObjectTypeAttributesParams struct {
	ObjectTypeID string // Object type ID
}

type ValidateObjectParams struct {
	ObjectTypeID string                 // Object type ID
	Data         map[string]interface{} // Object data to validate
}

type CompleteObjectParams struct {
	ObjectTypeID string                 // Object type ID
	Data         map[string]interface{} // Partial object data
}

type ListSchemasParams struct {
	// No parameters needed for listing schemas
}

type GetSchemaParams struct {
	SchemaID string // Schema ID
}

type BrowseSchemaParams struct {
	SchemaID string // Schema ID
}

type TraceRelationshipsParams struct {
	ObjectID string // Object ID
	Depth    int    // Relationship depth
}

// Response wrapper for consistent output
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// NewSuccessResponse creates a successful response
func NewSuccessResponse(data interface{}) *Response {
	return &Response{
		Success: true,
		Data:    data,
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(err error) *Response {
	return &Response{
		Success: false,
		Error:   err.Error(),
	}
}

// Client interface for dependency injection
type ClientInterface interface {
	SearchObjectsWithPagination(ctx context.Context, query string, limit int, offset int) (*client.Response, error)
	ListObjectsWithPagination(ctx context.Context, schemaID string, limit int, offset int) (*client.Response, error)
	GetObject(ctx context.Context, objectID string) (*client.Response, error)
	CreateObjectType(ctx context.Context, schemaID, name, description, iconID string, parentObjectTypeID *string) (*client.Response, error)
	CreateObject(ctx context.Context, objectTypeID string, attributes map[string]interface{}) (*client.Response, error)
	DeleteObject(ctx context.Context, objectID string) (*client.Response, error)
	GetObjectTypeAttributes(ctx context.Context, objectTypeID string) (*client.Response, error)
	ListSchemas(ctx context.Context) (*client.Response, error)
	GetSchema(ctx context.Context, schemaID string) (*client.Response, error)
	GetObjectTypes(ctx context.Context, schemaID string) (*client.Response, error)
	Close() error
}