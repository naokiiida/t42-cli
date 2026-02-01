package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/naokiiida/t42-cli/internal/api"
)

var campusCmd = &cobra.Command{
	Use:     "campus",
	Aliases: []string{"c"},
	Short:   "Campus management commands",
	Long: `Query 42 campuses.

This command group allows you to list all campuses and search for specific ones.`,
}

var listCampusesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all campuses",
	Long: `List all 42 campuses with their IDs.

Examples:
  # List all campuses
  t42 campus list

  # Search for a specific campus
  t42 campus list --search tokyo

  # Output in JSON format
  t42 campus list --json`,
	RunE: runListCampuses,
}

var showCampusCmd = &cobra.Command{
	Use:   "show <id-or-name>",
	Short: "Show campus details",
	Long: `Show detailed information about a specific campus.

You can specify a campus by ID or name.`,
	Args: cobra.ExactArgs(1),
	RunE: runShowCampus,
}

func init() {
	// Add campus subcommands
	campusCmd.AddCommand(listCampusesCmd)
	campusCmd.AddCommand(showCampusCmd)

	// Add campus command to root
	rootCmd.AddCommand(campusCmd)

	// List command flags
	listCampusesCmd.Flags().String("search", "", "Search campuses by name or city")
	listCampusesCmd.Flags().Bool("active-only", false, "Show only active campuses")
}

func runListCampuses(cmd *cobra.Command, args []string) error {
	client, err := NewAPIClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	search, _ := cmd.Flags().GetString("search")
	activeOnly, _ := cmd.Flags().GetBool("active-only")

	campuses, err := client.ListCampuses(ctx)
	if err != nil {
		return fmt.Errorf("failed to list campuses: %w", err)
	}

	// Filter campuses
	filtered := make([]api.Campus, 0)
	searchLower := strings.ToLower(search)

	for _, c := range campuses {
		// Filter by active status
		if activeOnly && !c.Active {
			continue
		}

		// Filter by search term
		if search != "" {
			nameLower := strings.ToLower(c.Name)
			cityLower := strings.ToLower(c.City)
			countryLower := strings.ToLower(c.Country)

			if !strings.Contains(nameLower, searchLower) &&
				!strings.Contains(cityLower, searchLower) &&
				!strings.Contains(countryLower, searchLower) {
				continue
			}
		}

		filtered = append(filtered, c)
	}

	if GetJSONOutput() {
		output := map[string]interface{}{
			"campuses": filtered,
			"count":    len(filtered),
		}
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON output: %w", err)
		}
		fmt.Println(string(jsonData))
	} else {
		if len(filtered) == 0 {
			fmt.Println("No campuses found matching criteria.")
			return nil
		}

		fmt.Printf("%-6s %-25s %-20s %-15s %s\n", "ID", "NAME", "CITY", "COUNTRY", "ACTIVE")
		fmt.Println(strings.Repeat("-", 80))
		for _, c := range filtered {
			activeStr := "No"
			if c.Active {
				activeStr = "Yes"
			}
			fmt.Printf("%-6d %-25s %-20s %-15s %s\n",
				c.ID,
				truncateString(c.Name, 25),
				truncateString(c.City, 20),
				truncateString(c.Country, 15),
				activeStr)
		}
		fmt.Printf("\nTotal: %d campuses\n", len(filtered))
	}

	return nil
}

func runShowCampus(cmd *cobra.Command, args []string) error {
	client, err := NewAPIClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	query := args[0]

	campuses, err := client.ListCampuses(ctx)
	if err != nil {
		return fmt.Errorf("failed to list campuses: %w", err)
	}

	var found *api.Campus

	// Try to find by ID first
	if id, err := strconv.Atoi(query); err == nil {
		for i := range campuses {
			if campuses[i].ID == id {
				found = &campuses[i]
				break
			}
		}
	}

	// If not found by ID, search by name
	if found == nil {
		queryLower := strings.ToLower(query)
		for i := range campuses {
			if strings.ToLower(campuses[i].Name) == queryLower ||
				strings.ToLower(campuses[i].City) == queryLower {
				found = &campuses[i]
				break
			}
		}
	}

	if found == nil {
		return fmt.Errorf("campus %q not found", query)
	}

	if GetJSONOutput() {
		jsonData, err := json.MarshalIndent(found, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON output: %w", err)
		}
		fmt.Println(string(jsonData))
	} else {
		printCampusDetails(found)
	}

	return nil
}

func printCampusDetails(c *api.Campus) {
	fmt.Printf("Campus: %s (ID: %d)\n", c.Name, c.ID)
	fmt.Println(strings.Repeat("=", 40))

	fmt.Printf("City:       %s\n", c.City)
	fmt.Printf("Country:    %s\n", c.Country)
	fmt.Printf("Address:    %s\n", c.Address)
	if c.Zip != "" {
		fmt.Printf("ZIP:        %s\n", c.Zip)
	}
	fmt.Printf("Timezone:   %s\n", c.TimeZone)
	fmt.Printf("Users:      %d\n", c.UsersCount)

	activeStr := "No"
	if c.Active {
		activeStr = "Yes"
	}
	fmt.Printf("Active:     %s\n", activeStr)

	if c.Website != "" {
		fmt.Printf("Website:    %s\n", c.Website)
	}
}
