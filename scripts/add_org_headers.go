package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	notesDir := "/Users/andy_dawes/OneDrive/Repos/notes/org"

	// Read all .org files
	files, err := os.ReadDir(notesDir)
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		os.Exit(1)
	}

	// Process each .org file
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".org") {
			continue
		}

		filePath := filepath.Join(notesDir, file.Name())
		if err := processOrgFile(filePath); err != nil {
			fmt.Printf("Error processing %s: %v\n", file.Name(), err)
		}
	}

	fmt.Println("Finished processing all .org files")
}

func processOrgFile(filePath string) error {
	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	contentStr := string(content)

	// Check if headers already exist
	hasTitle := regexp.MustCompile(`(?m)^#\+TITLE:`).MatchString(contentStr)
	hasAuthor := regexp.MustCompile(`(?m)^#\+AUTHOR:`).MatchString(contentStr)
	hasStartup := regexp.MustCompile(`(?m)^#\+STARTUP:`).MatchString(contentStr)

	// If all headers exist, skip this file
	if hasTitle && hasAuthor && hasStartup {
		fmt.Printf("Skipping %s (headers already present)\n", filepath.Base(filePath))
		return nil
	}

	// Generate title from filename
	title := generateTitleFromFilename(filepath.Base(filePath))

	// Prepare new headers
	var newHeaders strings.Builder
	if !hasTitle {
		newHeaders.WriteString(fmt.Sprintf("#+TITLE: %s\n", title))
	}
	if !hasAuthor {
		newHeaders.WriteString("#+AUTHOR: Andy Dawes\n")
	}
	if !hasStartup {
		newHeaders.WriteString("#+STARTUP: showall\n")
	}

	// If we need to add headers, insert them at the beginning
	if newHeaders.Len() > 0 {
		// Remove any existing incomplete headers
		lines := strings.Split(contentStr, "\n")
		var filteredLines []string

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			// Skip incomplete or empty headers
			if trimmed == "#+TITLE" || trimmed == "#+AUTHOR" || trimmed == "#+STARTUP" {
				continue
			}
			filteredLines = append(filteredLines, line)
		}

		// Reconstruct content with new headers
		newContent := newHeaders.String() + "\n" + strings.Join(filteredLines, "\n")

		// Write back to file
		if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}

		fmt.Printf("Updated %s with headers\n", filepath.Base(filePath))
	}

	return nil
}

func generateTitleFromFilename(filename string) string {
	// Remove .org extension
	name := strings.TrimSuffix(filename, ".org")

	// Replace underscores and hyphens with spaces
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "-", " ")

	// Handle special cases
	switch name {
	case "26 in 52":
		return "26 in 52 Reading Challenge"
	case "365 bible":
		return "365 Bible Reading Plan"
	case "fair fall the bones":
		return "Fair Fall the Bones"
	case "ww2":
		return "World War II Notes"
	case "tasks org archive":
		return "Tasks Archive"
	case "habits org archive":
		return "Habits Archive"
	case "bible org":
		return "Bible Study Notes"
	case "affirm org":
		return "Affirmations"
	case "birthdays":
		return "Birthdays and Anniversaries"
	case "addresses":
		return "Addresses and Contact Information"
	case "holidays":
		return "Holidays and Travel"
	case "habits":
		return "Habits and Routines"
	case "tasks":
		return "Tasks and Projects"
	case "library":
		return "Library and Books"
	case "music":
		return "Music Collection"
	case "someday":
		return "Someday Maybe"
	case "vacations":
		return "Vacations and Travel"
	case "walking":
		return "Walking and Exercise"
	case "wargame":
		return "Wargaming Notes"
	case "whisky":
		return "Whisky Collection"
	case "ideas":
		return "Ideas and Inspiration"
	case "journal":
		return "Journal Entries"
	case "archive":
		return "Archive"
	case "refile":
		return "Refile"
	case "storeroom":
		return "Storeroom"
	default:
		// Title case the name
		return strings.Title(name)
	}
}
