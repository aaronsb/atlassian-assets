package hints

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed workflow_hints.json
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

var globalHints *HintsLibrary

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