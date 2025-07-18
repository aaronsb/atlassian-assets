package main

import (
	"fmt"

	"github.com/aaronsb/atlassian-assets/internal/client"
	"github.com/aaronsb/atlassian-assets/internal/config"
	"github.com/aaronsb/atlassian-assets/internal/hints"
	"github.com/aaronsb/atlassian-assets/cmd/assets/common"
)

// getClient creates and returns a configured Assets client
func getClient() (common.ClientInterface, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override with command-line flags if provided
	if workspaceID != "" {
		cfg.WorkspaceID = workspaceID
	}
	if profile != "" {
		cfg.Profile = profile
	}

	client, err := client.NewAssetsClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return client, nil
}

// ResponseWithHints represents a response with contextual hints
type ResponseWithHints struct {
	*common.Response
	NextSteps []string `json:"next_steps,omitempty"`
	AIGuidance *hints.AIGuidance `json:"ai_guidance,omitempty"`
}

// addCLIHints adds CLI-specific contextual hints to a response
func addCLIHints(response *common.Response, commandType string, context map[string]interface{}) *ResponseWithHints {
	// Add success to context for hint evaluation
	context["success"] = response.Success
	
	// Get CLI hints
	contextualHints := hints.GetContextualHints(commandType, context)
	
	return &ResponseWithHints{
		Response:  response,
		NextSteps: contextualHints,
	}
}

// addAIGuidance adds AI-specific guidance to a response
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

// getResultCount extracts result count from response data
func getResultCount(response *common.Response) int {
	if !response.Success || response.Data == nil {
		return 0
	}
	
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return 0
	}
	
	if total, ok := data["total"].(int); ok {
		return total
	}
	
	if objects, ok := data["objects"].([]interface{}); ok {
		return len(objects)
	}
	
	return 0
}

// extractFirstResultID extracts the first result ID from a response
func extractFirstResultID(response *common.Response) string {
	if !response.Success || response.Data == nil {
		return ""
	}
	
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return ""
	}
	
	objects, ok := data["objects"].([]interface{})
	if !ok || len(objects) == 0 {
		return ""
	}
	
	firstObject, ok := objects[0].(map[string]interface{})
	if !ok {
		return ""
	}
	
	if id, ok := firstObject["id"].(string); ok {
		return id
	}
	
	return ""
}

// buildContext builds context information for hint systems
func buildContext(params interface{}, response *common.Response) map[string]interface{} {
	context := map[string]interface{}{
		"success":      response.Success,
		"result_count": getResultCount(response),
	}
	
	// Add first result ID if available
	if firstID := extractFirstResultID(response); firstID != "" {
		context["first_result_id"] = firstID
	}
	
	// Add parameter-specific context
	switch p := params.(type) {
	case common.SearchParams:
		context["search_query"] = p.Query
		context["simple_term"] = p.Simple
		context["schema_id"] = p.Schema
	case common.ListParams:
		context["schema_id"] = p.Schema
	case common.GetParams:
		context["object_id"] = p.ID
	case common.CreateObjectTypeParams:
		context["schema_id"] = p.Schema
		context["object_type_name"] = p.Name
		context["has_parent"] = p.Parent != nil
		context["has_description"] = p.Description != ""
	case common.CreateObjectParams:
		context["object_type_id"] = p.ObjectTypeID
	}
	
	return context
}