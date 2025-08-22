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
	showContent bool
	showTags    bool
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all notes",
	Long: `List all notes in the notes directory.
You can optionally show content and tags for each note.`,
	Run: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Local flags
	listCmd.Flags().BoolVarP(&showContent, "content", "c", false, "Show note content")
	listCmd.Flags().BoolVarP(&showTags, "tags", "t", false, "Show note tags")
}

func runList(cmd *cobra.Command, args []string) {
	// Get config
	cfg := getConfig()

	// Create note manager with all directories
	noteManager := notes.NewManagerWithDirs(cfg.NotesDirs)

	// List notes
	notes, err := noteManager.ListNotes()
	if err != nil {
		fmt.Printf("Error listing notes: %v\n", err)
		os.Exit(1)
	}

	if len(notes) == 0 {
		fmt.Println("No notes found.")
		return
	}

	heading := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render(fmt.Sprintf("Found %d notes", len(notes)))
	fmt.Printf("%s\n\n", heading)

	for i, note := range notes {
		ts := lipgloss.NewStyle().Foreground(lipgloss.Color("#7C8DA6")).Render(note.Created.Format("2006-01-02 15:04"))
		fmtTag := lipgloss.NewStyle().Foreground(lipgloss.Color("#81A1C1")).Render("[" + note.Format + "]")
		title := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Render(note.Title)
		fmt.Printf("%2d. %s  %s  %s\n", i+1, ts, fmtTag, title)

		if showTags && len(note.Tags) > 0 {
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

		if showContent && note.Content != "" {
			// Truncate content if too long
			content := note.Content
			if len(content) > 100 {
				content = content[:100] + "..."
			}
			fmt.Printf("    %s %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("#7C8DA6")).Render("Content:"), content)
		}

		fmt.Printf("    %s %s\n\n", lipgloss.NewStyle().Foreground(lipgloss.Color("#7C8DA6")).Render("ID:"), note.ID)
	}
}
