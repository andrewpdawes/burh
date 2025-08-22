package cmd

import (
	"fmt"
	"os"

	"burh/config"

	"github.com/spf13/cobra"
)

var removeDirPath string

// removeDirCmd represents the remove-dir command
var removeDirCmd = &cobra.Command{
	Use:   "remove-dir",
	Short: "Remove a notes directory",
	Long: `Remove a directory from the list of directories where Burh looks for notes.
At least one directory must remain in the configuration.`,
	Run: runRemoveDir,
}

func init() {
	removeDirCmd.Flags().StringVarP(&removeDirPath, "path", "p", "", "Path to the notes directory to remove (required)")
	removeDirCmd.MarkFlagRequired("path")
}

func runRemoveDir(cmd *cobra.Command, args []string) {
	if err := config.RemoveNotesDirectory(removeDirPath); err != nil {
		fmt.Printf("Error removing directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully removed notes directory: %s\n", removeDirPath)

	// Reload config to get updated list
	globalConfig = nil // Force reload
	cfg := getConfig()
	fmt.Printf("\nCurrent notes directories:\n")
	for i, dir := range cfg.NotesDirs {
		fmt.Printf("  %d. %s\n", i+1, dir)
	}
}
