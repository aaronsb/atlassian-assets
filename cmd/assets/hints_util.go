package main

import (
	"github.com/aaronsb/atlassian-assets/internal/hints"
	"github.com/aaronsb/atlassian-assets/internal/client"
)

// Helper function to add contextual hints using centralized system
func addNextStepHints(response interface{}, commandType string, context map[string]interface{}) interface{} {
	// Convert response to map for modification
	responseMap := make(map[string]interface{})
	
	// Handle different response types
	switch r := response.(type) {
	case *Response:
		responseMap["success"] = r.Success
		responseMap["data"] = r.Data
		if r.Error != "" {
			responseMap["error"] = r.Error
		}
		// Add success to context for hint evaluation
		context["success"] = r.Success
	case *client.Response:
		responseMap["success"] = r.Success
		responseMap["data"] = r.Data
		if r.Error != "" {
			responseMap["error"] = r.Error
		}
		// Add success to context for hint evaluation
		context["success"] = r.Success
	case map[string]interface{}:
		responseMap = r
		// Add success to context for hint evaluation
		if success, ok := r["success"].(bool); ok {
			context["success"] = success
		}
	default:
		return response // Return as-is if we can't parse
	}
	
	// Get contextual hints from centralized system
	contextualHints := hints.GetContextualHints(commandType, context)
	
	if len(contextualHints) > 0 {
		responseMap["next_steps"] = contextualHints
	}
	
	return responseMap
}