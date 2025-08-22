package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"burh/config"
	"burh/notes"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// Model represents the main TUI model
type Model struct {
	notes        []*notes.Note
	selected     int
	searchQuery  string
	searching    bool
	editing      bool
	noteManager  *notes.Manager
	config       *config.Config
	styles       *Styles
	state        string // "list", "edit", "create", "search", "confirm_delete"
	currentNote  *notes.Note
	titleInput   string
	contentInput string
	tagsInput    string
	formatInput  string
	currentField int    // 0=title, 1=tags, 2=format, 3=content
	deleteTarget string // ID of note to be deleted

	// Enhanced search fields
	searchType   string // "keyword", "tag", "date"
	keywordQuery string
	tagQuery     string
	dateQuery    string
	searchField  int // 0=type, 1=keyword, 2=tag, 3=date

	// Pagination fields
	pageSize   int // Number of notes to show per page (29)
	startIndex int // Starting index for current page
}

// Styles contains all the styling for the TUI
type Styles struct {
	primary   lipgloss.Style
	secondary lipgloss.Style
	success   lipgloss.Style
	warning   lipgloss.Style
	error     lipgloss.Style
	info      lipgloss.Style
	muted     lipgloss.Style
	title     lipgloss.Style
	item      lipgloss.Style
	selected  lipgloss.Style
	border    lipgloss.Style
}

// NewStyles creates new styles based on config
func NewStyles(cfg *config.Config) *Styles {
	return &Styles{
		primary:   lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.Theme.Primary)).Bold(true),
		secondary: lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.Theme.Secondary)),
		success:   lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.Theme.Success)),
		warning:   lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.Theme.Warning)),
		error:     lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.Theme.Error)),
		info:      lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.Theme.Info)),
		muted:     lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.Theme.Muted)),
		title:     lipgloss.NewStyle().Bold(true),
		item:      lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true),
		selected:  lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.Theme.Success)),
		border:    lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color(cfg.Theme.Primary)),
	}
}

// NewModel creates a new TUI model
func NewModel(noteManager *notes.Manager, cfg *config.Config) *Model {
	return &Model{
		notes:        []*notes.Note{},
		selected:     0,
		searchQuery:  "",
		searching:    false,
		editing:      false,
		noteManager:  noteManager,
		config:       cfg,
		styles:       NewStyles(cfg),
		state:        "list",
		titleInput:   "",
		contentInput: "",
		tagsInput:    "",
		formatInput:  "txt",
		currentField: 0,
		deleteTarget: "",

		// Enhanced search fields
		searchType:   "keyword",
		keywordQuery: "",
		tagQuery:     "",
		dateQuery:    "",
		searchField:  0,

		// Pagination fields
		pageSize:   29, // Changed from 15 to 29 notes per page
		startIndex: 0,
	}
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return tea.Cmd(m.loadNotes)
}

// Update handles user input and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case "list":
			return m.handleListKey(msg)
		case "search":
			return m.handleSearchKey(msg)
		case "edit":
			return m.handleEditKey(msg)
		case "create":
			return m.handleCreateKey(msg)
		case "confirm_delete":
			return m.handleConfirmDeleteKey(msg)
		}
	case notesLoadedMsg:
		m.notes = msg.notes
		// Reset pagination when notes are loaded
		m.selected = 0
		m.startIndex = 0
		return m, nil
	case editorClosedMsg:
		return m, tea.Cmd(m.loadNotes)
	case errorMsg:
		// Handle error - could show a notification
		return m, nil
	}
	return m, nil
}

// View renders the TUI
func (m *Model) View() string {
	switch m.state {
	case "list":
		return m.renderList()
	case "search":
		return m.renderSearch()
	case "edit":
		return m.renderEdit()
	case "create":
		return m.renderCreate()
	case "confirm_delete":
		return m.renderConfirmDelete()
	default:
		return m.renderList()
	}
}

// handleListKey handles key events in list mode
func (m *Model) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "j", "down":
		if m.selected < len(m.notes)-1 {
			m.selected++
			// Adjust page if needed
			if m.selected >= m.startIndex+m.pageSize {
				m.startIndex = m.selected - m.pageSize + 1
			}
		}
	case "k", "up":
		if m.selected > 0 {
			m.selected--
			// Adjust page if needed
			if m.selected < m.startIndex {
				m.startIndex = m.selected
			}
		}
	case "J":
		// Jump to bottom of list
		if len(m.notes) > 0 {
			m.selected = len(m.notes) - 1
			// Adjust page to show the bottom
			if len(m.notes) > m.pageSize {
				m.startIndex = len(m.notes) - m.pageSize
			} else {
				m.startIndex = 0
			}
			// Ensure startIndex doesn't go negative
			if m.startIndex < 0 {
				m.startIndex = 0
			}
		}
	case "K":
		// Jump to top of list
		m.selected = 0
		m.startIndex = 0
	case "enter":
		if len(m.notes) > 0 && m.selected < len(m.notes) {
			n := m.notes[m.selected]
			fullPath := filepath.Join(m.noteManager.GetNotesDir(), n.Filename)
			return m, openEditorCmd(fullPath)
		}
	case "n":
		m.state = "create"
		m.titleInput = ""
		m.contentInput = ""
		m.tagsInput = ""
		m.formatInput = "txt"
		m.currentField = 0
	case "s":
		m.state = "search"
		m.searchQuery = ""
		m.searchType = "keyword"
		m.keywordQuery = ""
		m.tagQuery = ""
		m.dateQuery = ""
		m.searchField = 0
	case "d":
		if len(m.notes) > 0 && m.selected < len(m.notes) {
			m.deleteTarget = m.notes[m.selected].ID
			m.state = "confirm_delete"
		}
	case "r":
		return m, tea.Cmd(m.loadNotes)
	}
	return m, nil
}

// handleSearchKey handles key events in search mode
func (m *Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = "list"
		m.searchQuery = ""
		m.searchType = "keyword"
		m.keywordQuery = ""
		m.tagQuery = ""
		m.dateQuery = ""
		m.searchField = 0
	case "enter":
		// Perform search based on current search type and fields
		m.performSearch()
		m.state = "list"
	case "tab":
		// Cycle through search fields
		m.searchField = (m.searchField + 1) % 4
	case "shift+tab":
		// Cycle backwards through search fields
		m.searchField = (m.searchField - 1 + 4) % 4
	case "backspace":
		// Handle backspace for current search field
		switch m.searchField {
		case 0: // search type
			// Cycle through search types
			switch m.searchType {
			case "keyword":
				m.searchType = "date"
			case "tag":
				m.searchType = "keyword"
			case "date":
				m.searchType = "tag"
			}
		case 1: // keyword query
			if len(m.keywordQuery) > 0 {
				m.keywordQuery = m.keywordQuery[:len(m.keywordQuery)-1]
			}
		case 2: // tag query
			if len(m.tagQuery) > 0 {
				m.tagQuery = m.tagQuery[:len(m.tagQuery)-1]
			}
		case 3: // date query
			if len(m.dateQuery) > 0 {
				m.dateQuery = m.dateQuery[:len(m.dateQuery)-1]
			}
		}
	case "space":
		// Toggle search type when on search type field
		if m.searchField == 0 {
			switch m.searchType {
			case "keyword":
				m.searchType = "tag"
			case "tag":
				m.searchType = "date"
			case "date":
				m.searchType = "keyword"
			}
		} else {
			// Add space to current field
			switch m.searchField {
			case 1:
				m.keywordQuery += " "
			case 2:
				m.tagQuery += " "
			case 3:
				m.dateQuery += " "
			}
		}
	default:
		// Handle regular text input
		if len(msg.String()) == 1 {
			switch m.searchField {
			case 0: // search type - ignore text input
				break
			case 1: // keyword query
				m.keywordQuery += msg.String()
			case 2: // tag query
				m.tagQuery += msg.String()
			case 3: // date query
				m.dateQuery += msg.String()
			}
		}
	}
	return m, nil
}

// handleEditKey handles key events in edit mode
func (m *Model) handleEditKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = "list"
	case "ctrl+s":
		m.saveNote()
		m.state = "list"
		return m, tea.Cmd(m.loadNotes)
	case "tab":
		// Cycle through input fields
		// This is a simplified version - in a real app you'd have more sophisticated field management
	}
	return m, nil
}

// handleCreateKey handles key events in create mode
func (m *Model) handleCreateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = "list"
		m.currentField = 0
	case "ctrl+s":
		m.createNote()
		m.state = "list"
		m.currentField = 0
		return m, tea.Cmd(m.loadNotes)
	case "tab":
		// Cycle through input fields
		m.currentField = (m.currentField + 1) % 4
	case "shift+tab":
		// Cycle backwards through input fields
		m.currentField = (m.currentField - 1 + 4) % 4
	case "backspace":
		// Handle backspace for current field
		switch m.currentField {
		case 0: // title
			if len(m.titleInput) > 0 {
				m.titleInput = m.titleInput[:len(m.titleInput)-1]
			}
		case 1: // tags
			if len(m.tagsInput) > 0 {
				m.tagsInput = m.tagsInput[:len(m.tagsInput)-1]
			}
		case 2: // format
			if len(m.formatInput) > 0 {
				m.formatInput = m.formatInput[:len(m.formatInput)-1]
			}
		case 3: // content
			if len(m.contentInput) > 0 {
				m.contentInput = m.contentInput[:len(m.contentInput)-1]
			}
		}
	case "enter":
		// Move to next field or save if on content field
		if m.currentField == 3 {
			m.createNote()
			m.state = "list"
			m.currentField = 0
			return m, tea.Cmd(m.loadNotes)
		} else {
			m.currentField = (m.currentField + 1) % 4
		}
	default:
		// Handle regular text input
		if len(msg.String()) == 1 {
			switch m.currentField {
			case 0: // title
				m.titleInput += msg.String()
			case 1: // tags
				m.tagsInput += msg.String()
			case 2: // format
				m.formatInput += msg.String()
			case 3: // content
				m.contentInput += msg.String()
			}
		}
	}
	return m, nil
}

// handleConfirmDeleteKey handles key events in confirm delete mode
func (m *Model) handleConfirmDeleteKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		if m.deleteTarget != "" {
			m.deleteNote(m.deleteTarget)
		}
		m.state = "list"
		m.deleteTarget = ""
	case "n":
		m.state = "list"
		m.deleteTarget = ""
	}
	return m, nil
}

// getTerminalWidth returns the width of the terminal
func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80 // Default width if we can't get terminal size
	}
	return width
}

// centerText centers text within the given width and returns the centered text and its original length
func centerText(text string, width int) (string, int) {
	if len(text) >= width {
		return text, len(text)
	}
	padding := (width - len(text)) / 2
	centered := strings.Repeat(" ", padding) + text
	return centered, len(text)
}

// renderList renders the note list view
func (m *Model) renderList() string {
	var sb strings.Builder

	// Header - centered
	terminalWidth := getTerminalWidth()
	headerText := "BURH - NOTE MANAGER"
	centeredHeader, _ := centerText(headerText, terminalWidth)
	header := m.styles.title.Render(centeredHeader)
	sb.WriteString(header)
	sb.WriteString("\n\n")

	// Help text
	help := m.styles.muted.Render("  n: new | s: search | enter: edit | d: delete | r: refresh | q: quit | J: bottom | K: top")
	sb.WriteString(help)
	sb.WriteString("\n\n")

	// Notes list
	if len(m.notes) == 0 {
		sb.WriteString(m.styles.muted.Render("  No notes found. Press 'n' to create a new note."))
	} else {
		// Header row
		header := fmt.Sprintf("  %-16s  %-7s  %-40s  %s", "Date", "Format", "Title", "Tags")
		sb.WriteString(m.styles.primary.Render(header))
		sb.WriteString("\n")

		// Calculate the width to extend to the border
		contentWidth := terminalWidth - 8 // Account for left and right border padding plus 2 spaces
		if contentWidth < 70 {
			contentWidth = 70 // Minimum width
		}
		sb.WriteString(m.styles.muted.Render("  " + strings.Repeat("═", contentWidth)))
		sb.WriteString("\n")

		// Calculate pagination
		totalNotes := len(m.notes)
		endIndex := m.startIndex + m.pageSize
		if endIndex > totalNotes {
			endIndex = totalNotes
		}

		// Show pagination info if there are more notes than page size
		if totalNotes > m.pageSize {
			paginationInfo := fmt.Sprintf("  Showing %d-%d of %d notes", m.startIndex+1, endIndex, totalNotes)
			sb.WriteString(m.styles.muted.Render(paginationInfo))
			sb.WriteString("\n")
		}

		// Add blank line above the first note
		sb.WriteString("\n")

		// Render only the notes for the current page
		for i := m.startIndex; i < endIndex; i++ {
			note := m.notes[i]
			rowStyle := m.styles.item
			if i == m.selected {
				rowStyle = m.styles.selected
			}

			dateStr := note.Created.Format("2006-01-02 15:04")
			formatStr := note.Format
			titleStr := note.Title
			if len(titleStr) > 40 {
				titleStr = titleStr[:37] + "..."
			}
			// Truncate tags to show only first 6
			tagsToShow := note.Tags
			if len(note.Tags) > 6 {
				tagsToShow = note.Tags[:6]
			}
			tagsStr := strings.Join(tagsToShow, ", ")
			if len(note.Tags) > 6 {
				tagsStr += "..."
			}

			row := fmt.Sprintf("  %-16s  %-7s  %-40s  %s", dateStr, formatStr, titleStr, tagsStr)
			sb.WriteString(rowStyle.Render(row))
			sb.WriteString("\n")
		}

		// Show navigation hints if there are more pages
		if totalNotes > m.pageSize {
			sb.WriteString("\n")
			if m.startIndex > 0 {
				sb.WriteString(m.styles.muted.Render("  ↑ Previous page (k/up) "))
			}
			if endIndex < totalNotes {
				sb.WriteString(m.styles.muted.Render("  ↓ Next page (j/down) "))
			}
		}
	}

	return m.styles.border.Render(sb.String())
}

// renderSearch renders the search view
func (m *Model) renderSearch() string {
	var sb strings.Builder

	header := m.styles.title.Render("SEARCH NOTES")
	sb.WriteString(header)
	sb.WriteString("\n\n")

	// Search type field
	typeLabel := "  Search Type: "
	if m.searchField == 0 {
		typeLabel = m.styles.selected.Render("  Search Type: ")
	}
	sb.WriteString(typeLabel)
	sb.WriteString(m.searchType)
	if m.searchField == 0 {
		sb.WriteString(m.styles.selected.Render("█"))
	}
	sb.WriteString("\n")

	// Keyword field
	keywordLabel := "  Keyword: "
	if m.searchField == 1 {
		keywordLabel = m.styles.selected.Render("  Keyword: ")
	}
	sb.WriteString(keywordLabel)
	sb.WriteString(m.keywordQuery)
	if m.searchField == 1 {
		sb.WriteString(m.styles.selected.Render("█"))
	}
	sb.WriteString("\n")

	// Tag field
	tagLabel := "  Tag: "
	if m.searchField == 2 {
		tagLabel = m.styles.selected.Render("  Tag: ")
	}
	sb.WriteString(tagLabel)
	sb.WriteString(m.tagQuery)
	if m.searchField == 2 {
		sb.WriteString(m.styles.selected.Render("█"))
	}
	sb.WriteString("\n")

	// Date field
	dateLabel := "  Date: "
	if m.searchField == 3 {
		dateLabel = m.styles.selected.Render("  Date: ")
	}
	sb.WriteString(dateLabel)
	sb.WriteString(m.dateQuery)
	if m.searchField == 3 {
		sb.WriteString(m.styles.selected.Render("█"))
	}
	sb.WriteString("\n\n")

	help := m.styles.muted.Render("  Tab: Next field | Shift+Tab: Previous field | Space: Toggle search type | Enter: Search | Esc: Cancel")
	sb.WriteString(help)
	sb.WriteString("\n\n")

	// Show search type help
	switch m.searchType {
	case "keyword":
		sb.WriteString(m.styles.info.Render("  Keyword search: Searches in title, content, and tags"))
	case "tag":
		sb.WriteString(m.styles.info.Render("  Tag search: Searches only in note tags"))
	case "date":
		sb.WriteString(m.styles.info.Render("  Date search: Searches by creation date (formats: YYYY-MM-DD, MM/DD/YYYY, etc.)"))
	}

	return m.styles.border.Render(sb.String())
}

// renderEdit renders the edit view
func (m *Model) renderEdit() string {
	var sb strings.Builder

	header := m.styles.title.Render("EDIT NOTE")
	sb.WriteString(header)
	sb.WriteString("\n\n")

	// Title field
	titleLabel := "  Title: "
	if m.currentField == 0 {
		titleLabel = m.styles.selected.Render("  Title: ")
	}
	sb.WriteString(titleLabel)
	sb.WriteString(m.titleInput)
	if m.currentField == 0 {
		sb.WriteString(m.styles.selected.Render("█"))
	}
	sb.WriteString("\n")

	// Tags field
	tagsLabel := "  Tags: "
	if m.currentField == 1 {
		tagsLabel = m.styles.selected.Render("  Tags: ")
	}
	sb.WriteString(tagsLabel)
	sb.WriteString(m.tagsInput)
	if m.currentField == 1 {
		sb.WriteString(m.styles.selected.Render("█"))
	}
	sb.WriteString("\n")

	// Format field
	formatLabel := "  Format: "
	if m.currentField == 2 {
		formatLabel = m.styles.selected.Render("  Format: ")
	}
	sb.WriteString(formatLabel)
	sb.WriteString(m.formatInput)
	if m.currentField == 2 {
		sb.WriteString(m.styles.selected.Render("█"))
	}
	sb.WriteString("\n")

	sb.WriteString("\n")

	// Content field
	contentLabel := "  Content: "
	if m.currentField == 3 {
		contentLabel = m.styles.selected.Render("  Content: ")
	}
	sb.WriteString(contentLabel)
	sb.WriteString("\n")
	sb.WriteString("  " + m.contentInput)
	if m.currentField == 3 {
		sb.WriteString(m.styles.selected.Render("█"))
	}
	sb.WriteString("\n\n")

	help := m.styles.muted.Render("  Tab: Next field | Shift+Tab: Previous field | Enter: Next/Save | Ctrl+S: Save | Esc: Cancel")
	sb.WriteString(help)

	return m.styles.border.Render(sb.String())
}

// renderCreate renders the create view
func (m *Model) renderCreate() string {
	var sb strings.Builder

	header := m.styles.title.Render("CREATE NEW NOTE")
	sb.WriteString(header)
	sb.WriteString("\n\n")

	// Title field
	titleLabel := "  Title: "
	if m.currentField == 0 {
		titleLabel = m.styles.selected.Render("  Title: ")
	}
	sb.WriteString(titleLabel)
	sb.WriteString(m.titleInput)
	if m.currentField == 0 {
		sb.WriteString(m.styles.selected.Render("█"))
	}
	sb.WriteString("\n")

	// Tags field
	tagsLabel := "  Tags: "
	if m.currentField == 1 {
		tagsLabel = m.styles.selected.Render("  Tags: ")
	}
	sb.WriteString(tagsLabel)
	sb.WriteString(m.tagsInput)
	if m.currentField == 1 {
		sb.WriteString(m.styles.selected.Render("█"))
	}
	sb.WriteString("\n")

	// Format field
	formatLabel := "  Format: "
	if m.currentField == 2 {
		formatLabel = m.styles.selected.Render("  Format: ")
	}
	sb.WriteString(formatLabel)
	sb.WriteString(m.formatInput)
	if m.currentField == 2 {
		sb.WriteString(m.styles.selected.Render("█"))
	}
	sb.WriteString("\n")

	sb.WriteString("\n")

	// Content field
	contentLabel := "  Content: "
	if m.currentField == 3 {
		contentLabel = m.styles.selected.Render("  Content: ")
	}
	sb.WriteString(contentLabel)
	sb.WriteString("\n")
	sb.WriteString("  " + m.contentInput)
	if m.currentField == 3 {
		sb.WriteString(m.styles.selected.Render("█"))
	}
	sb.WriteString("\n\n")

	help := m.styles.muted.Render("  Tab: Next field | Shift+Tab: Previous field | Enter: Next/Save | Ctrl+S: Save | Esc: Cancel")
	sb.WriteString(help)

	return m.styles.border.Render(sb.String())
}

// renderConfirmDelete renders the confirmation view for deleting a note
func (m *Model) renderConfirmDelete() string {
	var sb strings.Builder

	header := m.styles.title.Render("CONFIRM DELETE")
	sb.WriteString(header)
	sb.WriteString("\n\n")

	message := fmt.Sprintf("  Are you sure you want to delete note '%s'? This action cannot be undone.", m.notes[m.selected].Title)
	sb.WriteString(m.styles.warning.Render(message))
	sb.WriteString("\n\n")

	help := m.styles.muted.Render("  Y: Confirm | N: Cancel")
	sb.WriteString(help)

	return m.styles.border.Render(sb.String())
}

// loadNotes loads all notes
func (m *Model) loadNotes() tea.Msg {
	notes, err := m.noteManager.ListNotes()
	if err != nil {
		return errorMsg{err}
	}
	return notesLoadedMsg{notes}
}

// searchNotes searches for notes
func (m *Model) searchNotes(query string) {
	results, err := m.noteManager.SearchNotes(query)
	if err != nil {
		return
	}
	m.notes = results
	m.selected = 0
}

// performSearch performs search based on current search type and fields
func (m *Model) performSearch() {
	var results []*notes.Note
	var err error

	switch m.searchType {
	case "keyword":
		if m.keywordQuery != "" {
			results, err = m.noteManager.SearchNotes(m.keywordQuery)
		}
	case "tag":
		if m.tagQuery != "" {
			results, err = m.noteManager.SearchByTag(m.tagQuery)
		}
	case "date":
		if m.dateQuery != "" {
			results, err = m.noteManager.SearchByDate(m.dateQuery)
		}
	}

	if err != nil {
		return
	}

	if results != nil {
		m.notes = results
		m.selected = 0
		m.startIndex = 0 // Reset pagination for search results
	}
}

// saveNote saves the current note
func (m *Model) saveNote() {
	if m.currentNote == nil {
		return
	}

	tags := strings.Split(m.tagsInput, ",")
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}

	m.noteManager.UpdateNote(m.currentNote.ID, m.titleInput, m.contentInput, tags)
}

// createNote creates a new note
func (m *Model) createNote() {
	if m.titleInput == "" {
		return
	}

	tags := strings.Split(m.tagsInput, ",")
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}

	m.noteManager.CreateNote(m.titleInput, m.contentInput, tags, m.formatInput)
}

// deleteNote deletes a note
func (m *Model) deleteNote(id string) {
	err := m.noteManager.DeleteNote(id)
	if err != nil {
		// Could show an error message here
		return
	}
	// Reload notes to reflect the deletion
	m.notes, _ = m.noteManager.ListNotes()
	// Adjust selected index if needed
	if m.selected >= len(m.notes) && len(m.notes) > 0 {
		m.selected = len(m.notes) - 1
	}
	// Reset pagination when notes are reloaded
	m.startIndex = 0
}

// Message types
type notesLoadedMsg struct {
	notes []*notes.Note
}

type errorMsg struct {
	err error
}

// message emitted when the editor closes
type editorClosedMsg struct{}

// openEditorCmd opens the given file in the user's preferred editor and waits for it to close
func openEditorCmd(path string) tea.Cmd {
	return func() tea.Msg {
		editor := os.Getenv("VISUAL")
		if editor == "" {
			editor = os.Getenv("EDITOR")
		}

		var cmd *exec.Cmd
		if editor != "" {
			cmd = exec.Command(editor, path)
		} else {
			// Fallback to OS default opener
			switch runtime.GOOS {
			case "darwin":
				cmd = exec.Command("open", path)
			case "linux":
				cmd = exec.Command("xdg-open", path)
			case "windows":
				cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", path)
			default:
				// If unknown OS, do nothing gracefully
				return editorClosedMsg{}
			}
		}

		_ = cmd.Run()
		return editorClosedMsg{}
	}
}
