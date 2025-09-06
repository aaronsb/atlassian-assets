package foundation

import (
	"fmt"
	"strings"

	"github.com/aaronsb/atlassian-assets/cmd/assets/common"
)

// Icon represents an available icon in Assets
type Icon struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// GetAvailableIcons retrieves available icons (hardcoded common set)
func GetAvailableIcons() []Icon {
	// Common icon IDs used in Atlassian Assets
	return []Icon{
		{ID: "1", Name: "Generic"},
		{ID: "2", Name: "Server"},
		{ID: "3", Name: "Desktop Computer"},
		{ID: "4", Name: "Laptop"},
		{ID: "5", Name: "Mobile Device"},
		{ID: "6", Name: "Printer"},
		{ID: "7", Name: "Network Device"},
		{ID: "8", Name: "Storage"},
		{ID: "9", Name: "Virtual Machine"},
		{ID: "10", Name: "Cloud"},
		{ID: "11", Name: "Database"},
		{ID: "12", Name: "Application"},
		{ID: "13", Name: "Service"},
		{ID: "14", Name: "User"},
		{ID: "15", Name: "Group"},
		{ID: "16", Name: "Location"},
		{ID: "17", Name: "Building"},
		{ID: "18", Name: "Department"},
		{ID: "19", Name: "Contract"},
		{ID: "20", Name: "License"},
		{ID: "143", Name: "Computer/Workstation"},
		{ID: "144", Name: "Server Rack"},
		{ID: "145", Name: "Network Router"},
		{ID: "146", Name: "Firewall"},
		{ID: "147", Name: "Switch"},
		{ID: "148", Name: "Wireless Access Point"},
		{ID: "149", Name: "Monitor"},
		{ID: "150", Name: "Keyboard"},
		{ID: "151", Name: "Mouse"},
		{ID: "152", Name: "Headset"},
		{ID: "153", Name: "Phone"},
		{ID: "154", Name: "Tablet"},
		{ID: "155", Name: "Scanner"},
		{ID: "156", Name: "Camera"},
		{ID: "157", Name: "Projector"},
		{ID: "158", Name: "Software"},
		{ID: "159", Name: "Operating System"},
		{ID: "160", Name: "Backup"},
		{ID: "161", Name: "Security"},
		{ID: "162", Name: "Certificate"},
		{ID: "163", Name: "Key"},
		{ID: "164", Name: "Lock"},
		{ID: "165", Name: "Shield"},
		{ID: "166", Name: "Warning"},
		{ID: "167", Name: "Information"},
		{ID: "168", Name: "Success"},
		{ID: "169", Name: "Error"},
		{ID: "170", Name: "Tool"},
		{ID: "171", Name: "Settings"},
		{ID: "172", Name: "Configuration"},
		{ID: "173", Name: "Process"},
		{ID: "174", Name: "Workflow"},
		{ID: "175", Name: "Document"},
		{ID: "176", Name: "Folder"},
		{ID: "177", Name: "Archive"},
		{ID: "178", Name: "Email"},
		{ID: "179", Name: "Message"},
		{ID: "180", Name: "Chat"},
	}
}

// ListIcons lists available icons with optional search filtering
func ListIcons(client common.ClientInterface, params common.ListIconsParams) (*common.Response, error) {
	icons := GetAvailableIcons()

	// Filter by search term if provided
	if params.Search != "" {
		var filtered []Icon
		searchLower := strings.ToLower(params.Search)
		for _, icon := range icons {
			if strings.Contains(strings.ToLower(icon.Name), searchLower) {
				filtered = append(filtered, icon)
			}
		}
		icons = filtered
	}

	// Apply limit
	if params.Limit > 0 && len(icons) > params.Limit {
		icons = icons[:params.Limit]
	}

	return common.NewSuccessResponse(map[string]interface{}{
		"icons":     icons,
		"total":     len(icons),
		"search":    params.Search,
		"operation": "list_icons",
	}), nil
}

// SuggestIcons suggests relevant icons based on asset type name
func SuggestIcons(client common.ClientInterface, params common.SuggestIconsParams) (*common.Response, error) {
	if params.TypeName == "" {
		return common.NewErrorResponse(fmt.Errorf("type name is required")), nil
	}

	suggestions := suggestIconsForType(strings.ToLower(params.TypeName))
	
	return common.NewSuccessResponse(map[string]interface{}{
		"type_name":    params.TypeName,
		"suggestions":  suggestions,
		"recommended":  suggestions[0],
		"total":        len(suggestions),
		"operation":    "suggest_icons",
	}), nil
}

// suggestIconsForType suggests relevant icons based on type name keywords
func suggestIconsForType(typeName string) []Icon {
	allIcons := GetAvailableIcons()
	var suggestions []Icon

	// Keyword mappings to icon IDs
	keywordMap := map[string][]string{
		// Hardware
		"workstation": {"143", "3", "4"},
		"computer":    {"143", "3", "4"},
		"laptop":      {"4", "143", "3"},
		"desktop":     {"3", "143"},
		"server":      {"2", "144", "9"},
		"rack":        {"144", "2"},
		"storage":     {"8", "160", "177"},
		"backup":      {"160", "8", "177"},
		
		// Networking
		"network":     {"7", "145", "147"},
		"router":      {"145", "7"},
		"switch":      {"147", "7"},
		"firewall":    {"146", "161", "165"},
		"wireless":    {"148", "7"},
		"wifi":        {"148", "7"},
		
		// Mobile & Devices
		"mobile":      {"5", "154", "153"},
		"phone":       {"153", "5"},
		"tablet":      {"154", "5"},
		"device":      {"5", "7", "170"},
		
		// Peripherals
		"printer":     {"6", "155"},
		"scanner":     {"155", "6"},
		"monitor":     {"149"},
		"keyboard":    {"150"},
		"mouse":       {"151"},
		"headset":     {"152"},
		"camera":      {"156"},
		"projector":   {"157"},
		
		// Software & Services
		"software":    {"158", "12", "159"},
		"application": {"12", "158"},
		"app":         {"12", "158"},
		"service":     {"13", "173"},
		"database":    {"11", "8"},
		"db":          {"11", "8"},
		"system":      {"159", "171", "172"},
		"os":          {"159"},
		
		// Cloud & Virtual
		"cloud":       {"10", "9", "13"},
		"virtual":     {"9", "10"},
		"vm":          {"9", "10"},
		"container":   {"9", "10", "13"},
		
		// Security
		"security":    {"161", "165", "164"},
		"certificate": {"162", "161"},
		"key":         {"163", "164"},
		"lock":        {"164", "161"},
		"shield":      {"165", "161"},
		"auth":        {"163", "164", "161"},
		
		// Organization
		"user":        {"14"},
		"person":      {"14"},
		"group":       {"15", "18"},
		"team":        {"15", "18"},
		"department":  {"18", "15"},
		"location":    {"16", "17"},
		"building":    {"17", "16"},
		"office":      {"17", "16"},
		
		// Documents & Contracts
		"contract":    {"19", "175"},
		"license":     {"20", "162"},
		"document":    {"175", "176"},
		"file":        {"175", "176"},
		"folder":      {"176", "177"},
		"archive":     {"177", "176"},
		
		// Communication
		"email":       {"178", "179"},
		"message":     {"179", "180"},
		"chat":        {"180", "179"},
		
		// AI/ML specific
		"ai":          {"143", "2", "9", "11"},
		"ml":          {"143", "2", "9", "11"},
		"gpu":         {"143", "2", "8"},
		"training":    {"143", "2", "173"},
		"model":       {"11", "173", "12"},
		
		// Process & Config
		"process":     {"173", "174"},
		"workflow":    {"174", "173"},
		"config":      {"172", "171"},
		"settings":    {"171", "172"},
		"tool":        {"170", "171"},
	}

	// Check each keyword
	foundIconIDs := make(map[string]bool)
	for keyword, iconIDs := range keywordMap {
		if strings.Contains(typeName, keyword) {
			for _, id := range iconIDs {
				if !foundIconIDs[id] {
					// Find the icon with this ID
					for _, icon := range allIcons {
						if icon.ID == id {
							suggestions = append(suggestions, icon)
							foundIconIDs[id] = true
							break
						}
					}
				}
			}
		}
	}

	// If no specific matches, suggest generic options
	if len(suggestions) == 0 {
		suggestions = append(suggestions, Icon{ID: "1", Name: "Generic"})
		suggestions = append(suggestions, Icon{ID: "170", Name: "Tool"})
		suggestions = append(suggestions, Icon{ID: "171", Name: "Settings"})
	}

	return suggestions
}

// GetIconForType returns the best icon ID for a given type name
func GetIconForType(typeName string) string {
	suggestions := suggestIconsForType(strings.ToLower(typeName))
	if len(suggestions) > 0 {
		return suggestions[0].ID
	}
	return "1" // Default to generic icon
}