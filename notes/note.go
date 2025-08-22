package notes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Note represents a single note
type Note struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Content  string    `json:"content"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
	Tags     []string  `json:"tags"`
	Format   string    `json:"format"` // "org", "txt", or "md"
	Filename string    `json:"filename"`
}

// Manager handles note operations
type Manager struct {
	notesDirs []string // Changed from notesDir to notesDirs
}

// NewManager creates a new note manager
func NewManager(notesDir string) *Manager {
	return &Manager{
		notesDirs: []string{notesDir},
	}
}

// NewManagerWithDirs creates a new note manager with multiple directories
func NewManagerWithDirs(notesDirs []string) *Manager {
	return &Manager{
		notesDirs: notesDirs,
	}
}

// GetNotesDir returns the primary notes directory path
func (m *Manager) GetNotesDir() string {
	if len(m.notesDirs) == 0 {
		return ""
	}
	return m.notesDirs[0] // Assuming the first directory is the primary one
}

// GetNotesDirs returns all notes directories
func (m *Manager) GetNotesDirs() []string {
	return m.notesDirs
}

// CreateNote creates a new note with a unique ID
func (m *Manager) CreateNote(title, content string, tags []string, format string) (*Note, error) {
	now := time.Now()

	// Generate unique ID: timestamp + sanitized title
	sanitizedTitle := sanitizeTitle(title)
	id := fmt.Sprintf("%s_%s", now.Format("20060102_150405"), sanitizedTitle)

	// Ensure format is valid
	if format != "org" && format != "txt" && format != "md" {
		format = "txt"
	}

	// Create filename
	filename := fmt.Sprintf("%s.%s", id, format)

	note := &Note{
		ID:       id,
		Title:    title,
		Content:  content,
		Created:  now,
		Modified: now,
		Tags:     tags,
		Format:   format,
		Filename: filename,
	}

	// Ensure notes directory exists
	if err := os.MkdirAll(m.notesDirs[0], 0755); err != nil {
		return nil, fmt.Errorf("failed to create notes directory: %w", err)
	}

	// Save note to file
	if err := m.saveNoteToFile(note); err != nil {
		return nil, fmt.Errorf("failed to save note: %w", err)
	}

	return note, nil
}

// GetNote retrieves a note by ID
func (m *Manager) GetNote(id string) (*Note, error) {
	// Find the note file
	files, err := os.ReadDir(m.notesDirs[0]) // Assuming the first directory is the primary one
	if err != nil {
		return nil, fmt.Errorf("failed to read notes directory: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), id) {
			return m.loadNoteFromFile(filepath.Join(m.notesDirs[0], file.Name()))
		}
	}

	return nil, fmt.Errorf("note not found: %s", id)
}

// UpdateNote updates an existing note
func (m *Manager) UpdateNote(id, title, content string, tags []string) (*Note, error) {
	note, err := m.GetNote(id)
	if err != nil {
		return nil, err
	}

	note.Title = title
	note.Content = content
	note.Tags = tags
	note.Modified = time.Now()

	if err := m.saveNoteToFile(note); err != nil {
		return nil, fmt.Errorf("failed to save updated note: %w", err)
	}

	return note, nil
}

// DeleteNote deletes a note by ID
func (m *Manager) DeleteNote(id string) error {
	note, err := m.GetNote(id)
	if err != nil {
		return err
	}

	filepath := filepath.Join(m.notesDirs[0], note.Filename)
	return os.Remove(filepath)
}

// ListNotes returns all notes
func (m *Manager) ListNotes() ([]*Note, error) {
	var allNotes []*Note
	for _, notesDir := range m.notesDirs {
		files, err := os.ReadDir(notesDir)
		if err != nil {
			return nil, fmt.Errorf("failed to read notes directory %s: %w", notesDir, err)
		}

		for _, file := range files {
			if !file.IsDir() && (strings.HasSuffix(file.Name(), ".org") || strings.HasSuffix(file.Name(), ".txt") || strings.HasSuffix(file.Name(), ".md")) {
				note, err := m.loadNoteFromFile(filepath.Join(notesDir, file.Name()))
				if err != nil {
					continue // Skip files that can't be loaded
				}
				allNotes = append(allNotes, note)
			}
		}
	}

	return allNotes, nil
}

// SearchNotes searches notes by title, content, or tags
func (m *Manager) SearchNotes(query string) ([]*Note, error) {
	notes, err := m.ListNotes()
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var results []*Note

	for _, note := range notes {
		if strings.Contains(strings.ToLower(note.Title), query) ||
			strings.Contains(strings.ToLower(note.Content), query) ||
			containsTag(note.Tags, query) {
			results = append(results, note)
		}
	}

	return results, nil
}

// SearchByTag searches notes by specific tag
func (m *Manager) SearchByTag(tag string) ([]*Note, error) {
	notes, err := m.ListNotes()
	if err != nil {
		return nil, err
	}

	tag = strings.ToLower(strings.TrimSpace(tag))
	var results []*Note

	for _, note := range notes {
		if containsTag(note.Tags, tag) {
			results = append(results, note)
		}
	}

	return results, nil
}

// SearchByDate searches notes by date (supports various formats)
func (m *Manager) SearchByDate(dateQuery string) ([]*Note, error) {
	notes, err := m.ListNotes()
	if err != nil {
		return nil, err
	}

	dateQuery = strings.ToLower(strings.TrimSpace(dateQuery))
	var results []*Note

	// Try to parse the date query
	var targetDate time.Time
	var err2 error

	// Try different date formats
	formats := []string{
		"2006-01-02",
		"2006/01/02",
		"01/02/2006",
		"02/01/2006",
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
	}

	for _, format := range formats {
		targetDate, err2 = time.Parse(format, dateQuery)
		if err2 == nil {
			break
		}
	}

	if err2 != nil {
		// If we can't parse as a specific date, try to match date strings
		for _, note := range notes {
			noteDateStr := note.Created.Format("2006-01-02")
			if strings.Contains(strings.ToLower(noteDateStr), dateQuery) {
				results = append(results, note)
			}
		}
		return results, nil
	}

	// Search for notes created on the target date
	targetDateStart := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())
	targetDateEnd := targetDateStart.Add(24 * time.Hour)

	for _, note := range notes {
		if note.Created.After(targetDateStart) && note.Created.Before(targetDateEnd) {
			results = append(results, note)
		}
	}

	return results, nil
}

// saveNoteToFile saves a note to its file
func (m *Manager) saveNoteToFile(note *Note) error {
	filepath := filepath.Join(m.notesDirs[0], note.Filename)

	var content string
	if note.Format == "org" {
		content = m.formatOrgNote(note)
	} else {
		content = m.formatTxtNote(note)
	}

	return os.WriteFile(filepath, []byte(content), 0644)
}

// loadNoteFromFile loads a note from its file
func (m *Manager) loadNoteFromFile(filePath string) (*Note, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	filename := filepath.Base(filePath)
	ext := filepath.Ext(filename)
	id := strings.TrimSuffix(filename, ext)

	// Parse content based on format
	var title, noteContent string
	var tags []string

	if ext == ".org" {
		title, noteContent, tags = m.parseOrgNote(string(content))
	} else {
		title, noteContent, tags = m.parseTxtNote(string(content))
	}

	// Try to extract creation time from ID
	var created time.Time
	if len(id) >= 15 {
		if t, err := time.Parse("20060102_150405", id[:15]); err == nil {
			created = t
		}
	}
	if created.IsZero() {
		created = time.Now()
	}

	return &Note{
		ID:       id,
		Title:    title,
		Content:  noteContent,
		Created:  created,
		Modified: time.Now(),
		Tags:     tags,
		Format:   strings.TrimPrefix(ext, "."),
		Filename: filename,
	}, nil
}

// formatOrgNote formats a note as Org mode
func (m *Manager) formatOrgNote(note *Note) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("#+TITLE: %s\n", note.Title))
	sb.WriteString(fmt.Sprintf("#+DATE: %s\n", note.Created.Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("#+MODIFIED: %s\n", note.Modified.Format("2006-01-02")))

	if len(note.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("#+TAGS: %s\n", strings.Join(note.Tags, " ")))
	}

	sb.WriteString("\n")
	sb.WriteString("* CONTENT\n")
	sb.WriteString(strings.ReplaceAll(note.Content, "\\n", "\n"))

	return sb.String()
}

// formatTxtNote formats a note as plain text
func (m *Manager) formatTxtNote(note *Note) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Title: %s\n", note.Title))
	sb.WriteString(fmt.Sprintf("Created: %s\n", note.Created.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("Modified: %s\n", note.Modified.Format("2006-01-02 15:04:05")))

	if len(note.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(note.Tags, ", ")))
	}

	sb.WriteString("\n")
	sb.WriteString(strings.ReplaceAll(note.Content, "\\n", "\n"))

	return sb.String()
}

// parseOrgNote parses an Org mode note
func (m *Manager) parseOrgNote(content string) (title, noteContent string, tags []string) {
	lines := strings.Split(content, "\n")

	// Collect tags in a set to avoid duplicates
	tagSet := map[string]struct{}{}

	// Helper to add tags from a directive string
	addTags := func(tagLine string) {
		// Org filetags can be in forms like ":tag1:tag2:" or "tag1 tag2"
		trimmed := strings.TrimSpace(tagLine)
		if trimmed == "" {
			return
		}
		// Replace colons with spaces to normalize, then split
		normalized := strings.ReplaceAll(trimmed, ":", " ")
		for _, t := range strings.Fields(normalized) {
			if t == "" {
				continue
			}
			t = strings.TrimSpace(t)
			if t == "" {
				continue
			}
			tagSet[t] = struct{}{}
		}
	}

	// Determine content start and extract metadata
	contentStart := -1
	for i, raw := range lines {
		line := strings.TrimSpace(raw)
		upper := strings.ToUpper(line)

		if strings.HasPrefix(upper, "#+TITLE:") {
			// Case-insensitive title directive
			maybe := strings.TrimSpace(line[len("#+TITLE:"):])
			if maybe != "" {
				title = maybe
			}
			continue
		}
		if strings.HasPrefix(upper, "#+FILETAGS:") {
			addTags(line[len("#+FILETAGS:"):])
			continue
		}
		if strings.HasPrefix(upper, "#+TAGS:") {
			addTags(line[len("#+TAGS:"):])
			continue
		}

		// Headline tags like: * Heading text :tag1:tag2:
		if strings.HasPrefix(line, "*") {
			// Find trailing colon block
			lastSpace := strings.LastIndex(line, " ")
			if lastSpace != -1 && lastSpace < len(line)-1 {
				tagBlock := strings.TrimSpace(line[lastSpace+1:])
				if strings.HasPrefix(tagBlock, ":") && strings.HasSuffix(tagBlock, ":") {
					addTags(tagBlock)
				}
			}
		}

		// Determine start of content (first non-directive, non-empty line)
		if contentStart == -1 {
			if line == "" {
				continue
			}
			if strings.HasPrefix(strings.TrimSpace(line), "#+") {
				continue
			}
			contentStart = i
		}
	}

	if contentStart != -1 {
		noteContent = strings.TrimSpace(strings.Join(lines[contentStart:], "\n"))
	}

	// Convert tag set to slice
	for t := range tagSet {
		tags = append(tags, t)
	}

	return title, noteContent, tags
}

// parseTxtNote parses a plain text note
func (m *Manager) parseTxtNote(content string) (title, noteContent string, tags []string) {
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "Title:") {
			title = strings.TrimSpace(strings.TrimPrefix(line, "Title:"))
		} else if strings.HasPrefix(line, "Tags:") {
			tagStr := strings.TrimSpace(strings.TrimPrefix(line, "Tags:"))
			tags = strings.Split(tagStr, ",")
			for j, tag := range tags {
				tags[j] = strings.TrimSpace(tag)
			}
		} else if strings.HasPrefix(line, "Created:") || strings.HasPrefix(line, "Modified:") {
			continue // Skip metadata
		} else if line == "" {
			continue // Skip empty lines
		} else {
			// Start of content
			contentStart := strings.Index(content, line)
			if contentStart != -1 {
				noteContent = strings.TrimSpace(content[contentStart:])
			}
			break
		}
	}

	return title, noteContent, tags
}

// sanitizeTitle creates a filesystem-safe title
func sanitizeTitle(title string) string {
	// Replace spaces and special characters with underscores
	title = strings.ReplaceAll(title, " ", "_")
	title = strings.ReplaceAll(title, "/", "_")
	title = strings.ReplaceAll(title, "\\", "_")
	title = strings.ReplaceAll(title, ":", "_")
	title = strings.ReplaceAll(title, "*", "_")
	title = strings.ReplaceAll(title, "?", "_")
	title = strings.ReplaceAll(title, "\"", "_")
	title = strings.ReplaceAll(title, "<", "_")
	title = strings.ReplaceAll(title, ">", "_")
	title = strings.ReplaceAll(title, "|", "_")

	// Convert to lowercase
	title = strings.ToLower(title)

	// Limit length
	if len(title) > 50 {
		title = title[:50]
	}

	return title
}

// containsTag checks if a tag list contains a specific tag
func containsTag(tags []string, query string) bool {
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}
