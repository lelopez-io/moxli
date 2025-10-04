package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lelopez-io/moxli/internal/bookmark"
	"github.com/lelopez-io/moxli/internal/importer"
	"github.com/lelopez-io/moxli/internal/merge"
	"github.com/lelopez-io/moxli/internal/session"
)

// Lipgloss styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Bold(true)

	urlStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(1, 0, 0, 2)

	statStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("63"))
)

// View represents the different screens in the TUI
type View int

const (
	WelcomeView View = iota
	FileSelectionView
	BrowserView
	DetailView
)

// welcomeChoice represents the user's selection on the welcome screen
type welcomeChoice int

const (
	continueSession welcomeChoice = iota
	newSession
)

// fileSelectionMode represents the current mode in file selection
type fileSelectionMode int

const (
	inputMode fileSelectionMode = iota
	selectionMode
)

// Model is the main application state
type Model struct {
	// Current view
	currentView View

	// Session management
	sessionMgr     *session.Manager
	currentSession *session.Session
	hasSession     bool

	// Welcome screen state
	welcomeSelected welcomeChoice

	// File selection state
	fileSelectionMode fileSelectionMode
	pathInput         textinput.Model
	fileDiscovery     *FileDiscovery
	fileSelectedIdx   int

	// Browser state
	collection       *bookmark.Collection
	browserOffset    int // Scroll offset for browser list
	browserSelected  int // Currently selected bookmark index
	filterMode       bool
	filterInput      textinput.Model
	filteredBookmarks []*bookmark.Bookmark

	// Application state
	width  int
	height int

	// Error state
	err error
}

// NewModel creates a new TUI model
func NewModel() (*Model, error) {
	sessionMgr, err := session.NewManager()
	if err != nil {
		return nil, err
	}

	// Check if a previous session exists
	hasSession := sessionMgr.Exists()
	var currentSession *session.Session
	sessionValid := false

	if hasSession {
		// Load the session to display info
		sess, err := sessionMgr.Load()
		if err != nil {
			return nil, fmt.Errorf("failed to load session: %w", err)
		}
		currentSession = sess

		// Validate that session paths still exist
		workingDirExists := pathExists(sess.WorkingDir)
		currentFileExists := pathExists(sess.CurrentFile)
		sessionValid = workingDirExists && currentFileExists
	}

	// Initialize text input for path entry
	ti := textinput.New()
	ti.Placeholder = "Enter directory or file path (e.g., ~/Downloads)"
	ti.CharLimit = 256
	ti.Width = 60

	// Initialize filter input
	filterTI := textinput.New()
	filterTI.Placeholder = "Search bookmarks..."
	filterTI.CharLimit = 100
	filterTI.Width = 60

	model := &Model{
		currentView:       WelcomeView,
		sessionMgr:        sessionMgr,
		hasSession:        hasSession && sessionValid,
		currentSession:    currentSession,
		fileSelectionMode: inputMode,
		pathInput:         ti,
		filterInput:       filterTI,
		fileDiscovery:     NewFileDiscovery(),
		fileSelectedIdx:   0,
	}

	// If no valid session exists, skip welcome and go straight to file selection
	if !hasSession || !sessionValid {
		model.currentView = FileSelectionView
		model.pathInput.Focus()
	}

	return model, nil
}

// pathExists checks if a file or directory exists
func pathExists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global quit only with Ctrl+C
		// 'q' is handled per-view to allow different behaviors
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			// q always quits from any view
			return m, tea.Quit
		}

		// Handle view-specific key presses
		switch m.currentView {
		case WelcomeView:
			return m.updateWelcome(msg)
		case FileSelectionView:
			return m.updateFileSelection(msg)
		case BrowserView:
			return m.updateBrowser(msg)
		case DetailView:
			return m.updateDetail(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m Model) updateWelcome(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.welcomeSelected > 0 {
			m.welcomeSelected--
		}
	case "down", "j":
		if m.welcomeSelected < newSession {
			m.welcomeSelected++
		}
	case "enter":
		switch m.welcomeSelected {
		case continueSession:
			// Load existing session
			sess, err := m.sessionMgr.Load()
			if err != nil {
				m.err = fmt.Errorf("failed to load session: %w", err)
				return m, nil
			}
			m.currentSession = sess

			// Load the collection file from session
			if err := m.loadCollectionFromSession(); err != nil {
				m.err = fmt.Errorf("failed to load collection: %w", err)
				return m, nil
			}

			m.currentView = BrowserView
		case newSession:
			m.currentView = FileSelectionView
		}
	}
	return m, nil
}

func (m Model) updateFileSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.fileSelectionMode {
	case inputMode:
		return m.updateFileSelectionInput(msg)
	case selectionMode:
		return m.updateFileSelectionList(msg)
	}
	return m, nil
}

func (m Model) updateFileSelectionInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "enter":
		// User pressed enter - scan the path
		path := m.pathInput.Value()
		if path != "" {
			err := m.fileDiscovery.DiscoverPath(path)
			if err != nil {
				m.err = err
				return m, nil
			}

			// If files were found, switch to selection mode
			if len(m.fileDiscovery.Files()) > 0 {
				m.fileSelectionMode = selectionMode
				m.fileSelectedIdx = 0
				m.pathInput.Blur()
			} else {
				m.err = fmt.Errorf("no bookmark files found in %s", path)
			}
		}
		return m, nil

	case "ctrl+r":
		// Reset and allow new path
		m.fileDiscovery.Clear()
		m.pathInput.SetValue("")
		m.fileSelectionMode = inputMode
		m.pathInput.Focus()
		m.err = nil
		return m, nil
	}

	// Pass through to text input
	m.pathInput, cmd = m.pathInput.Update(msg)
	return m, cmd
}

func (m Model) updateFileSelectionList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	files := m.fileDiscovery.Files()
	if len(files) == 0 {
		return m, nil
	}

	switch msg.String() {
	case "up", "k":
		if m.fileSelectedIdx > 0 {
			m.fileSelectedIdx--
		}
	case "down", "j":
		if m.fileSelectedIdx < len(files)-1 {
			m.fileSelectedIdx++
		}
	case " ":
		// Toggle selection
		files[m.fileSelectedIdx].Selected = !files[m.fileSelectedIdx].Selected
	case "b":
		// Mark as base (clear other base flags first)
		for _, f := range files {
			f.IsBase = false
		}
		files[m.fileSelectedIdx].IsBase = true
		files[m.fileSelectedIdx].Selected = true
	case "enter":
		// Confirm selection and proceed
		// Validate that we have at least one base and one source
		hasBase := false
		hasSource := false
		for _, f := range files {
			if f.IsBase {
				hasBase = true
			}
			if f.Selected && !f.IsBase {
				hasSource = true
			}
		}

		if !hasBase {
			m.err = fmt.Errorf("please mark one file as base with 'b'")
			return m, nil
		}
		if !hasSource {
			m.err = fmt.Errorf("please select at least one source file with space")
			return m, nil
		}

		// Load and merge files
		if err := m.loadAndMergeFiles(); err != nil {
			m.err = fmt.Errorf("merge failed: %w", err)
			return m, nil
		}

		// Proceed to browser view
		m.currentView = BrowserView
		return m, nil

	case "ctrl+r":
		// Reset and go back to path input
		m.fileDiscovery.Clear()
		m.pathInput.SetValue("")
		m.fileSelectionMode = inputMode
		m.pathInput.Focus()
		m.err = nil
		return m, nil
	}

	return m, nil
}

func (m Model) updateBrowser(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.collection == nil || len(m.collection.Bookmarks) == 0 {
		return m, nil
	}

	// Handle filter mode
	if m.filterMode {
		switch msg.String() {
		case "esc":
			// Exit filter mode
			m.filterMode = false
			m.filterInput.Blur()
			m.filterInput.SetValue("")
			m.filteredBookmarks = nil
			m.browserSelected = 0
			m.browserOffset = 0
			return m, nil
		case "enter":
			// Apply filter
			m.filterMode = false
			m.filterInput.Blur()
			m.applyFilter()
			return m, nil
		default:
			// Update filter input
			var cmd tea.Cmd
			m.filterInput, cmd = m.filterInput.Update(msg)
			return m, cmd
		}
	}

	// Get current bookmark list (filtered or full)
	bookmarks := m.collection.Bookmarks
	if m.filteredBookmarks != nil {
		bookmarks = m.filteredBookmarks
	}
	bookmarkCount := len(bookmarks)

	pageSize := 10
	halfPage := pageSize / 2

	switch msg.String() {
	case "/":
		// Enter filter mode
		m.filterMode = true
		m.filterInput.Focus()
		return m, nil
	case "up", "k":
		if m.browserSelected > 0 {
			m.browserSelected--
			// Scroll up if needed
			if m.browserSelected < m.browserOffset {
				m.browserOffset = m.browserSelected
			}
		}
	case "down", "j":
		if m.browserSelected < bookmarkCount-1 {
			m.browserSelected++
			// Scroll down if needed (show 10 items at a time)
			if m.browserSelected >= m.browserOffset+pageSize {
				m.browserOffset = m.browserSelected - (pageSize - 1)
			}
		}
	case "g":
		// Go to top
		m.browserSelected = 0
		m.browserOffset = 0
	case "G":
		// Go to bottom
		m.browserSelected = bookmarkCount - 1
		m.browserOffset = max(0, m.browserSelected-(pageSize-1))
	case "J":
		// Scroll down half page (Shift-J for bigger jump)
		m.browserSelected = min(m.browserSelected+halfPage, bookmarkCount-1)
		// Center the selection in the window
		m.browserOffset = max(0, min(m.browserSelected-(pageSize/2), bookmarkCount-pageSize))
	case "K":
		// Scroll up half page (Shift-K for bigger jump)
		m.browserSelected = max(0, m.browserSelected-halfPage)
		// Center the selection in the window
		m.browserOffset = max(0, min(m.browserSelected-(pageSize/2), bookmarkCount-pageSize))
	case "h":
		// Scroll up full page - land at top of new page
		m.browserOffset = max(0, m.browserOffset-pageSize)
		m.browserSelected = m.browserOffset
	case "H":
		// Scroll down full page - land at top of new page
		m.browserOffset = min(m.browserOffset+pageSize, bookmarkCount-pageSize)
		if m.browserOffset < 0 {
			m.browserOffset = 0
		}
		m.browserSelected = m.browserOffset
	case " ":
		// Toggle detail preview overlay
		m.currentView = DetailView
		return m, nil
	}

	return m, nil
}

func (m Model) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case " ":
		// Close preview overlay and return to browser
		m.currentView = BrowserView
		return m, nil
	case "enter":
		// Open URL in default browser
		bookmarks := m.collection.Bookmarks
		if m.filteredBookmarks != nil {
			bookmarks = m.filteredBookmarks
		}
		if m.browserSelected < len(bookmarks) {
			bm := bookmarks[m.browserSelected]
			if bm.URL != "" {
				// Open URL in default browser (macOS uses 'open')
				cmd := exec.Command("open", bm.URL)
				return m, tea.ExecProcess(cmd, nil)
			}
		}
		return m, nil
	}

	return m, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// View renders the current view
func (m Model) View() string {
	switch m.currentView {
	case WelcomeView:
		return m.welcomeView()
	case FileSelectionView:
		return m.fileSelectionView()
	case BrowserView:
		return m.browserView()
	case DetailView:
		return m.detailView()
	default:
		return "Unknown view"
	}
}

func (m Model) welcomeView() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit", m.err)
	}

	s := "\n"
	s += "  ╔═══════════════════════════════════════╗\n"
	s += "  ║                                       ║\n"
	s += "  ║        📚 Moxli (Amoxtli)            ║\n"
	s += "  ║     Bookmark Management System        ║\n"
	s += "  ║                                       ║\n"
	s += "  ╚═══════════════════════════════════════╝\n"
	s += "\n"

	if m.hasSession && m.currentSession != nil {
		s += "  Previous session found:\n"
		s += fmt.Sprintf("  Working Directory: %s\n", m.currentSession.WorkingDir)
		s += fmt.Sprintf("  Current File: %s\n", m.currentSession.CurrentFile)
		s += fmt.Sprintf("  Merge History: %d records\n", len(m.currentSession.MergeHistory))
		s += "\n"
	}

	s += "  What would you like to do?\n\n"

	// Continue session option
	if m.welcomeSelected == continueSession {
		s += "  ▶ Continue previous session\n"
	} else {
		s += "    Continue previous session\n"
	}

	// New session option
	if m.welcomeSelected == newSession {
		s += "  ▶ Start new session\n"
	} else {
		s += "    Start new session\n"
	}

	s += "\n"
	s += "  ↑/k: up  ↓/j: down  enter: select  q: quit\n"

	return s
}

func (m Model) fileSelectionView() string {
	s := "\n"
	s += "  ╔═══════════════════════════════════════════════════════════════════╗\n"
	s += "  ║                     📁 File Discovery                            ║\n"
	s += "  ╚═══════════════════════════════════════════════════════════════════╝\n"
	s += "\n"

	if m.err != nil {
		s += fmt.Sprintf("  ⚠️  Error: %v\n\n", m.err)
	}

	switch m.fileSelectionMode {
	case inputMode:
		s += "  Enter a directory path or individual file path to scan for bookmarks.\n"
		s += "  Supported: Anybox JSON, Anybox HTML, Firefox HTML, Safari HTML\n\n"
		s += "  Path: " + m.pathInput.View() + "\n\n"
		s += "  Press enter to scan  |  ctrl+r to reset  |  q to quit\n"

	case selectionMode:
		files := m.fileDiscovery.Files()
		s += fmt.Sprintf("  Found %d bookmark file(s):\n\n", len(files))

		for i, file := range files {
			cursor := "  "
			if i == m.fileSelectedIdx {
				cursor = "▶ "
			}

			marker := "[ ]"
			if file.IsBase {
				marker = "[B]"
			} else if file.Selected {
				marker = "[✓]"
			}

			formatStr := string(file.Format)
			s += fmt.Sprintf("%s%s %-12s  %s\n", cursor, marker, formatStr, file.Path)
		}

		s += "\n"
		s += "  ↑/k: up  ↓/j: down  space: toggle  b: mark as base\n"
		s += "  enter: continue  ctrl+r: reset  q: quit\n"
	}

	return s
}

// loadAndMergeFiles loads selected files and performs merge
func (m *Model) loadAndMergeFiles() error {
	files := m.fileDiscovery.Files()

	// Find base and source files
	var baseFile *DiscoveredFile
	var sourceFiles []*DiscoveredFile

	for _, f := range files {
		if f.IsBase {
			baseFile = f
		} else if f.Selected {
			sourceFiles = append(sourceFiles, f)
		}
	}

	// Load base collection
	baseCollection, err := m.loadFile(baseFile)
	if err != nil {
		return fmt.Errorf("failed to load base file: %w", err)
	}

	// Load source collections
	sourceCollections := make([]*bookmark.Collection, 0, len(sourceFiles))
	for _, sf := range sourceFiles {
		coll, err := m.loadFile(sf)
		if err != nil {
			return fmt.Errorf("failed to load source file %s: %w", sf.Path, err)
		}
		sourceCollections = append(sourceCollections, coll)
	}

	// Perform merge
	merger := merge.New(baseCollection, sourceCollections...)
	merged, err := merger.Merge()
	if err != nil {
		return fmt.Errorf("merge operation failed: %w", err)
	}

	m.collection = merged

	// Save session for future continuation
	if err := m.saveSession(); err != nil {
		// Log error but don't fail the merge
		// User can still use the merged collection
		return fmt.Errorf("warning: failed to save session: %w", err)
	}

	return nil
}

// loadCollectionFromSession loads the collection from the current session
func (m *Model) loadCollectionFromSession() error {
	if m.currentSession == nil || m.currentSession.CurrentFile == "" {
		return fmt.Errorf("no current file in session")
	}

	// Load the collection from the current file
	file, err := os.Open(m.currentSession.CurrentFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Assume it's Anybox JSON format (the merged output format)
	imp := &importer.AnyboxImporter{}
	collection, err := imp.Parse(file)
	if err != nil {
		return err
	}

	m.collection = collection
	return nil
}

// saveSession saves the current session state
func (m *Model) saveSession() error {
	files := m.fileDiscovery.Files()

	// Find base and source files
	var baseFile *DiscoveredFile
	var sourceFiles []*DiscoveredFile

	for _, f := range files {
		if f.IsBase {
			baseFile = f
		} else if f.Selected {
			sourceFiles = append(sourceFiles, f)
		}
	}

	if baseFile == nil {
		return fmt.Errorf("no base file")
	}

	// Determine working directory from base file path
	workingDir := baseFile.Path
	if idx := strings.LastIndex(workingDir, "/"); idx >= 0 {
		workingDir = workingDir[:idx]
	}

	// Create or update session
	sess := &session.Session{
		WorkingDir:  workingDir,
		CurrentFile: baseFile.Path,
	}

	// Add merge record
	sourcePaths := make([]string, len(sourceFiles))
	for i, sf := range sourceFiles {
		sourcePaths[i] = sf.Path
	}

	sess.AddMergeRecord(baseFile.Path, sourcePaths, len(m.collection.Bookmarks))

	m.currentSession = sess
	return m.sessionMgr.Save(sess)
}

// loadFile loads a bookmark file using the appropriate importer
func (m *Model) loadFile(f *DiscoveredFile) (*bookmark.Collection, error) {
	file, err := os.Open(f.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var imp importer.Importer
	switch f.Format {
	case FormatAnybox:
		imp = &importer.AnyboxImporter{}
	case FormatAnyboxHTML:
		imp = &importer.AnyboxHTMLImporter{}
	case FormatFirefox:
		imp = &importer.FirefoxImporter{}
	case FormatSafari:
		imp = &importer.SafariImporter{}
	default:
		return nil, fmt.Errorf("unknown format: %s", f.Format)
	}

	return imp.Parse(file)
}

func (m Model) browserView() string {
	if m.collection == nil {
		return "\n" + titleStyle.Render("📖 Bookmark Browser") + "\n\n  No collection loaded.\n\n  Press q to quit\n"
	}

	var s strings.Builder

	// Header
	s.WriteString("\n")
	s.WriteString(headerStyle.Render("📖 Bookmark Browser"))
	s.WriteString("\n\n")

	// Determine which bookmarks to show
	bookmarks := m.collection.Bookmarks
	if m.filteredBookmarks != nil {
		bookmarks = m.filteredBookmarks
	}

	// Stats
	stats := fmt.Sprintf("Total: %d bookmarks │ Selected: %d/%d",
		len(bookmarks), m.browserSelected+1, len(bookmarks))
	if m.filteredBookmarks != nil {
		stats = fmt.Sprintf("Filtered: %d/%d bookmarks │ Selected: %d/%d",
			len(m.filteredBookmarks), len(m.collection.Bookmarks), m.browserSelected+1, len(bookmarks))
	}
	s.WriteString("  " + statStyle.Render(stats) + "\n")

	// Filter input (if active)
	if m.filterMode {
		s.WriteString("\n  " + m.filterInput.View() + "\n")
		s.WriteString("  " + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("enter: apply  esc: cancel") + "\n")
	}
	s.WriteString("\n")

	// Show window of bookmarks (10 at a time)
	pageSize := 10
	start := m.browserOffset
	end := min(start+pageSize, len(bookmarks))

	for i := start; i < end; i++ {
		bm := bookmarks[i]

		title := bm.Title
		if title == "" {
			title = "(no title)"
		}

		// Render item
		if i == m.browserSelected {
			s.WriteString("  ▶ " + selectedItemStyle.Render(title) + "\n")
		} else {
			s.WriteString("    " + title + "\n")
		}

		if bm.URL != "" {
			s.WriteString("      " + urlStyle.Render(bm.URL) + "\n")
		}
		s.WriteString("\n")
	}

	// Help text
	help := `Navigation:
  j/↓: down         k/↑: up
  J: half-page down K: half-page up
  H: page down      h: page up
  G: bottom         g: top
  space: preview    /: filter

q: quit`
	s.WriteString(helpStyle.Render(help))

	return s.String()
}

// applyFilter filters bookmarks based on search query
func (m *Model) applyFilter() {
	query := strings.ToLower(strings.TrimSpace(m.filterInput.Value()))
	if query == "" {
		m.filteredBookmarks = nil
		return
	}

	var filtered []*bookmark.Bookmark
	for _, bm := range m.collection.Bookmarks {
		// Search in title, URL, description, tags
		if strings.Contains(strings.ToLower(bm.Title), query) ||
			strings.Contains(strings.ToLower(bm.URL), query) ||
			strings.Contains(strings.ToLower(bm.Description), query) ||
			strings.Contains(strings.ToLower(bm.Comment), query) ||
			m.matchesTag(bm, query) {
			filtered = append(filtered, bm)
		}
	}

	m.filteredBookmarks = filtered
	m.browserSelected = 0
	m.browserOffset = 0
}

// matchesTag checks if any tag contains the query
func (m *Model) matchesTag(bm *bookmark.Bookmark, query string) bool {
	for _, tagHierarchy := range bm.Tags {
		for _, tag := range tagHierarchy {
			if strings.Contains(strings.ToLower(tag), query) {
				return true
			}
		}
	}
	return false
}

func (m Model) detailView() string {
	if m.collection == nil || len(m.collection.Bookmarks) == 0 {
		return "\nNo bookmark selected\n\nPress esc to return"
	}

	// Get bookmark from filtered list if active
	bookmarks := m.collection.Bookmarks
	if m.filteredBookmarks != nil {
		bookmarks = m.filteredBookmarks
	}

	if m.browserSelected >= len(bookmarks) {
		return "\nNo bookmark selected\n\nPress esc to return"
	}

	bm := bookmarks[m.browserSelected]
	var s strings.Builder

	// Header
	s.WriteString("\n")
	s.WriteString(headerStyle.Render("📑 Bookmark Details"))
	s.WriteString("\n\n")

	// Title
	title := bm.Title
	if title == "" {
		title = "(no title)"
	}
	s.WriteString("  " + selectedItemStyle.Render(title) + "\n\n")

	// URL
	if bm.URL != "" {
		s.WriteString("  " + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("URL:") + "\n")
		s.WriteString("  " + urlStyle.Render(bm.URL) + "\n\n")
	}

	// Description
	if bm.Description != "" {
		s.WriteString("  " + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Description:") + "\n")
		s.WriteString("  " + bm.Description + "\n\n")
	}

	// Comment
	if bm.Comment != "" {
		s.WriteString("  " + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Comment:") + "\n")
		s.WriteString("  " + bm.Comment + "\n\n")
	}

	// Tags
	if len(bm.Tags) > 0 {
		s.WriteString("  " + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Tags:") + "\n")
		for _, tagHierarchy := range bm.Tags {
			tagStr := strings.Join(tagHierarchy, " → ")
			s.WriteString("    • " + lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Render(tagStr) + "\n")
		}
		s.WriteString("\n")
	}

	// Folder
	if len(bm.Folder) > 0 {
		s.WriteString("  " + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Folder:") + "\n")
		folderStr := strings.Join(bm.Folder, " / ")
		s.WriteString("    📁 " + folderStr + "\n\n")
	}

	// Dates
	if !bm.DateAdded.IsZero() {
		s.WriteString("  " + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Date Added:") + "\n")
		s.WriteString("    " + bm.DateAdded.Format("2006-01-02 15:04:05") + "\n\n")
	}

	if !bm.LastModified.IsZero() {
		s.WriteString("  " + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Last Modified:") + "\n")
		s.WriteString("    " + bm.LastModified.Format("2006-01-02 15:04:05") + "\n\n")
	}

	// Metadata
	if bm.IsStarred {
		s.WriteString("  ⭐ Starred\n\n")
	}

	if bm.Keyword != "" {
		s.WriteString("  " + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Keyword:") + " " + bm.Keyword + "\n\n")
	}

	if bm.Source != "" {
		s.WriteString("  " + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Source:") + " " + bm.Source + "\n\n")
	}

	// Help
	s.WriteString(helpStyle.Render("space: close preview  |  enter: open in browser  |  q: quit"))

	return s.String()
}
