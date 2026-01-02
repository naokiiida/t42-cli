package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/naokiiida/t42-cli/internal/api"
)

var projectCmd = &cobra.Command{
	Use:     "project",
	Aliases: []string{"pj"},
	Short:   "Project management commands",
	Long: `Manage your 42 projects.

This command group allows you to list projects, view project details,
and clone project repositories to your local machine.`,
}

var listProjectsCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects",
	Long: `List projects from the 42 API.

You can filter projects by cursus and control pagination options.
Use --mine to show only your projects.`,
	RunE: runListProjects,
}

var showProjectCmd = &cobra.Command{
	Use:   "show <project-slug>",
	Short: "Show project details",
	Long: `Show detailed information about a specific project.

You can specify a project by its slug (e.g., 'libft', 'get_next_line').`,
	Args: cobra.ExactArgs(1),
	RunE: runShowProject,
}

var cloneProjectCmd = &cobra.Command{
	Use:   "clone <project-slug> [directory]",
	Short: "Clone a project repository",
	Long: `Clone a project's Git repository to your local machine.

If no directory is specified, the project will be cloned into a
directory named after the project slug.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runCloneProject,
}

var cloneMineCmd = &cobra.Command{
	Use:   "clone-mine <project-slug> [directory]",
	Short: "Clone your project repository",
	Long: `Clone your own project repository to your local machine.

This command finds your team's repository for the specified project
and clones it using the repo_url from your team data. If you have
multiple teams for the same project, it will use the most recent one.

If no directory is specified, the project will be cloned into a
directory named after the project slug with your login as suffix.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runCloneMine,
}

func init() {
	// Add project subcommands
	projectCmd.AddCommand(listProjectsCmd)
	projectCmd.AddCommand(showProjectCmd)
	projectCmd.AddCommand(cloneProjectCmd)
	projectCmd.AddCommand(cloneMineCmd)
	
	// Add project command to root
	rootCmd.AddCommand(projectCmd)
	
	// List command flags
	listProjectsCmd.Flags().Bool("mine", false, "Show only my projects")
	listProjectsCmd.Flags().IntP("page", "p", 1, "Page number")
	listProjectsCmd.Flags().Int("per-page", 20, "Number of projects per page")
	listProjectsCmd.Flags().Int("cursus", 0, "Filter by cursus ID")
	listProjectsCmd.Flags().StringP("sort", "s", "", "Sort by field (name, id, created_at)")
	
	// Clone command flags
	cloneProjectCmd.Flags().Bool("no-clone", false, "Show clone command without executing")
	cloneProjectCmd.Flags().Bool("force", false, "Force clone even if directory exists")
	
	// Clone mine command flags
	cloneMineCmd.Flags().Bool("no-clone", false, "Show clone command without executing")
	cloneMineCmd.Flags().Bool("force", false, "Force clone even if directory exists")
	cloneMineCmd.Flags().Bool("latest", true, "Use the latest team (default: true)")
}

func runListProjects(cmd *cobra.Command, args []string) error {
	// Create API client with automatic token refresh
	client, err := NewAPIClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	
	// Get flags
	mine, _ := cmd.Flags().GetBool("mine")
	page, _ := cmd.Flags().GetInt("page")
	perPage, _ := cmd.Flags().GetInt("per-page")
	cursusID, _ := cmd.Flags().GetInt("cursus")
	sort, _ := cmd.Flags().GetString("sort")
	
	if mine {
		// List user's projects
		user, err := client.GetMe(ctx)
		if err != nil {
			return fmt.Errorf("failed to get user info: %w", err)
		}
		
		opts := &api.ListUserProjectsOptions{
			Page:    page,
			PerPage: perPage,
			Sort:    sort,
		}
		
		projectUsers, meta, err := client.ListUserProjects(ctx, user.ID, opts)
		if err != nil {
			return fmt.Errorf("failed to list user projects: %w", err)
		}
		
		if GetJSONOutput() {
			output := map[string]interface{}{
				"projects": projectUsers,
				"meta":     meta,
			}
			jsonData, _ := json.MarshalIndent(output, "", "  ")
			fmt.Println(string(jsonData))
		} else {
			printUserProjectsTable(projectUsers, meta)
		}
	} else {
		// List all projects
		opts := &api.ListProjectsOptions{
			Page:     page,
			PerPage:  perPage,
			CursusID: cursusID,
			Sort:     sort,
		}
		
		projects, meta, err := client.ListProjects(ctx, opts)
		if err != nil {
			return fmt.Errorf("failed to list projects: %w", err)
		}
		
		if GetJSONOutput() {
			output := map[string]interface{}{
				"projects": projects,
				"meta":     meta,
			}
			jsonData, _ := json.MarshalIndent(output, "", "  ")
			fmt.Println(string(jsonData))
		} else {
			printProjectsTable(projects, meta)
		}
	}
	
	return nil
}

func runShowProject(cmd *cobra.Command, args []string) error {
	projectSlug := args[0]

	// Create API client with automatic token refresh
	client, err := NewAPIClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	
	// Get project by slug
	project, err := client.GetProjectBySlug(ctx, projectSlug)
	if err != nil {
		return fmt.Errorf("failed to get project '%s': %w", projectSlug, err)
	}
	
	if GetJSONOutput() {
		jsonData, _ := json.MarshalIndent(project, "", "  ")
		fmt.Println(string(jsonData))
	} else {
		printProjectDetails(project)
	}
	
	return nil
}

func runCloneProject(cmd *cobra.Command, args []string) error {
	projectSlug := args[0]
	var targetDir string

	if len(args) > 1 {
		targetDir = args[1]
	} else {
		targetDir = projectSlug
	}

	// Create API client with automatic token refresh
	client, err := NewAPIClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	
	// Get project details
	project, err := client.GetProjectBySlug(ctx, projectSlug)
	if err != nil {
		return fmt.Errorf("failed to get project '%s': %w", projectSlug, err)
	}
	
	// Check if project has a Git URL
	if project.GitURL == "" {
		return fmt.Errorf("project '%s' does not have a Git repository", projectSlug)
	}
	
	// Get flags
	noClone, _ := cmd.Flags().GetBool("no-clone")
	force, _ := cmd.Flags().GetBool("force")
	
	// Check if directory exists
	if _, err := os.Stat(targetDir); err == nil && !force {
		if GetJSONOutput() {
			fmt.Printf(`{"error":"Directory '%s' already exists. Use --force to override."}%s`, targetDir, "\n")
			return nil
		} else {
			var overwrite bool
			err := huh.NewConfirm().
				Title(fmt.Sprintf("Directory '%s' already exists", targetDir)).
				Description("Do you want to remove it and clone fresh?").
				Value(&overwrite).
				Run()
			
			if err != nil {
				return fmt.Errorf("failed to get user confirmation: %w", err)
			}
			
			if !overwrite {
				fmt.Println("Clone cancelled.")
				return nil
			}
			
			// Remove existing directory
			if err := os.RemoveAll(targetDir); err != nil {
				return fmt.Errorf("failed to remove existing directory: %w", err)
			}
		}
	}
	
	// Prepare git clone command
	gitCmd := []string{"git", "clone", project.GitURL, targetDir}
	
	if noClone || GetJSONOutput() {
		result := map[string]interface{}{
			"project":    project.Name,
			"slug":       project.Slug,
			"git_url":    project.GitURL,
			"directory":  targetDir,
			"command":    strings.Join(gitCmd, " "),
		}
		
		if noClone {
			result["executed"] = false
		}
		
		jsonData, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(jsonData))
		
		if noClone {
			return nil
		}
	} else {
		fmt.Printf("📦 Cloning project: %s\n", project.Name)
		fmt.Printf("🔗 Repository: %s\n", project.GitURL)
		fmt.Printf("📁 Target directory: %s\n", targetDir)
		fmt.Printf("⚡ Running: %s\n\n", strings.Join(gitCmd, " "))
	}
	
	// Execute git clone
	cmd_exec := exec.Command("git", "clone", project.GitURL, targetDir)
	cmd_exec.Stdout = os.Stdout
	cmd_exec.Stderr = os.Stderr
	
	if err := cmd_exec.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	
	if !GetJSONOutput() {
		fmt.Printf("\n✅ Successfully cloned %s to %s!\n", project.Name, targetDir)
		
		// Show next steps
		fmt.Printf("\n📝 Next steps:\n")
		fmt.Printf("   cd %s\n", targetDir)
		fmt.Printf("   # Start working on your project!\n")
	}
	
	return nil
}

func printProjectsTable(projects []api.Project, meta *api.PaginationMeta) {
	if len(projects) == 0 {
		fmt.Println("No projects found.")
		return
	}
	
	// Header
	fmt.Printf("%-40s %-20s %-10s %s\n", "NAME", "SLUG", "TIER", "DESCRIPTION")
	fmt.Printf("%s\n", strings.Repeat("-", 100))
	
	// Projects
	for _, project := range projects {
		name := truncateString(project.Name, 38)
		slug := truncateString(project.Slug, 18)
		description := truncateString(project.Description, 30)
		
		fmt.Printf("%-40s %-20s %-10d %s\n", name, slug, project.Tier, description)
	}
	
	// Pagination info
	if meta != nil {
		fmt.Printf("\n📄 Page %d of %d (%d total projects)\n", meta.Page, meta.TotalPages, meta.TotalCount)
		if meta.Page < meta.TotalPages {
			fmt.Printf("   Use --page %d to see the next page\n", meta.Page+1)
		}
	}
}

func printUserProjectsTable(projectUsers []api.ProjectUser, meta *api.PaginationMeta) {
	if len(projectUsers) == 0 {
		fmt.Println("No projects found.")
		return
	}
	
	// Header
	fmt.Printf("%-30s %-15s %-10s %-15s %s\n", "PROJECT", "STATUS", "MARK", "VALIDATED", "MARKED AT")
	fmt.Printf("%s\n", strings.Repeat("-", 100))
	
	// Projects
	for _, pu := range projectUsers {
		name := truncateString(pu.Project.Name, 28)
		status := truncateString(pu.Status, 13)
		
		mark := "N/A"
		if pu.FinalMark != nil {
			mark = strconv.Itoa(*pu.FinalMark)
		}
		
		validated := "N/A"
		if pu.Validated != nil {
			if *pu.Validated {
				validated = "✅ Yes"
			} else {
				validated = "❌ No"
			}
		}
		
		markedAt := "N/A"
		if pu.MarkedAt != nil {
			markedAt = pu.MarkedAt.Format("2006-01-02")
		}
		
		fmt.Printf("%-30s %-15s %-10s %-15s %s\n", name, status, mark, validated, markedAt)
	}
	
	// Pagination info
	if meta != nil {
		fmt.Printf("\n📄 Page %d (%d projects shown)\n", meta.Page, len(projectUsers))
	}
}

func printProjectDetails(project *api.Project) {
	fmt.Printf("📦 Project: %s\n", project.Name)
	fmt.Printf("🏷️  Slug: %s\n", project.Slug)
	fmt.Printf("⭐ Tier: %d\n", project.Tier)
	
	if project.GitURL != "" {
		fmt.Printf("🔗 Repository: %s\n", project.GitURL)
	}
	
	if project.Description != "" {
		fmt.Printf("\n📄 Description:\n%s\n", wrapText(project.Description, 80))
	}
	
	if len(project.Objectives) > 0 {
		fmt.Printf("\n🎯 Objectives:\n")
		for i, objective := range project.Objectives {
			fmt.Printf("   %d. %s\n", i+1, objective)
		}
	}
	
	if len(project.Cursus) > 0 {
		fmt.Printf("\n📚 Cursus:\n")
		for _, cursus := range project.Cursus {
			fmt.Printf("   • %s (%s)\n", cursus.Name, cursus.Slug)
		}
	}
	
	if project.Parent != nil {
		fmt.Printf("\n⬆️  Parent Project: %s\n", project.Parent.Name)
	}
	
	if len(project.Children) > 0 {
		fmt.Printf("\n⬇️  Child Projects:\n")
		for _, child := range project.Children {
			fmt.Printf("   • %s\n", child.Name)
		}
	}
	
	fmt.Printf("\n📅 Created: %s\n", project.CreatedAt.Format(time.RFC3339))
	fmt.Printf("🔄 Updated: %s\n", project.UpdatedAt.Format(time.RFC3339))
	
	if project.GitURL != "" {
		fmt.Printf("\n💡 To clone this project:\n")
		fmt.Printf("   t42 project clone %s\n", project.Slug)
	}
}

func runCloneMine(cmd *cobra.Command, args []string) error {
	projectSlug := args[0]

	// Create API client with automatic token refresh
	client, err := NewAPIClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	
	// Get current user
	user, err := client.GetMe(ctx)
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}
	
	// Find the project in user's projects
	userProjects, _, err := client.ListUserProjects(ctx, user.ID, &api.ListUserProjectsOptions{
		PerPage: 100, // Get enough to find the project
	})
	if err != nil {
		return fmt.Errorf("failed to get user projects: %w", err)
	}
	
	var targetProjectUser *api.ProjectUser
	for _, pu := range userProjects {
		if pu.Project.Slug == projectSlug {
			targetProjectUser = &pu
			break
		}
	}
	
	if targetProjectUser == nil {
		return fmt.Errorf("project '%s' not found in your projects", projectSlug)
	}
	
	// Get full project user details to access teams
	fullProjectUser, err := client.GetProjectUser(ctx, targetProjectUser.ID)
	if err != nil {
		return fmt.Errorf("failed to get project user details: %w", err)
	}
	
	// Find the team with repo_url
	var repoURL string
	var teamName string
	
	// Use latest team by default, or find the first one with a repo_url
	latest, _ := cmd.Flags().GetBool("latest")
	
	if latest && len(fullProjectUser.Teams) > 0 {
		// Use the most recent team (teams are usually ordered by creation date)
		team := fullProjectUser.Teams[len(fullProjectUser.Teams)-1]
		if team.RepoURL != "" {
			repoURL = team.RepoURL
			teamName = team.Name
		}
	}
	
	// If no repo URL found from latest, try all teams
	if repoURL == "" {
		for _, team := range fullProjectUser.Teams {
			if team.RepoURL != "" {
				repoURL = team.RepoURL
				teamName = team.Name
				break
			}
		}
	}
	
	if repoURL == "" {
		return fmt.Errorf("no repository URL found for project '%s' in your teams", projectSlug)
	}
	
	// Determine target directory
	var targetDir string
	if len(args) > 1 {
		targetDir = args[1]
	} else {
		targetDir = fmt.Sprintf("%s-%s", projectSlug, user.Login)
	}
	
	// Get flags
	noClone, _ := cmd.Flags().GetBool("no-clone")
	force, _ := cmd.Flags().GetBool("force")
	
	// Check if directory exists
	if _, err := os.Stat(targetDir); err == nil && !force {
		if GetJSONOutput() {
			fmt.Printf(`{"error":"Directory '%s' already exists. Use --force to override."}%s`, targetDir, "\n")
			return nil
		} else {
			var overwrite bool
			err := huh.NewConfirm().
				Title(fmt.Sprintf("Directory '%s' already exists", targetDir)).
				Description("Do you want to remove it and clone fresh?").
				Value(&overwrite).
				Run()
			
			if err != nil {
				return fmt.Errorf("failed to get user confirmation: %w", err)
			}
			
			if !overwrite {
				fmt.Println("Clone cancelled.")
				return nil
			}
			
			// Remove existing directory
			if err := os.RemoveAll(targetDir); err != nil {
				return fmt.Errorf("failed to remove existing directory: %w", err)
			}
		}
	}
	
	// Prepare git clone command
	gitCmd := []string{"git", "clone", repoURL, targetDir}
	
	if noClone || GetJSONOutput() {
		result := map[string]interface{}{
			"project":     fullProjectUser.Project.Name,
			"slug":        fullProjectUser.Project.Slug,
			"team_name":   teamName,
			"repo_url":    repoURL,
			"directory":   targetDir,
			"command":     strings.Join(gitCmd, " "),
			"status":      fullProjectUser.Status,
		}
		
		if fullProjectUser.FinalMark != nil {
			result["final_mark"] = *fullProjectUser.FinalMark
		}
		if fullProjectUser.Validated != nil {
			result["validated"] = *fullProjectUser.Validated
		}
		
		if noClone {
			result["executed"] = false
		}
		
		jsonData, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(jsonData))
		
		if noClone {
			return nil
		}
	} else {
		fmt.Printf("📦 Cloning your project: %s\n", fullProjectUser.Project.Name)
		fmt.Printf("👤 Team: %s\n", teamName)
		fmt.Printf("📊 Status: %s\n", fullProjectUser.Status)
		if fullProjectUser.FinalMark != nil {
			fmt.Printf("🎯 Final Mark: %d\n", *fullProjectUser.FinalMark)
		}
		if fullProjectUser.Validated != nil {
			if *fullProjectUser.Validated {
				fmt.Printf("✅ Validated: Yes\n")
			} else {
				fmt.Printf("❌ Validated: No\n")
			}
		}
		fmt.Printf("🔗 Repository: %s\n", repoURL)
		fmt.Printf("📁 Target directory: %s\n", targetDir)
		fmt.Printf("⚡ Running: %s\n\n", strings.Join(gitCmd, " "))
	}
	
	// Execute git clone
	cmd_exec := exec.Command("git", "clone", repoURL, targetDir)
	cmd_exec.Stdout = os.Stdout
	cmd_exec.Stderr = os.Stderr
	
	if err := cmd_exec.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	
	if !GetJSONOutput() {
		fmt.Printf("\n✅ Successfully cloned your %s repository to %s!\n", fullProjectUser.Project.Name, targetDir)
		
		// Show next steps
		fmt.Printf("\n📝 Next steps:\n")
		fmt.Printf("   cd %s\n", targetDir)
		fmt.Printf("   # Continue working on your project!\n")
	}
	
	return nil
}

func wrapText(text string, width int) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}
	
	var lines []string
	var currentLine []string
	currentLength := 0
	
	for _, word := range words {
		if currentLength+len(word)+len(currentLine) > width && len(currentLine) > 0 {
			lines = append(lines, strings.Join(currentLine, " "))
			currentLine = []string{word}
			currentLength = len(word)
		} else {
			currentLine = append(currentLine, word)
			currentLength += len(word)
		}
	}
	
	if len(currentLine) > 0 {
		lines = append(lines, strings.Join(currentLine, " "))
	}
	
	return strings.Join(lines, "\n")
}