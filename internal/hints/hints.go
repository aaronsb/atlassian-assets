package hints

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed workflow_hints.json ai_workflow_hints.json ai_tool_semantics.json ai_error_recovery.json
var hintsFS embed.FS

// HintsLibrary represents the centralized hints configuration
type HintsLibrary struct {
	Version     string                        `json:"version"`
	Description string                        `json:"description"`
	Workflows   map[string]Workflow           `json:"workflows"`
	Contexts    map[string]Context            `json:"contexts"`
	Templates   map[string]string             `json:"command_templates"`
	Categories  map[string]Category           `json:"categories"`
}

// Workflow represents a complete workflow with steps
type Workflow struct {
	Description string `json:"description"`
	Steps       []Step `json:"steps"`
}

// Step represents a single step in a workflow
type Step struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Command     string   `json:"command"`
	NextSteps   []string `json:"next_steps"`
}

// Context represents contextual hints for a specific command
type Context struct {
	Hints []Hint `json:"hints"`
}

// Hint represents a single contextual hint
type Hint struct {
	Condition string `json:"condition"`
	Message   string `json:"message"`
	Priority  string `json:"priority"`
	Category  string `json:"category"`
}

// Category represents a hint category with metadata
type Category struct {
	Description string `json:"description"`
	Priority    int    `json:"priority"`
	Icon        string `json:"icon"`
}

// AI-specific hint structures
type AIHintsLibrary struct {
	Version     string                    `json:"version"`
	Description string                    `json:"description"`
	AIWorkflows map[string]AIWorkflow     `json:"ai_workflows"`
}

type AIWorkflow struct {
	Description     string    `json:"description"`
	SemanticContext string    `json:"semantic_context"`
	Steps           []AIStep  `json:"steps"`
}

type AIStep struct {
	ID               string                 `json:"id"`
	ToolName         string                 `json:"tool_name"`
	Parameters       map[string]interface{} `json:"parameters"`
	SuccessIndicators []string              `json:"success_indicators"`
	NextActions      []AIAction             `json:"next_actions"`
}

type AIAction struct {
	Tool       string                 `json:"tool"`
	Reason     string                 `json:"reason"`
	Confidence string                 `json:"confidence"`
	Parameters map[string]interface{} `json:"parameters"`
}

type AIToolSemanticsLibrary struct {
	Version         string                    `json:"version"`
	Description     string                    `json:"description"`
	AIToolSemantics map[string]AIToolSemantic `json:"ai_tool_semantics"`
}

type AIToolSemantic struct {
	SemanticDescription string                    `json:"semantic_description"`
	OperationType       string                    `json:"operation_type"`
	ParameterSemantics  map[string]ParameterInfo  `json:"parameter_semantics"`
	ReturnSemantics     ReturnSemantics           `json:"return_semantics"`
	AIDecisionFramework AIDecisionFramework       `json:"ai_decision_framework"`
}

type ParameterInfo struct {
	Description string   `json:"description"`
	Constraints string   `json:"constraints"`
	Examples    []string `json:"examples"`
}

type ReturnSemantics struct {
	SuccessPatterns     map[string]string `json:"success_patterns"`
	ContextPreservation []string          `json:"context_preservation"`
}

type AIDecisionFramework struct {
	WhenToUse    []string            `json:"when_to_use"`
	AvoidWhen    []string            `json:"avoid_when"`
	Alternatives map[string]string   `json:"alternatives"`
}

type AIErrorRecoveryLibrary struct {
	Version         string                           `json:"version"`
	Description     string                           `json:"description"`
	AIErrorRecovery map[string]map[string]ErrorInfo  `json:"ai_error_recovery"`
}

type ErrorInfo struct {
	ErrorPattern    string            `json:"error_pattern"`
	RecoveryActions []RecoveryAction  `json:"recovery_actions"`
	Explanation     string            `json:"explanation"`
	PreventionTip   string            `json:"prevention_tip"`
}

type RecoveryAction struct {
	Tool       string                 `json:"tool"`
	Reason     string                 `json:"reason"`
	Parameters map[string]interface{} `json:"parameters"`
	Confidence string                 `json:"confidence"`
}

// AI Guidance response structure
type AIGuidance struct {
	OperationSummary string                 `json:"operation_summary"`
	NextActions      []AIAction             `json:"next_recommended_actions"`
	WorkflowContext  WorkflowContext        `json:"workflow_context"`
	SemanticContext  map[string]interface{} `json:"semantic_context"`
}

type WorkflowContext struct {
	CurrentState          string   `json:"current_state"`
	CompletionPercentage  int      `json:"completion_percentage"`
	AvailableNextSteps    []string `json:"available_next_steps"`
	RequiredNextSteps     []string `json:"required_next_steps"`
}

var globalHints *HintsLibrary
var globalAIHints *AIHintsLibrary
var globalAIToolSemantics *AIToolSemanticsLibrary
var globalAIErrorRecovery *AIErrorRecoveryLibrary

// LoadHints loads the hints library from embedded JSON
func LoadHints() (*HintsLibrary, error) {
	if globalHints != nil {
		return globalHints, nil
	}

	data, err := hintsFS.ReadFile("workflow_hints.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read hints file: %w", err)
	}

	var hints HintsLibrary
	if err := json.Unmarshal(data, &hints); err != nil {
		return nil, fmt.Errorf("failed to parse hints JSON: %w", err)
	}

	globalHints = &hints
	return globalHints, nil
}

// GetContextualHints generates contextual hints for a given command context
func GetContextualHints(contextType string, variables map[string]interface{}) []string {
	hints, err := LoadHints()
	if err != nil {
		return []string{"⚠️ Unable to load contextual hints"}
	}

	context, exists := hints.Contexts[contextType]
	if !exists {
		return []string{}
	}

	var results []string
	for _, hint := range context.Hints {
		if evaluateCondition(hint.Condition, variables) {
			message := substituteVariables(hint.Message, variables, hints.Templates)
			
			// Add category icon if available
			if category, exists := hints.Categories[hint.Category]; exists {
				message = category.Icon + " " + strings.TrimPrefix(message, "💡 ")
			}
			
			results = append(results, message)
		}
	}

	// Sort by priority (high -> medium -> low)
	return sortHintsByPriority(results, context.Hints)
}

// GetWorkflowSteps returns the steps for a specific workflow
func GetWorkflowSteps(workflowID string) ([]Step, error) {
	hints, err := LoadHints()
	if err != nil {
		return nil, err
	}

	workflow, exists := hints.Workflows[workflowID]
	if !exists {
		return nil, fmt.Errorf("workflow '%s' not found", workflowID)
	}

	return workflow.Steps, nil
}

// GetNextSteps returns possible next steps for a given step ID
func GetNextSteps(stepID string) []string {
	hints, err := LoadHints()
	if err != nil {
		return []string{}
	}

	// Find the step across all workflows
	for _, workflow := range hints.Workflows {
		for _, step := range workflow.Steps {
			if step.ID == stepID {
				return step.NextSteps
			}
		}
	}

	return []string{}
}

// evaluateCondition evaluates a condition string against variables
func evaluateCondition(condition string, variables map[string]interface{}) bool {
	switch condition {
	case "always":
		return true
	case "success":
		if success, ok := variables["success"].(bool); ok {
			return success
		}
		return false
	case "!has_custom_icon":
		if hasIcon, ok := variables["has_custom_icon"].(bool); ok {
			return !hasIcon
		}
		return true
	case "!has_parent":
		if hasParent, ok := variables["has_parent"].(bool); ok {
			return !hasParent
		}
		return true
	case "has_suggestions":
		if suggestions, ok := variables["suggestions"]; ok {
			return suggestions != nil
		}
		return false
	case "has_references":
		if references, ok := variables["has_references"].(bool); ok {
			return references
		}
		return false
	case "has_conflicts":
		if conflicts, ok := variables["has_conflicts"].(bool); ok {
			return conflicts
		}
		return false
	case "has_results":
		if results, ok := variables["has_results"].(bool); ok {
			return results
		}
		return false
	case "no_results":
		if results, ok := variables["has_results"].(bool); ok {
			return !results
		}
		return false
	case "has_children":
		if children, ok := variables["has_children"].(bool); ok {
			return children
		}
		return false
	case "has_empty_types":
		if emptyTypes, ok := variables["has_empty_types"].(bool); ok {
			return emptyTypes
		}
		return false
	case "large_result_set":
		if count, ok := variables["result_count"].(int); ok {
			return count > 25
		}
		return false
	default:
		return false
	}
}

// substituteVariables replaces template variables in message strings
func substituteVariables(message string, variables map[string]interface{}, templates map[string]string) string {
	result := message
	
	// Replace template references like {set_icon_command}
	for key, template := range templates {
		placeholder := "{" + key + "}"
		if strings.Contains(result, placeholder) {
			// Substitute variables in the template
			substituted := substituteDirectVariables(template, variables)
			result = strings.ReplaceAll(result, placeholder, substituted)
		}
	}
	
	// Replace direct variable references
	result = substituteDirectVariables(result, variables)
	
	return result
}

// substituteDirectVariables replaces direct variable references like {object_type_name}
func substituteDirectVariables(text string, variables map[string]interface{}) string {
	result := text
	
	for key, value := range variables {
		placeholder := "{" + key + "}"
		if strings.Contains(result, placeholder) {
			result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
		}
	}
	
	return result
}

// sortHintsByPriority sorts hints by their priority
func sortHintsByPriority(messages []string, hints []Hint) []string {
	// Create a map of message to priority
	priorityMap := make(map[string]int)
	for i, hint := range hints {
		if i < len(messages) {
			switch hint.Priority {
			case "high":
				priorityMap[messages[i]] = 1
			case "medium":
				priorityMap[messages[i]] = 2
			case "low":
				priorityMap[messages[i]] = 3
			default:
				priorityMap[messages[i]] = 4
			}
		}
	}
	
	// Sort by priority (lower number = higher priority)
	// For now, return as-is since the complexity isn't worth it for the initial implementation
	return messages
}

// GetAvailableContexts returns all available context types
func GetAvailableContexts() []string {
	hints, err := LoadHints()
	if err != nil {
		return []string{}
	}

	var contexts []string
	for contextType := range hints.Contexts {
		contexts = append(contexts, contextType)
	}
	
	return contexts
}

// GetWorkflowVisualization returns a visualization of all workflows
func GetWorkflowVisualization() map[string]interface{} {
	hints, err := LoadHints()
	if err != nil {
		return map[string]interface{}{"error": "Failed to load hints"}
	}

	visualization := make(map[string]interface{})
	
	for workflowID, workflow := range hints.Workflows {
		steps := make([]map[string]interface{}, len(workflow.Steps))
		for i, step := range workflow.Steps {
			steps[i] = map[string]interface{}{
				"id":          step.ID,
				"name":        step.Name,
				"description": step.Description,
				"command":     step.Command,
				"next_steps":  step.NextSteps,
			}
		}
		
		visualization[workflowID] = map[string]interface{}{
			"description": workflow.Description,
			"steps":       steps,
		}
	}
	
	return visualization
}

// AI-specific loading functions
func LoadAIHints() (*AIHintsLibrary, error) {
	if globalAIHints != nil {
		return globalAIHints, nil
	}

	data, err := hintsFS.ReadFile("ai_workflow_hints.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read AI hints file: %w", err)
	}

	var aiHints AIHintsLibrary
	if err := json.Unmarshal(data, &aiHints); err != nil {
		return nil, fmt.Errorf("failed to parse AI hints JSON: %w", err)
	}

	globalAIHints = &aiHints
	return globalAIHints, nil
}

func LoadAIToolSemantics() (*AIToolSemanticsLibrary, error) {
	if globalAIToolSemantics != nil {
		return globalAIToolSemantics, nil
	}

	data, err := hintsFS.ReadFile("ai_tool_semantics.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read AI tool semantics file: %w", err)
	}

	var aiToolSemantics AIToolSemanticsLibrary
	if err := json.Unmarshal(data, &aiToolSemantics); err != nil {
		return nil, fmt.Errorf("failed to parse AI tool semantics JSON: %w", err)
	}

	globalAIToolSemantics = &aiToolSemantics
	return globalAIToolSemantics, nil
}

func LoadAIErrorRecovery() (*AIErrorRecoveryLibrary, error) {
	if globalAIErrorRecovery != nil {
		return globalAIErrorRecovery, nil
	}

	data, err := hintsFS.ReadFile("ai_error_recovery.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read AI error recovery file: %w", err)
	}

	var aiErrorRecovery AIErrorRecoveryLibrary
	if err := json.Unmarshal(data, &aiErrorRecovery); err != nil {
		return nil, fmt.Errorf("failed to parse AI error recovery JSON: %w", err)
	}

	globalAIErrorRecovery = &aiErrorRecovery
	return globalAIErrorRecovery, nil
}

// GetAIGuidance generates AI-specific guidance for a tool operation
func GetAIGuidance(toolName string, context map[string]interface{}) (*AIGuidance, error) {
	// Load AI tool semantics for context
	semantics, err := LoadAIToolSemantics()
	if err != nil {
		return nil, fmt.Errorf("failed to load AI tool semantics: %w", err)
	}

	// Load AI workflows for next steps
	workflows, err := LoadAIHints()
	if err != nil {
		return nil, fmt.Errorf("failed to load AI workflows: %w", err)
	}

	// Get tool-specific semantic information
	toolSemantic, exists := semantics.AIToolSemantics[toolName]
	if !exists {
		return &AIGuidance{
			OperationSummary: fmt.Sprintf("Executed %s operation", toolName),
			NextActions:      []AIAction{},
			WorkflowContext: WorkflowContext{
				CurrentState:         "operation_completed",
				CompletionPercentage: 100,
			},
			SemanticContext: context,
		}, nil
	}

	// Build operation summary
	operationSummary := buildOperationSummary(toolName, toolSemantic, context)

	// Determine next actions based on context and workflows
	nextActions := determineNextActions(toolName, toolSemantic, workflows, context)

	// Build workflow context
	workflowContext := buildWorkflowContext(toolName, context)

	return &AIGuidance{
		OperationSummary: operationSummary,
		NextActions:      nextActions,
		WorkflowContext:  workflowContext,
		SemanticContext:  context,
	}, nil
}

// GetAIErrorRecovery provides recovery suggestions for AI agents
func GetAIErrorRecovery(errorMessage string, context map[string]interface{}) ([]RecoveryAction, error) {
	errorRecovery, err := LoadAIErrorRecovery()
	if err != nil {
		return nil, fmt.Errorf("failed to load AI error recovery: %w", err)
	}

	// Search through error patterns
	for _, errorCategory := range errorRecovery.AIErrorRecovery {
		for _, errorInfo := range errorCategory {
			if strings.Contains(strings.ToLower(errorMessage), strings.ToLower(errorInfo.ErrorPattern)) {
				// Substitute context variables in recovery actions
				var recoveryActions []RecoveryAction
				for _, action := range errorInfo.RecoveryActions {
					substitutedAction := RecoveryAction{
						Tool:       action.Tool,
						Reason:     substituteDirectVariables(action.Reason, context),
						Confidence: action.Confidence,
						Parameters: substituteParameterVariables(action.Parameters, context),
					}
					recoveryActions = append(recoveryActions, substitutedAction)
				}
				return recoveryActions, nil
			}
		}
	}

	// Default recovery action if no specific pattern matches
	return []RecoveryAction{
		{
			Tool:       "assets_list_schemas",
			Reason:     "Check available schemas and permissions",
			Confidence: "medium",
			Parameters: map[string]interface{}{},
		},
	}, nil
}

// Helper functions for AI guidance
func buildOperationSummary(toolName string, toolSemantic AIToolSemantic, context map[string]interface{}) string {
	// Build a meaningful operation summary based on tool semantics and context
	if success, ok := context["success"].(bool); ok && success {
		if resultCount, ok := context["result_count"].(int); ok && resultCount > 0 {
			return fmt.Sprintf("Successfully executed %s - found %d results", toolName, resultCount)
		}
		return fmt.Sprintf("Successfully executed %s operation", toolName)
	}
	return fmt.Sprintf("Executed %s operation", toolName)
}

func determineNextActions(toolName string, toolSemantic AIToolSemantic, workflows *AIHintsLibrary, context map[string]interface{}) []AIAction {
	var nextActions []AIAction

	// Look for workflow-based next actions
	for _, workflow := range workflows.AIWorkflows {
		for _, step := range workflow.Steps {
			if step.ToolName == toolName {
				// Add next actions from the workflow
				for _, action := range step.NextActions {
					substitutedAction := AIAction{
						Tool:       action.Tool,
						Reason:     substituteDirectVariables(action.Reason, context),
						Confidence: action.Confidence,
						Parameters: substituteParameterVariables(action.Parameters, context),
					}
					nextActions = append(nextActions, substitutedAction)
				}
				break
			}
		}
	}

	// If no workflow actions found, provide semantic-based suggestions
	if len(nextActions) == 0 {
		nextActions = getSemanticBasedNextActions(toolName, toolSemantic, context)
	}

	return nextActions
}

func getSemanticBasedNextActions(toolName string, toolSemantic AIToolSemantic, context map[string]interface{}) []AIAction {
	var actions []AIAction

	// Provide context-appropriate next actions based on tool semantics
	switch toolSemantic.OperationType {
	case "foundation":
		if toolName == "assets_search" {
			if resultCount, ok := context["result_count"].(int); ok && resultCount > 0 {
				actions = append(actions, AIAction{
					Tool:       "assets_get",
					Reason:     "Get detailed information about the first result",
					Confidence: "high",
					Parameters: map[string]interface{}{
						"id": "{first_result_id}",
					},
				})
			}
		}
	case "composite":
		// Composite operations typically lead to validation or further analysis
		actions = append(actions, AIAction{
			Tool:       "assets_validate",
			Reason:     "Validate the results of the composite operation",
			Confidence: "medium",
			Parameters: map[string]interface{}{},
		})
	}

	return actions
}

func buildWorkflowContext(toolName string, context map[string]interface{}) WorkflowContext {
	// Determine workflow state based on tool name and context
	currentState := "operation_completed"
	completionPercentage := 100

	// Adjust based on tool semantics
	switch toolName {
	case "assets_search":
		currentState = "search_completed"
		completionPercentage = 25
	case "assets_create_object_type":
		currentState = "object_type_created"
		completionPercentage = 30
	case "assets_get":
		currentState = "object_analyzed"
		completionPercentage = 50
	}

	return WorkflowContext{
		CurrentState:         currentState,
		CompletionPercentage: completionPercentage,
		AvailableNextSteps:   []string{"validation", "enhancement", "analysis"},
		RequiredNextSteps:    []string{},
	}
}

func substituteParameterVariables(parameters map[string]interface{}, context map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range parameters {
		if strValue, ok := value.(string); ok {
			result[key] = substituteDirectVariables(strValue, context)
		} else {
			result[key] = value
		}
	}
	return result
}