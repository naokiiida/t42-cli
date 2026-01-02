package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/naokiiida/t42-cli/internal/api"
)

var userCmd = &cobra.Command{
	Use:     "user",
	Aliases: []string{"u"},
	Short:   "User management commands",
	Long: `Manage and query 42 users.

This command group allows you to list users, view user details,
and filter users by various criteria including campus location,
cursus progress, completed projects, and blackhole date.`,
}

var listUsersCmd = &cobra.Command{
	Use:   "list",
	Short: "List users",
	Long: `List users from the 42 API with filtering options.

You can filter users by:
  - Campus location (--campus or --campus-id)
  - Cursus (--cursus-id)
  - Completed projects count (--min-projects)
  - Blackhole date status (--blackhole-status: upcoming, past, active, none)
  - Active status (--active, --inactive)
  - Alumni status (--alumni, --non-alumni)
  - Staff status (--staff)

Examples:
  # List users from a specific campus
  t42 user list --campus-id 1

  # List users with upcoming blackhole
  t42 user list --blackhole-status upcoming --cursus-id 21

  # List active users with at least 10 completed projects
  t42 user list --active --min-projects 10

  # List users from Tokyo campus in 42cursus
  t42 user list --campus tokyo --cursus-id 21`,
	RunE: runListUsers,
}

var showUserCmd = &cobra.Command{
	Use:   "show <login>",
	Short: "Show user details",
	Long: `Show detailed information about a specific user.

You can specify a user by their login name (e.g., 'jdoe').`,
	Args: cobra.ExactArgs(1),
	RunE: runShowUser,
}

func init() {
	// Add user subcommands
	userCmd.AddCommand(listUsersCmd)
	userCmd.AddCommand(showUserCmd)

	// Add user command to root
	rootCmd.AddCommand(userCmd)

	// List command flags
	listUsersCmd.Flags().IntP("page", "p", 1, "Page number")
	listUsersCmd.Flags().Int("per-page", 20, "Number of users per page")
	listUsersCmd.Flags().Int("campus-id", 0, "Filter by campus ID")
	listUsersCmd.Flags().String("campus", "", "Filter by campus name (e.g., 'tokyo', 'paris')")
	listUsersCmd.Flags().Int("cursus-id", 0, "Filter by cursus ID (default: 21 for 42cursus)")
	listUsersCmd.Flags().StringP("sort", "s", "", "Sort by field (login, created_at, updated_at)")
	listUsersCmd.Flags().Bool("active", false, "Filter active users only")
	listUsersCmd.Flags().Bool("inactive", false, "Filter inactive users only")
	listUsersCmd.Flags().Bool("alumni", false, "Filter alumni only")
	listUsersCmd.Flags().Bool("non-alumni", false, "Filter non-alumni only")
	listUsersCmd.Flags().Bool("staff", false, "Filter staff only")
	listUsersCmd.Flags().Int("min-projects", 0, "Filter users with at least N completed projects")
	listUsersCmd.Flags().String("blackhole-status", "", "Filter by blackhole status (upcoming, past, active, none)")
	listUsersCmd.Flags().Int("blackhole-days", 30, "Number of days to consider for 'upcoming' blackhole status")
	listUsersCmd.Flags().Float64("min-level", 0, "Filter users with minimum cursus level")
	listUsersCmd.Flags().Float64("max-level", 0, "Filter users with maximum cursus level")
}

func runListUsers(cmd *cobra.Command, args []string) error {
	// Create API client with automatic token refresh
	client, err := NewAPIClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Get flags
	page, _ := cmd.Flags().GetInt("page")
	perPage, _ := cmd.Flags().GetInt("per-page")
	campusID, _ := cmd.Flags().GetInt("campus-id")
	campusName, _ := cmd.Flags().GetString("campus")
	cursusID, _ := cmd.Flags().GetInt("cursus-id")
	sort, _ := cmd.Flags().GetString("sort")
	active, _ := cmd.Flags().GetBool("active")
	inactive, _ := cmd.Flags().GetBool("inactive")
	alumni, _ := cmd.Flags().GetBool("alumni")
	nonAlumni, _ := cmd.Flags().GetBool("non-alumni")
	staff, _ := cmd.Flags().GetBool("staff")
	minProjects, _ := cmd.Flags().GetInt("min-projects")
	blackholeStatus, _ := cmd.Flags().GetString("blackhole-status")
	blackholeDays, _ := cmd.Flags().GetInt("blackhole-days")
	minLevel, _ := cmd.Flags().GetFloat64("min-level")
	maxLevel, _ := cmd.Flags().GetFloat64("max-level")

	// Resolve campus name to campus ID if provided
	if campusName != "" {
		campuses, err := client.ListCampuses(ctx)
		if err != nil {
			return fmt.Errorf("failed to list campuses: %w", err)
		}

		campusNameLower := strings.ToLower(campusName)
		for _, campus := range campuses {
			if strings.ToLower(campus.Name) == campusNameLower ||
			   strings.ToLower(campus.City) == campusNameLower {
				campusID = campus.ID
				break
			}
		}

		if campusID == 0 {
			return fmt.Errorf("campus '%s' not found", campusName)
		}
	}

	// Build options
	opts := &api.ListUsersOptions{
		Page:           page,
		PerPage:        perPage,
		FilterCampusID: campusID,
		FilterCursusID: cursusID,
		Sort:           sort,
	}

	// Handle active/inactive flags
	if active && !inactive {
		trueVal := true
		opts.FilterActive = &trueVal
	} else if inactive && !active {
		falseVal := false
		opts.FilterActive = &falseVal
	}

	// Handle alumni flags
	if alumni && !nonAlumni {
		trueVal := true
		opts.FilterAlumni = &trueVal
	} else if nonAlumni && !alumni {
		falseVal := false
		opts.FilterAlumni = &falseVal
	}

	// Handle staff flag
	if staff {
		trueVal := true
		opts.FilterStaff = &trueVal
	}

	// List users - use cursus_users endpoint when cursus filtering is needed for full data
	var users []api.User
	var meta *api.PaginationMeta

	// Use ListCursusUsers when we need level/blackhole data (cursus-id specified or filters requiring it)
	needsFullData := cursusID > 0 || minLevel > 0 || maxLevel > 0 || blackholeStatus != ""

	if needsFullData && cursusID > 0 {
		// Use cursus_users endpoint for full data (level, blackhole, etc.)
		cursusOpts := &api.ListCursusUsersOptions{
			Page:     page,
			PerPage:  perPage,
			CampusID: campusID,
			Sort:     sort,
		}

		cursusUsers, cursusMeta, err := client.ListCursusUsers(ctx, cursusID, cursusOpts)
		if err != nil {
			return fmt.Errorf("failed to list cursus users: %w", err)
		}

		// Convert CursusUser to User for filtering and display
		users = convertCursusUsersToUsers(cursusUsers, cursusID)
		meta = cursusMeta
	} else if campusID > 0 {
		users, meta, err = client.ListCampusUsers(ctx, campusID, opts)
	} else {
		users, meta, err = client.ListUsers(ctx, opts)
	}

	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	// Apply client-side filters
	filteredUsers := filterUsers(users, filterCriteria{
		minProjects:     minProjects,
		blackholeStatus: blackholeStatus,
		blackholeDays:   blackholeDays,
		cursusID:        cursusID,
		minLevel:        minLevel,
		maxLevel:        maxLevel,
	})

	if GetJSONOutput() {
		output := map[string]interface{}{
			"users": filteredUsers,
			"meta":  meta,
		}
		jsonData, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(jsonData))
	} else {
		printUsersTable(filteredUsers, meta, cursusID)
	}

	return nil
}

func runShowUser(cmd *cobra.Command, args []string) error {
	login := args[0]

	// Create API client with automatic token refresh
	client, err := NewAPIClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Get user by login
	user, err := client.GetUserByLogin(ctx, login)
	if err != nil {
		return fmt.Errorf("failed to get user '%s': %w", login, err)
	}

	if GetJSONOutput() {
		jsonData, _ := json.MarshalIndent(user, "", "  ")
		fmt.Println(string(jsonData))
	} else {
		printUserDetails(user)
	}

	return nil
}

type filterCriteria struct {
	minProjects     int
	blackholeStatus string
	blackholeDays   int
	cursusID        int
	minLevel        float64
	maxLevel        float64
}

// convertCursusUsersToUsers converts CursusUser objects to User objects for unified filtering and display
// This is needed because ListCursusUsers returns CursusUser with nested User, while other endpoints return User directly
func convertCursusUsersToUsers(cursusUsers []api.CursusUser, cursusID int) []api.User {
	users := make([]api.User, 0, len(cursusUsers))

	for _, cu := range cursusUsers {
		user := cu.User
		// Embed the cursus information into the user's CursusUsers slice
		user.CursusUsers = []api.CursusUser{
			{
				ID:           cu.ID,
				BeginAt:      cu.BeginAt,
				EndAt:        cu.EndAt,
				Grade:        cu.Grade,
				Level:        cu.Level,
				Skills:       cu.Skills,
				BlackholedAt: cu.BlackholedAt,
				Cursus:       cu.Cursus,
				HasCoalition: cu.HasCoalition,
			},
		}
		users = append(users, user)
	}

	return users
}

func filterUsers(users []api.User, criteria filterCriteria) []api.User {
	if criteria.minProjects == 0 && criteria.blackholeStatus == "" &&
	   criteria.minLevel == 0 && criteria.maxLevel == 0 {
		return users
	}

	filtered := make([]api.User, 0)
	now := time.Now()

	for _, user := range users {
		// Filter by completed projects
		if criteria.minProjects > 0 {
			completedCount := countCompletedProjects(user.ProjectsUsers)
			if completedCount < criteria.minProjects {
				continue
			}
		}

		// Filter by blackhole status
		if criteria.blackholeStatus != "" {
			cursusUser := findCursusUser(user.CursusUsers, criteria.cursusID)
			if !matchesBlackholeStatus(cursusUser, criteria.blackholeStatus, criteria.blackholeDays, now) {
				continue
			}
		}

		// Filter by level
		if criteria.minLevel > 0 || criteria.maxLevel > 0 {
			cursusUser := findCursusUser(user.CursusUsers, criteria.cursusID)
			if cursusUser == nil {
				continue
			}
			if criteria.minLevel > 0 && cursusUser.Level < criteria.minLevel {
				continue
			}
			if criteria.maxLevel > 0 && cursusUser.Level > criteria.maxLevel {
				continue
			}
		}

		filtered = append(filtered, user)
	}

	return filtered
}

func countCompletedProjects(projectUsers []api.ProjectUser) int {
	count := 0
	for _, pu := range projectUsers {
		if pu.Validated != nil && *pu.Validated {
			count++
		}
	}
	return count
}

func findCursusUser(cursusUsers []api.CursusUser, cursusID int) *api.CursusUser {
	if cursusID == 0 && len(cursusUsers) > 0 {
		// If no cursus ID specified, use the most recent active cursus
		for i := len(cursusUsers) - 1; i >= 0; i-- {
			if cursusUsers[i].EndAt == nil || cursusUsers[i].EndAt.After(time.Now()) {
				return &cursusUsers[i]
			}
		}
		return &cursusUsers[len(cursusUsers)-1]
	}

	for i := range cursusUsers {
		if cursusUsers[i].Cursus.ID == cursusID {
			return &cursusUsers[i]
		}
	}
	return nil
}

func matchesBlackholeStatus(cursusUser *api.CursusUser, status string, days int, now time.Time) bool {
	if cursusUser == nil {
		return status == "none"
	}

	switch status {
	case "none":
		return cursusUser.BlackholedAt == nil
	case "active":
		return cursusUser.BlackholedAt != nil && cursusUser.EndAt == nil
	case "past":
		return cursusUser.BlackholedAt != nil && cursusUser.BlackholedAt.Before(now)
	case "upcoming":
		if cursusUser.BlackholedAt == nil {
			return false
		}
		threshold := now.AddDate(0, 0, days)
		return cursusUser.BlackholedAt.After(now) && cursusUser.BlackholedAt.Before(threshold)
	default:
		return true
	}
}

func printUsersTable(users []api.User, meta *api.PaginationMeta, cursusID int) {
	if len(users) == 0 {
		fmt.Println("No users found.")
		return
	}

	// Header
	fmt.Printf("%-20s %-30s %-15s %-10s %-10s %s\n",
		"LOGIN", "NAME", "CAMPUS", "LEVEL", "PROJECTS", "BLACKHOLE")
	fmt.Printf("%s\n", strings.Repeat("-", 110))

	// Users
	for _, user := range users {
		login := truncateString(user.Login, 18)
		displayName := truncateString(user.DisplayName, 28)

		campus := "N/A"
		if len(user.Campus) > 0 {
			campus = truncateString(user.Campus[0].City, 13)
		}

		level := "N/A"
		blackhole := "N/A"

		// Find cursus user
		cursusUser := findCursusUser(user.CursusUsers, cursusID)
		if cursusUser != nil {
			level = fmt.Sprintf("%.2f", cursusUser.Level)

			if cursusUser.BlackholedAt != nil {
				daysUntil := int(time.Until(*cursusUser.BlackholedAt).Hours() / 24)
				if daysUntil > 0 {
					blackhole = fmt.Sprintf("%dd", daysUntil)
				} else {
					blackhole = "BH'd"
				}
			} else {
				blackhole = "-"
			}
		}

		projectCount := strconv.Itoa(countCompletedProjects(user.ProjectsUsers))

		fmt.Printf("%-20s %-30s %-15s %-10s %-10s %s\n",
			login, displayName, campus, level, projectCount, blackhole)
	}

	// Pagination info
	if meta != nil {
		fmt.Printf("\n📄 Page %d of %d (%d total users, showing %d)\n",
			meta.Page, meta.TotalPages, meta.TotalCount, len(users))
		if meta.Page < meta.TotalPages {
			fmt.Printf("   Use --page %d to see the next page\n", meta.Page+1)
		}
	}
}

func printUserDetails(user *api.User) {
	fmt.Printf("👤 User: %s (%s)\n", user.DisplayName, user.Login)
	fmt.Printf("📧 Email: %s\n", user.Email)

	if len(user.Campus) > 0 {
		fmt.Printf("🏫 Campus: %s (%s)\n", user.Campus[0].Name, user.Campus[0].City)
	}

	fmt.Printf("⚡ Correction Points: %d\n", user.CorrectionPoint)
	fmt.Printf("💰 Wallet: %d\n", user.Wallet)

	if user.PoolMonth != "" && user.PoolYear != "" {
		fmt.Printf("🏊 Pool: %s %s\n", user.PoolMonth, user.PoolYear)
	}

	if user.Active {
		fmt.Printf("✅ Status: Active\n")
	} else {
		fmt.Printf("❌ Status: Inactive\n")
	}

	if user.Alumni {
		fmt.Printf("🎓 Alumni: Yes\n")
	}

	if user.Staff {
		fmt.Printf("👔 Staff: Yes\n")
	}

	// Cursus information
	if len(user.CursusUsers) > 0 {
		fmt.Printf("\n📚 Cursus:\n")
		for _, cu := range user.CursusUsers {
			fmt.Printf("   • %s: Level %.2f", cu.Cursus.Name, cu.Level)
			if cu.Grade != nil {
				fmt.Printf(" (Grade: %s)", *cu.Grade)
			}
			if cu.BlackholedAt != nil {
				daysUntil := int(time.Until(*cu.BlackholedAt).Hours() / 24)
				if daysUntil > 0 {
					fmt.Printf(" - Blackhole in %d days", daysUntil)
				} else {
					fmt.Printf(" - Blackholed")
				}
			}
			fmt.Println()
		}
	}

	// Projects summary
	completedCount := countCompletedProjects(user.ProjectsUsers)
	fmt.Printf("\n🎯 Projects: %d completed\n", completedCount)

	fmt.Printf("\n📅 Created: %s\n", user.CreatedAt.Format(time.RFC3339))
	fmt.Printf("🔄 Updated: %s\n", user.UpdatedAt.Format(time.RFC3339))
}
