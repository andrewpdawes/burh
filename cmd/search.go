package cmd

import (
	"fmt"
	"os"
	"strings"

	"burh/notes"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	searchQuery       string
	showContentSearch bool
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search notes by title, content, or tags",
	Long: `Search for notes that match the given query.
The search is case-insensitive and looks in titles, content, and tags.`,
	Args: cobra.ExactArgs(1),
	Run:  runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)

	// Local flags
	searchCmd.Flags().BoolVarP(&showContentSearch, "content", "c", false, "Show note content")
}

func runSearch(cmd *cobra.Command, args []string) {
	searchQuery = args[0]

	// Get config
	cfg := getConfig()

	// Create note manager with all directories
	noteManager := notes.NewManagerWithDirs(cfg.NotesDirs)

	// Search notes
	results, err := noteManager.SearchNotes(searchQuery)
	if err != nil {
		fmt.Printf("Error searching notes: %v\n", err)
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Printf("No notes found matching '%s'\n", searchQuery)
		return
	}

	heading := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render(fmt.Sprintf("Found %d notes matching '%s'", len(results), searchQuery))
	fmt.Printf("%s\n\n", heading)

	for i, note := range results {
		ts := lipgloss.NewStyle().Foreground(lipgloss.Color("#7C8DA6")).Render(note.Created.Format("2006-01-02 15:04"))
		fmtTag := lipgloss.NewStyle().Foreground(lipgloss.Color("#81A1C1")).Render("[" + note.Format + "]")
		title := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Render(note.Title)
		fmt.Printf("%2d. %s  %s  %s\n", i+1, ts, fmtTag, title)

		if len(note.Tags) > 0 {
			// Truncate tags to show only first 6
			tagsToShow := note.Tags
			if len(note.Tags) > 6 {
				tagsToShow = note.Tags[:6]
			}
			tagsStr := strings.Join(tagsToShow, ", ")
			if len(note.Tags) > 6 {
				tagsStr += "..."
			}
			fmt.Printf("    %s %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("#7C8DA6")).Render("Tags:"), tagsStr)
		}

		if showContentSearch && note.Content != "" {
			content := note.Content
			if len(content) > 100 {
				content = content[:100] + "..."
			}
			fmt.Printf("    %s %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("#7C8DA6")).Render("Content:"), content)
		}

		fmt.Printf("    %s %s\n\n", lipgloss.NewStyle().Foreground(lipgloss.Color("#7C8DA6")).Render("ID:"), note.ID)
	}
}
