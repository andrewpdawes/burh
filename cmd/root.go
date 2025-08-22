package cmd

import (
	"fmt"
	"os"

	"burh/config"
	"burh/notes"
	"burh/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "burh",
	Short: "A simple note-taking tool with TUI and CLI interfaces",
	Long: `Burh is a note-taking tool inspired by Denote, providing both CLI and TUI interfaces.
It supports creating, editing, searching, and managing notes in both .org and .txt formats.
Each note gets a unique ID based on timestamp and title.`,
	Run: runTUI,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.burhrc.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&showContent, "content", "c", false, "Show note content in list/search results")

	// Add subcommands
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(listDirsCmd)
	rootCmd.AddCommand(addDirCmd)
	rootCmd.AddCommand(removeDirCmd)

	// Initialize config after flags are parsed
	cobra.OnInitialize(initConfig)
}

// Global config variable
var globalConfig *config.Config

// getConfig ensures the config is loaded and returns it
func getConfig() *config.Config {
	if globalConfig == nil {
		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Store config globally
		globalConfig = cfg
	}
	return globalConfig
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Just ensure config is loaded
	getConfig()
}

// runTUI starts the TUI interface
func runTUI(cmd *cobra.Command, args []string) {
	// Get config
	cfg := getConfig()

	// Create note manager with all directories
	noteManager := notes.NewManagerWithDirs(cfg.NotesDirs)

	// Create TUI model
	model := tui.NewModel(noteManager, cfg)

	// Run TUI
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
