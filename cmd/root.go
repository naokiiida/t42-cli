/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"os"
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "t42-cli",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.t42-cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// Add a hidden command for huh demo
	huhDemoCmd := &cobra.Command{
		Use:    "huh-demo",
		Short:  "Run a demo of huh interactive prompt (for dev only)",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			runHuhDemo()
		},
	}
	rootCmd.AddCommand(huhDemoCmd)
}

// Add a stub for huh integration
type demoFormData struct {
	Name string
}

func runHuhDemo() {
	var data demoFormData
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("What's your name?").Value(&data.Name),
		),
	)
	err := form.Run()
	if err != nil {
		fmt.Println("Prompt cancelled or error:", err)
		return
	}
	fmt.Println("Hello,", data.Name)
}


