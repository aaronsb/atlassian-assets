package property

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ctreminiom/go-atlassian/v2/pkg/infra/models"
	"github.com/aaronsb/atlassian-assets/internal/client"
)

// PropertyResolver handles resolution and validation of object properties
type PropertyResolver struct {
	client *client.AssetsClient
}

// NewPropertyResolver creates a new property resolver
func NewPropertyResolver(client *client.AssetsClient) *PropertyResolver {
	return &PropertyResolver{
		client: client,
	}
}

// AttributeMetadata holds metadata about an object type attribute
type AttributeMetadata struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	DataType            string   `json:"data_type"`
	Required            bool     `json:"required"`
	Editable            bool     `json:"editable"`
	System              bool     `json:"system"`
	MaxCardinality      int      `json:"max_cardinality"`
	MinCardinality      int      `json:"min_cardinality"`
	ReferenceObjectType string   `json:"reference_object_type,omitempty"`
	StatusValues        []string `json:"status_values,omitempty"`
	SelectOptions       []string `json:"select_options,omitempty"`
	Description         string   `json:"description,omitempty"`
}

// PropertyValue represents a resolved property value ready for API submission
type PropertyValue struct {
	AttributeID string      `json:"attribute_id"`
	Value       interface{} `json:"value"`
	DataType    string      `json:"data_type"`
}

// ValidationError represents a property validation error
type ValidationError struct {
	AttributeName string `json:"attribute_name"`
	Message       string `json:"message"`
	Code          string `json:"code"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for %s: %s", e.AttributeName, e.Message)
}

// GetObjectTypeMetadata retrieves and parses metadata for all attributes of an object type
func (pr *PropertyResolver) GetObjectTypeMetadata(ctx context.Context, objectTypeID string) (map[string]*AttributeMetadata, error) {
	response, err := pr.client.GetObjectTypeAttributes(ctx, objectTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get object type attributes: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API error: %s", response.Error)
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	attributesValue := data["attributes"]
	attributes, ok := attributesValue.([]*models.ObjectTypeAttributeScheme)
	if !ok {
		return nil, fmt.Errorf("attributes type mismatch: expected []*models.ObjectTypeAttributeScheme, got %T", attributesValue)
	}

	metadata := make(map[string]*AttributeMetadata)

	for _, attr := range attributes {
		if attr == nil {
			continue
		}

		meta := &AttributeMetadata{
			ID:             attr.ID,
			Name:           attr.Name,
			Required:       attr.MinimumCardinality == 1,
			Editable:       attr.Editable,
			System:         attr.System,
			MaxCardinality: attr.MaximumCardinality,
			MinCardinality: attr.MinimumCardinality,
			Description:    attr.Description,
		}

		// Determine data type
		if attr.DefaultType != nil {
			if attr.DefaultType.Name != "" {
				meta.DataType = attr.DefaultType.Name
			} else if attr.DefaultType.ID > 0 {
				meta.DataType = fmt.Sprintf("type_%d", attr.DefaultType.ID)
			}
		}

		// Handle reference attributes
		if attr.Type == 1 && attr.ReferenceObjectTypeID != "" {
			meta.DataType = "Reference"
			meta.ReferenceObjectType = attr.ReferenceObjectTypeID
		}

		// Handle status attributes (type 7)
		if attr.Type == 7 && len(attr.TypeValueMulti) > 0 {
			meta.DataType = "Status"
			meta.StatusValues = attr.TypeValueMulti
		}

		// Handle select attributes
		if meta.DataType == "Select" && attr.Options != "" {
			meta.SelectOptions = strings.Split(attr.Options, ", ")
		}

		// Use attribute name as key for easy lookup
		metadata[strings.ToLower(attr.Name)] = meta
	}

	return metadata, nil
}

// ValidateProperty validates a property value against its metadata
func (pr *PropertyResolver) ValidateProperty(meta *AttributeMetadata, value interface{}) *ValidationError {
	// Check if required field is missing
	if meta.Required && (value == nil || value == "") {
		return &ValidationError{
			AttributeName: meta.Name,
			Message:       "required field is missing",
			Code:          "REQUIRED_FIELD_MISSING",
		}
	}

	// Skip validation for nil/empty non-required fields
	if value == nil || value == "" {
		return nil
	}

	// Check if field is editable
	if !meta.Editable && !meta.System {
		return &ValidationError{
			AttributeName: meta.Name,
			Message:       "field is not editable",
			Code:          "FIELD_NOT_EDITABLE",
		}
	}

	valueStr := fmt.Sprintf("%v", value)

	// Validate based on data type
	switch meta.DataType {
	case "Date":
		if err := pr.validateDate(valueStr); err != nil {
			return &ValidationError{
				AttributeName: meta.Name,
				Message:       fmt.Sprintf("invalid date format: %s", err.Error()),
				Code:          "INVALID_DATE_FORMAT",
			}
		}
	case "DateTime", "type_6":
		if err := pr.validateDateTime(valueStr); err != nil {
			return &ValidationError{
				AttributeName: meta.Name,
				Message:       fmt.Sprintf("invalid datetime format: %s", err.Error()),
				Code:          "INVALID_DATETIME_FORMAT",
			}
		}
	case "Select":
		if !pr.validateSelectOption(valueStr, meta.SelectOptions) {
			return &ValidationError{
				AttributeName: meta.Name,
				Message:       fmt.Sprintf("invalid option '%s', allowed: %v", valueStr, meta.SelectOptions),
				Code:          "INVALID_SELECT_OPTION",
			}
		}
	case "Status":
		if !pr.validateStatusValue(valueStr, meta.StatusValues) {
			return &ValidationError{
				AttributeName: meta.Name,
				Message:       fmt.Sprintf("invalid status value '%s', allowed IDs: %v", valueStr, meta.StatusValues),
				Code:          "INVALID_STATUS_VALUE",
			}
		}
	}

	return nil
}

// validateDate validates date format (YYYY-MM-DD)
func (pr *PropertyResolver) validateDate(value string) error {
	_, err := time.Parse("2006-01-02", value)
	return err
}

// validateDateTime validates datetime format (ISO 8601)
func (pr *PropertyResolver) validateDateTime(value string) error {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	}
	
	for _, format := range formats {
		if _, err := time.Parse(format, value); err == nil {
			return nil
		}
	}
	
	return fmt.Errorf("unable to parse datetime, supported formats: ISO 8601, RFC3339")
}

// validateSelectOption validates select field option
func (pr *PropertyResolver) validateSelectOption(value string, options []string) bool {
	for _, option := range options {
		if strings.EqualFold(value, option) {
			return true
		}
	}
	return false
}

// validateStatusValue validates status field value (accepts both names and IDs)
func (pr *PropertyResolver) validateStatusValue(value string, allowedIDs []string) bool {
	// Check if value is a valid status ID
	for _, id := range allowedIDs {
		if value == id {
			return true
		}
	}
	
	// TODO: Add status name to ID resolution
	// For now, only accept direct status IDs
	return false
}

// ResolveProperty resolves a property value to the format expected by the API
func (pr *PropertyResolver) ResolveProperty(ctx context.Context, meta *AttributeMetadata, value interface{}) (*PropertyValue, error) {
	// Validate the property first
	if err := pr.ValidateProperty(meta, value); err != nil {
		return nil, err
	}

	// Skip nil/empty values for non-required fields
	if value == nil || value == "" {
		if !meta.Required {
			return nil, nil
		}
	}

	resolved := &PropertyValue{
		AttributeID: meta.ID,
		DataType:    meta.DataType,
	}

	valueStr := fmt.Sprintf("%v", value)

	switch meta.DataType {
	case "Reference":
		// TODO: Resolve reference object names to IDs
		// For now, assume the value is already an object ID
		if _, err := strconv.Atoi(valueStr); err != nil {
			return nil, fmt.Errorf("reference field '%s' requires object ID, got: %s", meta.Name, valueStr)
		}
		resolved.Value = valueStr

	case "Status":
		// Ensure status value is a valid ID
		if !pr.validateStatusValue(valueStr, meta.StatusValues) {
			return nil, fmt.Errorf("invalid status value '%s' for field '%s'", valueStr, meta.Name)
		}
		resolved.Value = valueStr

	case "Select":
		// Validate and normalize select option
		for _, option := range meta.SelectOptions {
			if strings.EqualFold(valueStr, option) {
				resolved.Value = option // Use the exact case from options
				break
			}
		}

	case "Date":
		// Parse and format date
		date, err := time.Parse("2006-01-02", valueStr)
		if err != nil {
			return nil, fmt.Errorf("invalid date format for field '%s': %s", meta.Name, valueStr)
		}
		resolved.Value = date.Format("2006-01-02")

	case "DateTime", "type_6":
		// Parse and format datetime
		var parsedTime time.Time
		var err error
		
		formats := []string{
			time.RFC3339,
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
		}
		
		for _, format := range formats {
			if parsedTime, err = time.Parse(format, valueStr); err == nil {
				break
			}
		}
		
		if err != nil {
			return nil, fmt.Errorf("invalid datetime format for field '%s': %s", meta.Name, valueStr)
		}
		
		resolved.Value = parsedTime.Format(time.RFC3339)

	default:
		// Text and other simple types
		resolved.Value = valueStr
	}

	return resolved, nil
}

// ResolveObjectProperties resolves a map of property names to values
func (pr *PropertyResolver) ResolveObjectProperties(ctx context.Context, objectTypeID string, properties map[string]interface{}) ([]*PropertyValue, []error) {
	metadata, err := pr.GetObjectTypeMetadata(ctx, objectTypeID)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to get object type metadata: %w", err)}
	}

	var resolved []*PropertyValue
	var errors []error

	// Process provided properties
	for propName, propValue := range properties {
		meta, exists := metadata[strings.ToLower(propName)]
		if !exists {
			errors = append(errors, fmt.Errorf("unknown property: %s", propName))
			continue
		}

		resolvedProp, err := pr.ResolveProperty(ctx, meta, propValue)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to resolve property '%s': %w", propName, err))
			continue
		}

		if resolvedProp != nil {
			resolved = append(resolved, resolvedProp)
		}
	}

	// Check for missing required properties
	for _, meta := range metadata {
		if meta.Required && !meta.System {
			found := false
			for propName := range properties {
				if strings.EqualFold(propName, meta.Name) {
					found = true
					break
				}
			}
			if !found {
				errors = append(errors, &ValidationError{
					AttributeName: meta.Name,
					Message:       "required field is missing",
					Code:          "REQUIRED_FIELD_MISSING",
				})
			}
		}
	}

	return resolved, errors
}