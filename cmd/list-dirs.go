package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// listDirsCmd represents the list-dirs command
var listDirsCmd = &cobra.Command{
	Use:   "list-dirs",
	Short: "List all notes directories",
	Long: `List all directories that Burh is configured to search for notes.
This shows the current configuration of notes directories.`,
	Run: runListDirs,
}

func init() {
	// No flags needed for this command
}

func runListDirs(cmd *cobra.Command, args []string) {
	cfg := getConfig()

	if len(cfg.NotesDirs) == 0 {
		fmt.Println("No notes directories configured.")
		return
	}

	fmt.Printf("Notes directories (%d total):\n", len(cfg.NotesDirs))
	for i, dir := range cfg.NotesDirs {
		fmt.Printf("  %d. %s\n", i+1, dir)
	}
}
