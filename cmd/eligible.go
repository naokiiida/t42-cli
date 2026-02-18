package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/naokiiida/t42-cli/internal/api"
	"github.com/naokiiida/t42-cli/internal/config"
)

var eligibleCmd = &cobra.Command{
	Use:   "eligible",
	Short: "Find users eligible for a project",
	Long: `Find users who meet the inscription requirements for a project.

This command reads the project's session rules from the API and checks
each user against the inscription requirements (quests validated, quests
not validated, projects not ongoing/validated).

By default, blackholed users are excluded. Users must have an active
cursus (not ended) to be considered eligible.

Examples:
  # Find users eligible for ft_transcendence at Tokyo campus
  t42 user eligible --project ft_transcendence --campus tokyo

  # With level range filter
  t42 user eligible --project ft_transcendence --campus tokyo --min-level 6 --max-level 9

  # Show more results
  t42 user eligible --project ft_transcendence --campus tokyo --limit 10

  # JSON output
  t42 user eligible --project ft_transcendence --campus tokyo --json`,
	RunE: runEligible,
}

func init() {
	eligibleCmd.Flags().String("project", "", "Project slug (required, e.g., ft_transcendence)")
	eligibleCmd.Flags().String("campus", "", "Campus name (e.g., tokyo)")
	eligibleCmd.Flags().Int("campus-id", 0, "Campus ID")
	eligibleCmd.Flags().Int("cursus-id", 21, "Cursus ID (default: 21 for 42cursus)")
	eligibleCmd.Flags().Float64("min-level", 0, "Minimum cursus level")
	eligibleCmd.Flags().Float64("max-level", 0, "Maximum cursus level")
	eligibleCmd.Flags().IntP("limit", "l", 5, "Maximum number of eligible users to find")

	if err := eligibleCmd.MarkFlagRequired("project"); err != nil {
		panic(fmt.Sprintf("failed to mark project flag required: %v", err))
	}

	userCmd.AddCommand(eligibleCmd)
}

// inscriptionRequirements holds the parsed inscription rules from a project session
type inscriptionRequirements struct {
	requiredQuests    []string // quest slugs that must be validated
	forbiddenQuests   []string // quest slugs that must NOT be validated
	forbiddenProjects []string // project slugs that must NOT be ongoing or validated
}

// parseInscriptionRules extracts inscription requirements from project session rules
func parseInscriptionRules(rules []api.ProjectSessionRule) inscriptionRequirements {
	var reqs inscriptionRequirements

	for _, rule := range rules {
		if rule.Rule.Kind != "inscription" {
			continue
		}

		switch rule.Rule.InternalName {
		case "QuestsValidated":
			// Params contain quest slugs that must be validated
			for _, p := range rule.Params {
				reqs.requiredQuests = append(reqs.requiredQuests, p.Value)
			}
		case "QuestsNotValidated":
			// Params contain quest slugs that must NOT be validated
			for _, p := range rule.Params {
				reqs.forbiddenQuests = append(reqs.forbiddenQuests, p.Value)
			}
		case "NeitherOngoingOrValidated":
			// Params contain project slugs that must NOT be ongoing or validated
			for _, p := range rule.Params {
				reqs.forbiddenProjects = append(reqs.forbiddenProjects, p.Value)
			}
		}
	}

	return reqs
}

// checkRequiredQuests returns true if the user has validated all required quests
func checkRequiredQuests(questUsers []api.QuestUser, required []string) bool {
	if len(required) == 0 {
		return true
	}

	validated := make(map[string]bool)
	for _, qu := range questUsers {
		if qu.ValidatedAt != nil {
			validated[qu.Quest.Slug] = true
		}
	}

	for _, slug := range required {
		if !validated[slug] {
			return false
		}
	}
	return true
}

// checkForbiddenQuests returns true if the user has NOT validated any forbidden quests
func checkForbiddenQuests(questUsers []api.QuestUser, forbidden []string) bool {
	if len(forbidden) == 0 {
		return true
	}

	forbiddenSet := make(map[string]bool)
	for _, slug := range forbidden {
		forbiddenSet[slug] = true
	}

	for _, qu := range questUsers {
		if qu.ValidatedAt != nil && forbiddenSet[qu.Quest.Slug] {
			return false
		}
	}
	return true
}

// checkForbiddenProjects returns true if the user does NOT have any forbidden projects ongoing or validated
func checkForbiddenProjects(projectUsers []api.ProjectUser, forbidden []string) bool {
	if len(forbidden) == 0 {
		return true
	}

	forbiddenSet := make(map[string]bool)
	for _, slug := range forbidden {
		forbiddenSet[slug] = true
	}

	for _, pu := range projectUsers {
		if !forbiddenSet[pu.Project.Slug] {
			continue
		}
		// Check if the project is ongoing or validated
		if pu.Status == "finished" && pu.Validated != nil && *pu.Validated {
			return false // already validated
		}
		if pu.Status == "in_progress" || pu.Status == "waiting_for_correction" || pu.Status == "creating_group" || pu.Status == "searching_a_group" {
			return false // ongoing
		}
	}
	return true
}

// eligibleUser represents a user that passed all eligibility checks
type eligibleUser struct {
	User       api.User       `json:"user"`
	Level      float64        `json:"level"`
	BlackholeD int            `json:"blackhole_days"`
	QuestsInfo []questInfo    `json:"quests_validated"`
}

type questInfo struct {
	Slug        string `json:"slug"`
	ValidatedAt string `json:"validated_at"`
}

func runEligible(cmd *cobra.Command, args []string) error {
	client, err := NewAPIClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Get flags
	projectSlug, _ := cmd.Flags().GetString("project")
	campusName, _ := cmd.Flags().GetString("campus")
	campusID, _ := cmd.Flags().GetInt("campus-id")
	cursusID, _ := cmd.Flags().GetInt("cursus-id")
	minLevel, _ := cmd.Flags().GetFloat64("min-level")
	maxLevel, _ := cmd.Flags().GetFloat64("max-level")
	limit, _ := cmd.Flags().GetInt("limit")

	// Resolve campus name to ID
	var resolvedCampus *api.Campus
	if campusName != "" {
		campuses, campusErr := client.ListCampuses(ctx)
		if campusErr != nil {
			return fmt.Errorf("failed to list campuses: %w", campusErr)
		}

		campusNameLower := strings.ToLower(campusName)
		for i := range campuses {
			if strings.ToLower(campuses[i].Name) == campusNameLower ||
				strings.ToLower(campuses[i].City) == campusNameLower {
				campusID = campuses[i].ID
				resolvedCampus = &campuses[i]
				break
			}
		}

		if campusID == 0 {
			var campusOptions []string
			for _, campus := range campuses {
				label := campus.Name
				cityLower := strings.ToLower(campus.City)
				nameLower := strings.ToLower(campus.Name)
				if campus.City != "" && cityLower != nameLower {
					label = fmt.Sprintf("%s (%s)", campus.Name, campus.City)
				}
				campusOptions = append(campusOptions, label)
			}
			if len(campusOptions) > 10 {
				return fmt.Errorf("campus %q not found; some available campuses: %s",
					campusName, strings.Join(campusOptions[:10], ", "))
			}
			return fmt.Errorf("campus %q not found. Available campuses: %s",
				campusName, strings.Join(campusOptions, ", "))
		}
	}

	// Resolve project slug → project ID + find campus session
	if GetVerbose() {
		fmt.Printf("Looking up project: %s\n", projectSlug)
	}
	project, err := client.GetProjectBySlug(ctx, projectSlug)
	if err != nil {
		return fmt.Errorf("failed to find project %q: %w", projectSlug, err)
	}

	// Get full project detail to find the campus-specific session ID
	projectDetail, err := client.GetProject(ctx, project.ID)
	if err != nil {
		return fmt.Errorf("failed to get project detail: %w", err)
	}

	// Find the session for our campus
	var sessionID int
	for _, ps := range projectDetail.ProjectSessions {
		if ps.CampusID == campusID && ps.CursusID == cursusID {
			sessionID = ps.ID
			break
		}
	}
	if sessionID == 0 {
		return fmt.Errorf("no project session found for %q at campus %d (cursus %d)", projectSlug, campusID, cursusID)
	}

	// Get full session detail including inscription rules
	// This requires a client_credentials token (project_sessions are not accessible with user tokens)
	if GetVerbose() {
		fmt.Printf("Getting session detail for session %d (using app credentials)\n", sessionID)
	}

	secrets, err := config.LoadDevelopmentSecrets()
	if err != nil {
		// Try config dir secrets as fallback
		secrets, err = config.LoadSecretsFromConfigDir()
		if err != nil {
			return fmt.Errorf("failed to load app credentials (needed for session rules): %w", err)
		}
	}

	appToken, err := api.GetClientCredentialsToken(ctx, secrets.ClientID, secrets.ClientSecret)
	if err != nil {
		return fmt.Errorf("failed to get app token: %w", err)
	}

	appClient := api.NewClient(appToken)
	session, err := appClient.GetProjectSessionDetail(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session detail: %w", err)
	}

	reqs := parseInscriptionRules(session.ProjectSessionsRules)

	if GetVerbose() {
		fmt.Printf("Inscription requirements:\n")
		fmt.Printf("  Required quests: %v\n", reqs.requiredQuests)
		fmt.Printf("  Forbidden quests: %v\n", reqs.forbiddenQuests)
		fmt.Printf("  Forbidden projects: %v\n", reqs.forbiddenProjects)
	}

	// Fetch cursus users with level range (server-side filtering)
	now := time.Now()
	var eligible []eligibleUser
	currentPage := 1
	totalChecked := 0
	totalAPIPages := 0

	for len(eligible) < limit {
		cursusOpts := &api.ListCursusUsersOptions{
			Page:     currentPage,
			PerPage:  100,
			CampusID: campusID,
			Sort:     "-level",
			MinLevel: minLevel,
			MaxLevel: maxLevel,
		}

		cursusUsers, meta, fetchErr := client.ListCursusUsers(ctx, cursusID, cursusOpts)
		if fetchErr != nil {
			return fmt.Errorf("failed to list cursus users: %w", fetchErr)
		}
		totalAPIPages++

		if GetVerbose() && currentPage == 1 && meta != nil {
			fmt.Printf("Total candidates in level range: %d\n", meta.TotalCount)
		}

		for _, cu := range cursusUsers {
			totalChecked++

			// Skip blackholed users (BH date in the past)
			if cu.BlackholedAt != nil && cu.BlackholedAt.Before(now) {
				continue
			}

			// Skip users whose cursus has ended (graduated/exited)
			if cu.EndAt != nil {
				continue
			}

			if GetVerbose() {
				fmt.Printf("  Checking %s (level %.2f)...\n", cu.User.Login, cu.Level)
			}

			// Get full user profile for projects_users
			fullUser, userErr := client.GetUser(ctx, cu.User.ID)
			if userErr != nil {
				if GetVerbose() {
					fmt.Printf("    Skip: failed to get user: %v\n", userErr)
				}
				continue
			}

			// Check forbidden projects (e.g., project not already ongoing/validated)
			if !checkForbiddenProjects(fullUser.ProjectsUsers, reqs.forbiddenProjects) {
				if GetVerbose() {
					fmt.Printf("    Skip: forbidden project active/validated\n")
				}
				continue
			}

			// Check quest requirements
			questUsers, questErr := client.ListUserQuestUsers(ctx, cu.User.ID)
			if questErr != nil {
				if GetVerbose() {
					fmt.Printf("    Skip: failed to get quests: %v\n", questErr)
				}
				continue
			}

			if !checkRequiredQuests(questUsers, reqs.requiredQuests) {
				if GetVerbose() {
					fmt.Printf("    Skip: required quest not validated\n")
				}
				continue
			}

			if !checkForbiddenQuests(questUsers, reqs.forbiddenQuests) {
				if GetVerbose() {
					fmt.Printf("    Skip: forbidden quest validated\n")
				}
				continue
			}

			// Embed campus and cursus info into the full user
			if resolvedCampus != nil && len(fullUser.Campus) == 0 {
				fullUser.Campus = []api.Campus{*resolvedCampus}
			}
			fullUser.CursusUsers = []api.CursusUser{{
				ID:           cu.ID,
				BeginAt:      cu.BeginAt,
				EndAt:        cu.EndAt,
				Grade:        cu.Grade,
				Level:        cu.Level,
				Skills:       cu.Skills,
				BlackholedAt: cu.BlackholedAt,
				Cursus:       cu.Cursus,
				HasCoalition: cu.HasCoalition,
			}}

			// Build quest info for display
			var qInfo []questInfo
			for _, qu := range questUsers {
				if qu.ValidatedAt != nil {
					qInfo = append(qInfo, questInfo{
						Slug:        qu.Quest.Slug,
						ValidatedAt: qu.ValidatedAt.Format("2006-01-02"),
					})
				}
			}

			bhDays := 0
			if cu.BlackholedAt != nil {
				bhDays = int(time.Until(*cu.BlackholedAt).Hours() / 24)
			}

			eligible = append(eligible, eligibleUser{
				User:       *fullUser,
				Level:      cu.Level,
				BlackholeD: bhDays,
				QuestsInfo: qInfo,
			})

			if GetVerbose() {
				fmt.Printf("    ELIGIBLE (%d/%d)\n", len(eligible), limit)
			}

			if len(eligible) >= limit {
				break
			}
		}

		// Stop if no more pages
		if len(cursusUsers) < 100 || (meta != nil && currentPage >= meta.TotalPages) {
			break
		}
		currentPage++
	}

	// Output
	if GetJSONOutput() {
		output := map[string]interface{}{
			"eligible_users": eligible,
			"criteria": map[string]interface{}{
				"project":           projectSlug,
				"campus_id":         campusID,
				"cursus_id":         cursusID,
				"min_level":         minLevel,
				"max_level":         maxLevel,
				"required_quests":   reqs.requiredQuests,
				"forbidden_quests":  reqs.forbiddenQuests,
				"forbidden_projects": reqs.forbiddenProjects,
			},
			"stats": map[string]interface{}{
				"eligible_found":  len(eligible),
				"total_checked":   totalChecked,
				"api_pages_used":  totalAPIPages,
				"limit":           limit,
			},
		}
		jsonData, jsonErr := json.MarshalIndent(output, "", "  ")
		if jsonErr != nil {
			return fmt.Errorf("failed to marshal JSON output: %w", jsonErr)
		}
		fmt.Println(string(jsonData))
	} else {
		printEligibleTable(eligible, project.Name, resolvedCampus, cursusID, reqs, totalChecked, limit)
	}

	return nil
}

func printEligibleTable(users []eligibleUser, projectName string, campus *api.Campus, cursusID int, reqs inscriptionRequirements, totalChecked int, limit int) {
	campusName := "Unknown"
	if campus != nil {
		campusName = campus.Name
	}

	reqCount := len(reqs.requiredQuests) + len(reqs.forbiddenQuests) + len(reqs.forbiddenProjects)
	fmt.Printf("ELIGIBLE USERS FOR: %s (%s, cursus %d)\n", projectName, campusName, cursusID)
	fmt.Printf("Not blackholed | %d inscription rules checked\n\n", reqCount)

	if len(users) == 0 {
		fmt.Println("No eligible users found.")
		fmt.Printf("\nChecked %d users\n", totalChecked)
		return
	}

	fmt.Printf("%-20s %-30s %-10s %s\n",
		"LOGIN", "NAME", "LEVEL", "BH")
	fmt.Printf("%s\n", strings.Repeat("-", 75))

	for _, eu := range users {
		login := truncateString(eu.User.Login, 18)
		displayName := truncateString(eu.User.DisplayName, 28)
		level := fmt.Sprintf("%.2f", eu.Level)

		bh := "-"
		if eu.BlackholeD > 0 {
			bh = fmt.Sprintf("%dd", eu.BlackholeD)
		}

		fmt.Printf("%-20s %-30s %-10s %s\n", login, displayName, level, bh)
	}

	fmt.Printf("\nShowing %d eligible users (checked %d candidates)\n", len(users), totalChecked)
	if len(users) >= limit {
		fmt.Printf("Use --limit %d to see more results\n", limit*2)
	}
}
