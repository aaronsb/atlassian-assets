package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/aaronsb/atlassian-assets/internal/hints"
)

// WORKFLOWS command for exploring available workflows
var workflowsCmd = &cobra.Command{
	Use:   "workflows",
	Short: "Explore available workflows and hint system",
	Long: `Explore available workflows, contextual hints, and command relationships.
	
This command provides insight into the intelligent guidance system and helps
users understand available workflows and next-step possibilities.`,
}

// WORKFLOWS LIST subcommand
var workflowsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available workflows",
	Long: `List all available workflows with their descriptions and steps.`,
	Example: `  # List all workflows
  assets workflows list
  
  # Get detailed workflow information
  assets workflows show --workflow object_type_creation`,
	RunE: runWorkflowsListCmd,
}

func runWorkflowsListCmd(cmd *cobra.Command, args []string) error {
	visualization := hints.GetWorkflowVisualization()
	
	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"action":    "list_workflows",
			"workflows": visualization,
		},
	}
	
	return outputResult(response)
}

// WORKFLOWS SHOW subcommand
var workflowsShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show detailed workflow information",
	Long: `Show detailed information about a specific workflow including all steps and transitions.`,
	Example: `  # Show object type creation workflow
  assets workflows show --workflow object_type_creation
  
  # Show attribute marketplace workflow
  assets workflows show --workflow attribute_marketplace`,
	RunE: runWorkflowsShowCmd,
}

var workflowsShowWorkflow string

func init() {
	workflowsShowCmd.Flags().StringVar(&workflowsShowWorkflow, "workflow", "", "Workflow ID to show (required)")
	workflowsShowCmd.MarkFlagRequired("workflow")
}

func runWorkflowsShowCmd(cmd *cobra.Command, args []string) error {
	steps, err := hints.GetWorkflowSteps(workflowsShowWorkflow)
	if err != nil {
		return outputResult(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}
	
	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"action":      "show_workflow",
			"workflow_id": workflowsShowWorkflow,
			"steps":       steps,
		},
	}
	
	return outputResult(response)
}

// WORKFLOWS CONTEXTS subcommand
var workflowsContextsCmd = &cobra.Command{
	Use:   "contexts",
	Short: "List all available hint contexts",
	Long: `List all available hint contexts and their associated conditions.`,
	Example: `  # List all contexts
  assets workflows contexts`,
	RunE: runWorkflowsContextsCmd,
}

func runWorkflowsContextsCmd(cmd *cobra.Command, args []string) error {
	contexts := hints.GetAvailableContexts()
	
	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"action":   "list_contexts",
			"contexts": contexts,
			"count":    len(contexts),
		},
	}
	
	return outputResult(response)
}

// WORKFLOWS SIMULATE subcommand
var workflowsSimulateCmd = &cobra.Command{
	Use:   "simulate",
	Short: "Simulate contextual hints for testing",
	Long: `Simulate contextual hints for a given context and variables.
	
This is useful for testing the hints system and understanding what
hints would be shown in different scenarios.`,
	Example: `  # Simulate object type creation hints
  assets workflows simulate --context create_object_type --variables '{"object_type_name":"Test","has_custom_icon":false}'`,
	RunE: runWorkflowsSimulateCmd,
}

var (
	workflowsSimulateContext   string
	workflowsSimulateVariables string
)

func init() {
	workflowsSimulateCmd.Flags().StringVar(&workflowsSimulateContext, "context", "", "Context type (required)")
	workflowsSimulateCmd.Flags().StringVar(&workflowsSimulateVariables, "variables", "{}", "Variables as JSON")
	workflowsSimulateCmd.MarkFlagRequired("context")
}

func runWorkflowsSimulateCmd(cmd *cobra.Command, args []string) error {
	// Parse variables JSON
	var variables map[string]interface{}
	if err := json.Unmarshal([]byte(workflowsSimulateVariables), &variables); err != nil {
		return outputResult(map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("failed to parse variables JSON: %v", err),
		})
	}
	
	// Get contextual hints
	contextualHints := hints.GetContextualHints(workflowsSimulateContext, variables)
	
	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"action":    "simulate_hints",
			"context":   workflowsSimulateContext,
			"variables": variables,
			"hints":     contextualHints,
			"count":     len(contextualHints),
		},
	}
	
	return outputResult(response)
}

func init() {
	// Add subcommands to workflows command
	workflowsCmd.AddCommand(workflowsListCmd)
	workflowsCmd.AddCommand(workflowsShowCmd)
	workflowsCmd.AddCommand(workflowsContextsCmd)
	workflowsCmd.AddCommand(workflowsSimulateCmd)
}