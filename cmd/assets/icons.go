package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	iconSearchTerm string
	iconLimit      int
	iconFormat     string
)

var iconsCmd = &cobra.Command{
	Use:   "icons",
	Short: "Manage and discover icons for asset types",
	Long: `View and search available icons that can be used when creating object types.
Icons help visually distinguish different types of assets in the Atlassian Assets interface.`,
	Example: `  # List all available icons
  assets icons list

  # Search for icons by keyword
  assets icons list --search "computer"
  assets icons list --search "server"

  # Show suggested icons for a type name
  assets icons suggest "AI Workstation"
  assets icons suggest "Network Router"`,
}

var iconsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available icons",
	Long: `List all available icons that can be used when creating object types.
Icons are identified by their ID number which can be used with the --icon flag when creating object types.`,
	Example: `  # List all icons
  assets icons list

  # Search for specific icons
  assets icons list --search "computer"
  
  # Limit results
  assets icons list --limit 20`,
	RunE: runIconsListCmd,
}

var iconsSuggestCmd = &cobra.Command{
	Use:   "suggest [type-name]",
	Short: "Suggest relevant icons for an asset type",
	Long: `Get smart icon suggestions based on the asset type name.
This command analyzes the type name and suggests the most relevant icons from the available set.`,
	Example: `  # Get icon suggestions for different asset types
  assets icons suggest "AI Workstation"
  assets icons suggest "Database Server"
  assets icons suggest "Network Switch"
  assets icons suggest "Software License"`,
	Args: cobra.ExactArgs(1),
	RunE: runIconsSuggestCmd,
}

func init() {
	rootCmd.AddCommand(iconsCmd)
	iconsCmd.AddCommand(iconsListCmd)
	iconsCmd.AddCommand(iconsSuggestCmd)

	// List command flags
	iconsListCmd.Flags().StringVar(&iconSearchTerm, "search", "", "Search icons by keyword")
	iconsListCmd.Flags().IntVar(&iconLimit, "limit", 50, "Maximum number of icons to display")
	iconsListCmd.Flags().StringVarP(&iconFormat, "output", "o", "table", "Output format (table, json, ids)")
}

// Icon represents an available icon in Assets
type Icon struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// getAvailableIcons retrieves available icons (hardcoded common set for now)
func getAvailableIcons() []Icon {
	// Common icon IDs used in Atlassian Assets
	// These are standard icons that are typically available
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

func runIconsListCmd(cmd *cobra.Command, args []string) error {
	icons := getAvailableIcons()

	// Filter by search term if provided
	if iconSearchTerm != "" {
		var filtered []Icon
		searchLower := strings.ToLower(iconSearchTerm)
		for _, icon := range icons {
			if strings.Contains(strings.ToLower(icon.Name), searchLower) {
				filtered = append(filtered, icon)
			}
		}
		icons = filtered
	}

	// Apply limit
	if iconLimit > 0 && len(icons) > iconLimit {
		icons = icons[:iconLimit]
	}

	// Output based on format
	switch iconFormat {
	case "json":
		data, _ := json.MarshalIndent(icons, "", "  ")
		fmt.Println(string(data))
	case "ids":
		for _, icon := range icons {
			fmt.Printf("%s: %s\n", icon.ID, icon.Name)
		}
	default: // table
		if len(icons) == 0 {
			fmt.Println("No icons found matching criteria")
			return nil
		}

		headerColor := color.New(color.FgCyan, color.Bold)
		headerColor.Println("\nAvailable Icons")
		fmt.Println(strings.Repeat("-", 40))

		for _, icon := range icons {
			fmt.Printf("ID: %-6s Name: %s\n", 
				color.YellowString(icon.ID), 
				color.GreenString(icon.Name))
		}

		fmt.Printf("\nTotal: %d icons\n", len(icons))
		
		// Display hints
		fmt.Println("\n💡 Next steps:")
		fmt.Println("  - Search for specific icons: assets icons list --search \"computer\"")
		fmt.Println("  - Get icon suggestions: assets icons suggest \"AI Workstation\"")
		fmt.Println("  - Use icon when creating types: assets schema create-type --icon 143")
	}

	return nil
}

func runIconsSuggestCmd(cmd *cobra.Command, args []string) error {
	typeName := args[0]
	typeNameLower := strings.ToLower(typeName)
	
	// Icon suggestion mapping based on common keywords
	suggestions := suggestIconsForType(typeNameLower)
	
	if len(suggestions) == 0 {
		fmt.Println("No specific icon suggestions found. Using default icon.")
		fmt.Println("\nDefault icon: ID 1 (Generic)")
		return nil
	}

	headerColor := color.New(color.FgCyan, color.Bold)
	headerColor.Printf("\nSuggested Icons for '%s'\n", typeName)
	fmt.Println(strings.Repeat("-", 50))

	fmt.Println("\nTop Recommendations:")
	for i, icon := range suggestions {
		if i >= 3 {
			break // Show top 3 recommendations
		}
		recommended := ""
		if i == 0 {
			recommended = color.GreenString(" [RECOMMENDED]")
		}
		fmt.Printf("  %d. ID: %-6s Name: %s%s\n", 
			i+1,
			color.YellowString(icon.ID), 
			color.CyanString(icon.Name),
			recommended)
	}

	if len(suggestions) > 3 {
		fmt.Println("\nOther Options:")
		for i := 3; i < len(suggestions); i++ {
			fmt.Printf("  - ID: %-6s Name: %s\n", 
				color.YellowString(suggestions[i].ID), 
				suggestions[i].Name)
		}
	}

	fmt.Printf("\nUsage: assets schema create-type --icon %s --name \"%s\" ...\n", 
		color.YellowString(suggestions[0].ID), typeName)
	
	// Display hints
	fmt.Println("\n💡 Next steps:")
	fmt.Printf("  - Create object type: assets schema create-type --icon %s --name \"%s\" --schema [SCHEMA_ID]\n", suggestions[0].ID, typeName)
	fmt.Println("  - View all icons: assets icons list")
	fmt.Println("  - Try different search: assets icons suggest \"network device\"")

	return nil
}

// suggestIconsForType suggests relevant icons based on type name keywords
func suggestIconsForType(typeName string) []Icon {
	allIcons := getAvailableIcons()
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