package validation

import (
	"context"
	"fmt"
	"strings"

	"github.com/aaronsb/atlassian-assets/internal/client"
	"github.com/aaronsb/atlassian-assets/internal/property"
)

// ObjectValidator provides comprehensive validation for Assets objects
type ObjectValidator struct {
	client           *client.AssetsClient
	propertyResolver *property.PropertyResolver
}

// NewObjectValidator creates a new object validator
func NewObjectValidator(client *client.AssetsClient) *ObjectValidator {
	return &ObjectValidator{
		client:           client,
		propertyResolver: property.NewPropertyResolver(client),
	}
}

// ValidationResult holds the result of object validation
type ValidationResult struct {
	Valid              bool                        `json:"valid"`
	ObjectTypeID       string                     `json:"object_type_id"`
	ResolvedProperties []*property.PropertyValue   `json:"resolved_properties,omitempty"`
	Errors             []ValidationError          `json:"errors,omitempty"`
	Warnings           []ValidationWarning        `json:"warnings,omitempty"`
	RequiredFields     []string                   `json:"required_fields,omitempty"`
	OptionalFields     []string                   `json:"optional_fields,omitempty"`
}

// ValidationError represents a validation error that prevents object creation/update
type ValidationError struct {
	Field       string `json:"field"`
	Message     string `json:"message"`
	Code        string `json:"code"`
	Severity    string `json:"severity"`
	Suggestion  string `json:"suggestion,omitempty"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Field, e.Message)
}

// ValidationWarning represents a validation warning that doesn't prevent operation
type ValidationWarning struct {
	Field      string `json:"field"`
	Message    string `json:"message"`
	Code       string `json:"code"`
	Suggestion string `json:"suggestion,omitempty"`
}

// ValidateObject validates an object's properties against its schema
func (ov *ObjectValidator) ValidateObject(ctx context.Context, objectTypeID string, properties map[string]interface{}) (*ValidationResult, error) {
	result := &ValidationResult{
		ObjectTypeID: objectTypeID,
		Valid:        true,
	}

	// Get object type metadata
	metadata, err := ov.propertyResolver.GetObjectTypeMetadata(ctx, objectTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get object type metadata: %w", err)
	}

	// Categorize fields
	for _, meta := range metadata {
		if meta.Required && !meta.System {
			result.RequiredFields = append(result.RequiredFields, meta.Name)
		} else if meta.Editable {
			result.OptionalFields = append(result.OptionalFields, meta.Name)
		}
	}

	// Resolve and validate properties
	resolvedProps, propErrors := ov.propertyResolver.ResolveObjectProperties(ctx, objectTypeID, properties)
	result.ResolvedProperties = resolvedProps

	// Convert property errors to validation errors
	for _, err := range propErrors {
		validationErr := ValidationError{
			Message:  err.Error(),
			Severity: "error",
			Code:     "PROPERTY_ERROR",
		}

		// Extract field name from error if possible
		if strings.Contains(err.Error(), "property '") {
			start := strings.Index(err.Error(), "property '") + 10
			end := strings.Index(err.Error()[start:], "'")
			if end > 0 {
				validationErr.Field = err.Error()[start : start+end]
			}
		}

		// Provide suggestions based on error type
		if strings.Contains(err.Error(), "required field") {
			validationErr.Code = "REQUIRED_FIELD_MISSING"
			validationErr.Suggestion = "Please provide a value for this required field"
		} else if strings.Contains(err.Error(), "unknown property") {
			validationErr.Code = "UNKNOWN_PROPERTY"
			validationErr.Suggestion = "Check the property name spelling or use 'assets attributes --type " + objectTypeID + "' to see available fields"
		} else if strings.Contains(err.Error(), "invalid date") {
			validationErr.Code = "INVALID_DATE_FORMAT"
			validationErr.Suggestion = "Use date format YYYY-MM-DD (e.g., 2024-01-15)"
		} else if strings.Contains(err.Error(), "invalid datetime") {
			validationErr.Code = "INVALID_DATETIME_FORMAT"
			validationErr.Suggestion = "Use ISO 8601 format (e.g., 2024-01-15T10:30:00Z)"
		} else if strings.Contains(err.Error(), "invalid option") {
			validationErr.Code = "INVALID_SELECT_OPTION"
			// Extract valid options from error message if available
			if strings.Contains(err.Error(), "allowed:") {
				start := strings.Index(err.Error(), "allowed:") + 9
				validationErr.Suggestion = "Valid options: " + err.Error()[start:]
			}
		} else if strings.Contains(err.Error(), "reference field") {
			validationErr.Code = "INVALID_REFERENCE"
			validationErr.Suggestion = "Reference fields require valid object IDs"
		}

		result.Errors = append(result.Errors, validationErr)
		result.Valid = false
	}

	// Add business rule validations
	ov.validateBusinessRules(ctx, metadata, properties, result)

	// Add warnings for best practices
	ov.generateWarnings(metadata, properties, result)

	return result, nil
}

// validateBusinessRules performs additional business logic validation
func (ov *ObjectValidator) validateBusinessRules(ctx context.Context, metadata map[string]*property.AttributeMetadata, properties map[string]interface{}, result *ValidationResult) {
	// Example business rules (these would be customized based on organization needs)
	
	// Check for reasonable asset tag format if provided
	if assetTag, exists := properties["asset_tag"]; exists && assetTag != "" {
		tagStr := fmt.Sprintf("%v", assetTag)
		if len(tagStr) < 3 {
			result.Errors = append(result.Errors, ValidationError{
				Field:      "asset_tag",
				Message:    "Asset tag should be at least 3 characters long",
				Code:       "ASSET_TAG_TOO_SHORT",
				Severity:   "error",
				Suggestion: "Use a longer, more descriptive asset tag",
			})
			result.Valid = false
		}
	}

	// Check for serial number format if provided
	if serialNum, exists := properties["serial_number"]; exists && serialNum != "" {
		serialStr := fmt.Sprintf("%v", serialNum)
		if strings.Contains(strings.ToLower(serialStr), "test") {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:      "serial_number",
				Message:    "Serial number contains 'test' - ensure this is a real serial number",
				Code:       "SUSPICIOUS_SERIAL_NUMBER",
				Suggestion: "Use the actual hardware serial number",
			})
		}
	}

	// Validate device type and ownership consistency
	deviceType, hasDeviceType := properties["device_type"]
	ownershipType, hasOwnershipType := properties["ownership_type"]
	
	if hasDeviceType && hasOwnershipType {
		deviceTypeStr := fmt.Sprintf("%v", deviceType)
		ownershipTypeStr := fmt.Sprintf("%v", ownershipType)
		
		if strings.EqualFold(deviceTypeStr, "Virtual") && strings.EqualFold(ownershipTypeStr, "BYOD") {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:      "ownership_type",
				Message:    "Virtual devices are typically not BYOD",
				Code:       "UNUSUAL_VIRTUAL_BYOD",
				Suggestion: "Consider if this virtual device should be 'Company owned'",
			})
		}
	}
}

// generateWarnings generates helpful warnings for best practices
func (ov *ObjectValidator) generateWarnings(metadata map[string]*property.AttributeMetadata, properties map[string]interface{}, result *ValidationResult) {
	// Warn about missing important optional fields
	importantOptionalFields := []string{"asset_tag", "serial_number", "model_name"}
	
	for _, fieldName := range importantOptionalFields {
		if _, provided := properties[strings.ToLower(fieldName)]; !provided {
			// Check if field exists in metadata
			if meta, exists := metadata[strings.ToLower(fieldName)]; exists && meta.Editable {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Field:      fieldName,
					Message:    fmt.Sprintf("Optional field '%s' not provided", fieldName),
					Code:       "MISSING_RECOMMENDED_FIELD",
					Suggestion: fmt.Sprintf("Consider providing %s for better asset tracking", fieldName),
				})
			}
		}
	}

	// Warn about very generic names
	if name, exists := properties["name"]; exists {
		nameStr := strings.ToLower(fmt.Sprintf("%v", name))
		genericNames := []string{"test", "laptop", "computer", "device", "asset"}
		
		for _, generic := range genericNames {
			if nameStr == generic {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Field:      "name",
					Message:    "Name is very generic",
					Code:       "GENERIC_NAME",
					Suggestion: "Consider using a more specific name that includes model, user, or location",
				})
				break
			}
		}
	}
}

// ValidateForCreate validates an object for creation (stricter validation)
func (ov *ObjectValidator) ValidateForCreate(ctx context.Context, objectTypeID string, properties map[string]interface{}) (*ValidationResult, error) {
	result, err := ov.ValidateObject(ctx, objectTypeID, properties)
	if err != nil {
		return nil, err
	}

	// Additional validation for creation
	// Ensure all required fields are provided
	metadata, err := ov.propertyResolver.GetObjectTypeMetadata(ctx, objectTypeID)
	if err != nil {
		return result, nil // Return partial result if metadata fails
	}

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
				result.Errors = append(result.Errors, ValidationError{
					Field:      meta.Name,
					Message:    fmt.Sprintf("Required field '%s' is missing", meta.Name),
					Code:       "REQUIRED_FIELD_MISSING",
					Severity:   "error",
					Suggestion: "This field must be provided when creating new objects",
				})
				result.Valid = false
			}
		}
	}

	return result, nil
}

// ValidateForUpdate validates an object for update (more lenient)
func (ov *ObjectValidator) ValidateForUpdate(ctx context.Context, objectTypeID string, properties map[string]interface{}) (*ValidationResult, error) {
	result, err := ov.ValidateObject(ctx, objectTypeID, properties)
	if err != nil {
		return nil, err
	}

	// For updates, we're more lenient about missing fields
	// Only validate the fields that are actually being updated
	return result, nil
}

// CompletionResult holds the result of intelligent object completion
type CompletionResult struct {
	Success            bool                        `json:"success"`
	ObjectTypeID       string                     `json:"object_type_id"`
	OriginalProperties map[string]interface{}     `json:"original_properties"`
	CompletedProperties map[string]interface{}    `json:"completed_properties"`
	ResolvedProperties []*property.PropertyValue  `json:"resolved_properties"`
	AppliedDefaults    []DefaultApplication       `json:"applied_defaults"`
	Suggestions        []CompletionSuggestion     `json:"suggestions"`
	Warnings           []ValidationWarning        `json:"warnings,omitempty"`
	MissingCritical    []string                   `json:"missing_critical,omitempty"`
}

// DefaultApplication represents an automatically applied default value
type DefaultApplication struct {
	Field       string      `json:"field"`
	Value       interface{} `json:"value"`
	Reason      string      `json:"reason"`
	Confidence  string      `json:"confidence"` // "high", "medium", "low"
}

// CompletionSuggestion represents a suggested completion for missing information
type CompletionSuggestion struct {
	Field       string        `json:"field"`
	Message     string        `json:"message"`
	Options     []interface{} `json:"options,omitempty"`
	Required    bool          `json:"required"`
	Priority    string        `json:"priority"` // "critical", "important", "optional"
}

// CompleteObject attempts to intelligently complete an object with reasonable defaults
func (ov *ObjectValidator) CompleteObject(ctx context.Context, objectTypeID string, partialProperties map[string]interface{}) (*CompletionResult, error) {
	result := &CompletionResult{
		ObjectTypeID:        objectTypeID,
		OriginalProperties:  partialProperties,
		CompletedProperties: make(map[string]interface{}),
		Success:             true,
	}

	// Copy original properties
	for k, v := range partialProperties {
		result.CompletedProperties[k] = v
	}

	// Get object type metadata
	metadata, err := ov.propertyResolver.GetObjectTypeMetadata(ctx, objectTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get object type metadata: %w", err)
	}

	// Apply intelligent defaults and completion
	ov.applyIntelligentDefaults(metadata, result)
	ov.generateCompletionSuggestions(metadata, result)

	// Try to resolve the completed properties
	resolvedProps, propErrors := ov.propertyResolver.ResolveObjectProperties(ctx, objectTypeID, result.CompletedProperties)
	result.ResolvedProperties = resolvedProps

	// Check if we still have critical missing fields
	for _, err := range propErrors {
		if strings.Contains(err.Error(), "required field") {
			// Extract field name
			if strings.Contains(err.Error(), "property '") {
				start := strings.Index(err.Error(), "property '") + 10
				end := strings.Index(err.Error()[start:], "'")
				if end > 0 {
					fieldName := err.Error()[start : start+end]
					result.MissingCritical = append(result.MissingCritical, fieldName)
				}
			}
		}
	}

	// If we have critical missing fields, mark as unsuccessful but still return the partial completion
	if len(result.MissingCritical) > 0 {
		result.Success = false
	}

	return result, nil
}

// applyIntelligentDefaults applies reasonable defaults for missing fields
func (ov *ObjectValidator) applyIntelligentDefaults(metadata map[string]*property.AttributeMetadata, result *CompletionResult) {
	// Default strategies based on field names and types
	
	// Asset Status defaults
	if _, exists := result.CompletedProperties["asset_status"]; !exists {
		if meta, found := metadata["asset status"]; found && len(meta.StatusValues) > 0 {
			// Default to first status value (often "In Stock" or similar)
			defaultStatus := meta.StatusValues[0]
			result.CompletedProperties["asset_status"] = defaultStatus
			result.AppliedDefaults = append(result.AppliedDefaults, DefaultApplication{
				Field:      "asset_status",
				Value:      defaultStatus,
				Reason:     "Applied default status value",
				Confidence: "medium",
			})
		}
	}

	// Device Type defaults
	if _, exists := result.CompletedProperties["device_type"]; !exists {
		if meta, found := metadata["device type"]; found && len(meta.SelectOptions) > 0 {
			// Default to "Physical" if available
			defaultType := "Physical"
			found := false
			for _, option := range meta.SelectOptions {
				if strings.EqualFold(option, "Physical") {
					defaultType = option
					found = true
					break
				}
			}
			if !found {
				defaultType = meta.SelectOptions[0] // Fallback to first option
			}
			
			result.CompletedProperties["device_type"] = defaultType
			result.AppliedDefaults = append(result.AppliedDefaults, DefaultApplication{
				Field:      "device_type",
				Value:      defaultType,
				Reason:     "Applied default device type",
				Confidence: "high",
			})
		}
	}

	// Ownership Type defaults
	if _, exists := result.CompletedProperties["ownership_type"]; !exists {
		if meta, found := metadata["ownership type"]; found && len(meta.SelectOptions) > 0 {
			// Default to "Company owned" if available
			defaultOwnership := "Company owned"
			found := false
			for _, option := range meta.SelectOptions {
				if strings.Contains(strings.ToLower(option), "company") {
					defaultOwnership = option
					found = true
					break
				}
			}
			if !found {
				defaultOwnership = meta.SelectOptions[0] // Fallback to first option
			}
			
			result.CompletedProperties["ownership_type"] = defaultOwnership
			result.AppliedDefaults = append(result.AppliedDefaults, DefaultApplication{
				Field:      "ownership_type",
				Value:      defaultOwnership,
				Reason:     "Applied default ownership type",
				Confidence: "medium",
			})
		}
	}

	// Generate Asset Tag if name is provided but asset tag isn't
	if name, hasName := result.CompletedProperties["name"]; hasName {
		if _, hasAssetTag := result.CompletedProperties["asset_tag"]; !hasAssetTag {
			nameStr := fmt.Sprintf("%v", name)
			// Generate a simple asset tag based on name
			assetTag := strings.ToUpper(strings.ReplaceAll(nameStr, " ", "-"))
			if len(assetTag) > 20 {
				assetTag = assetTag[:20] // Truncate if too long
			}
			
			result.CompletedProperties["asset_tag"] = assetTag
			result.AppliedDefaults = append(result.AppliedDefaults, DefaultApplication{
				Field:      "asset_tag",
				Value:      assetTag,
				Reason:     "Generated asset tag from name",
				Confidence: "low",
			})
		}
	}
}

// generateCompletionSuggestions creates suggestions for missing important fields
func (ov *ObjectValidator) generateCompletionSuggestions(metadata map[string]*property.AttributeMetadata, result *CompletionResult) {
	for _, meta := range metadata {
		fieldNameLower := strings.ToLower(meta.Name)
		
		// Skip if already provided
		if _, exists := result.CompletedProperties[fieldNameLower]; exists {
			continue
		}

		// Skip system fields
		if meta.System {
			continue
		}

		suggestion := CompletionSuggestion{
			Field:    meta.Name,
			Required: meta.Required,
		}

		if meta.Required {
			suggestion.Priority = "critical"
			suggestion.Message = fmt.Sprintf("Required field '%s' is missing", meta.Name)
		} else {
			// Determine importance of optional fields
			importantFields := []string{"serial_number", "model_name", "purchase_date"}
			isImportant := false
			for _, important := range importantFields {
				if strings.Contains(fieldNameLower, important) {
					isImportant = true
					break
				}
			}
			
			if isImportant {
				suggestion.Priority = "important"
				suggestion.Message = fmt.Sprintf("Important field '%s' would improve asset tracking", meta.Name)
			} else {
				suggestion.Priority = "optional"
				suggestion.Message = fmt.Sprintf("Optional field '%s' can be provided", meta.Name)
			}
		}

		// Add options for select and status fields
		if meta.DataType == "Select" && len(meta.SelectOptions) > 0 {
			for _, option := range meta.SelectOptions {
				suggestion.Options = append(suggestion.Options, option)
			}
		} else if meta.DataType == "Status" && len(meta.StatusValues) > 0 {
			for _, statusID := range meta.StatusValues {
				suggestion.Options = append(suggestion.Options, statusID)
			}
		}

		// Add field-specific guidance
		switch fieldNameLower {
		case "serial_number":
			suggestion.Message += " (helps with warranty tracking and asset identification)"
		case "model_name":
			suggestion.Message += " (links to hardware specifications and compatibility)"
		case "purchase_date":
			suggestion.Message += " (important for warranty and depreciation tracking)"
		case "asset_tag":
			suggestion.Message += " (unique identifier for physical asset management)"
		}

		result.Suggestions = append(result.Suggestions, suggestion)
	}
}

// GetValidationSummary returns a human-readable summary of validation results
func (ov *ObjectValidator) GetValidationSummary(result *ValidationResult) string {
	if result.Valid {
		summary := fmt.Sprintf("✅ Validation passed for object type %s", result.ObjectTypeID)
		if len(result.Warnings) > 0 {
			summary += fmt.Sprintf(" (with %d warnings)", len(result.Warnings))
		}
		return summary
	}

	summary := fmt.Sprintf("❌ Validation failed for object type %s with %d errors", result.ObjectTypeID, len(result.Errors))
	if len(result.Warnings) > 0 {
		summary += fmt.Sprintf(" and %d warnings", len(result.Warnings))
	}
	return summary
}

// GetCompletionSummary returns a human-readable summary of completion results
func (ov *ObjectValidator) GetCompletionSummary(result *CompletionResult) string {
	if result.Success {
		summary := fmt.Sprintf("✅ Object completion successful for type %s", result.ObjectTypeID)
		if len(result.AppliedDefaults) > 0 {
			summary += fmt.Sprintf(" (applied %d defaults)", len(result.AppliedDefaults))
		}
		return summary
	}

	summary := fmt.Sprintf("⚠️ Object completion partially successful for type %s", result.ObjectTypeID)
	if len(result.MissingCritical) > 0 {
		summary += fmt.Sprintf(" (missing %d critical fields)", len(result.MissingCritical))
	}
	return summary
}