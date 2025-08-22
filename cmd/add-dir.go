package cmd

import (
	"fmt"
	"os"

	"burh/config"

	"github.com/spf13/cobra"
)

var addDirPath string

// addDirCmd represents the add-dir command
var addDirCmd = &cobra.Command{
	Use:   "add-dir",
	Short: "Add a new notes directory",
	Long: `Add a new directory to the list of directories where Burh will look for notes.
The directory will be created if it doesn't exist.`,
	Run: runAddDir,
}

func init() {
	addDirCmd.Flags().StringVarP(&addDirPath, "path", "p", "", "Path to the notes directory (required)")
	addDirCmd.MarkFlagRequired("path")
}

func runAddDir(cmd *cobra.Command, args []string) {
	if err := config.AddNotesDirectory(addDirPath); err != nil {
		fmt.Printf("Error adding directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully added notes directory: %s\n", addDirPath)

	// Reload config to get updated list
	globalConfig = nil // Force reload
	cfg := getConfig()
	fmt.Printf("\nCurrent notes directories:\n")
	for i, dir := range cfg.NotesDirs {
		fmt.Printf("  %d. %s\n", i+1, dir)
	}
}
