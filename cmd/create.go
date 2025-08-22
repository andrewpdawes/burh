package cmd

import (
	"fmt"
	"os"
	"strings"

	"burh/notes"

	"github.com/spf13/cobra"
)

var (
	title   string
	content string
	tags    string
	format  string
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new note",
	Long: `Create a new note with the specified title, content, tags, and format.
The note will be saved with a unique ID based on timestamp and title.`,
	Run: runCreate,
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Local flags
	createCmd.Flags().StringVarP(&title, "title", "t", "", "Note title (required)")
	createCmd.Flags().StringVarP(&content, "content", "c", "", "Note content")
	createCmd.Flags().StringVarP(&tags, "tags", "g", "", "Comma-separated tags")
	createCmd.Flags().StringVarP(&format, "format", "f", "txt", "Note format (txt or org)")

	createCmd.MarkFlagRequired("title")
}

func runCreate(cmd *cobra.Command, args []string) {
	// Get config
	cfg := getConfig()

	// Validate format
	if format != "txt" && format != "org" {
		fmt.Println("Error: format must be 'txt' or 'org'")
		os.Exit(1)
	}

	// Parse tags
	var tagList []string
	if tags != "" {
		tagList = strings.Split(tags, ",")
		for i, tag := range tagList {
			tagList[i] = strings.TrimSpace(tag)
		}
	}

	// Create note manager with all directories
	noteManager := notes.NewManagerWithDirs(cfg.NotesDirs)

	// Create note
	note, err := noteManager.CreateNote(title, content, tagList, format)
	if err != nil {
		fmt.Printf("Error creating note: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Note created successfully!\n")
	fmt.Printf("ID: %s\n", note.ID)
	fmt.Printf("Title: %s\n", note.Title)
	fmt.Printf("Format: %s\n", note.Format)
	fmt.Printf("Filename: %s\n", note.Filename)
	if len(note.Tags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(note.Tags, ", "))
	}
}
