package tui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lelopez-io/moxli/internal/session"
)

// View represents the different screens in the TUI
type View int

const (
	WelcomeView View = iota
	FileSelectionView
	BrowserView
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

	// Application state
	// collection *bookmark.Collection // TODO: Will be used when browsing bookmarks
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

	model := &Model{
		currentView:       WelcomeView,
		sessionMgr:        sessionMgr,
		hasSession:        hasSession && sessionValid,
		currentSession:    currentSession,
		fileSelectionMode: inputMode,
		pathInput:         ti,
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
		switch msg.String() {
		case "ctrl+c", "q":
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
			m.currentView = BrowserView // TODO: Should load the collection and go to browser
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

		// TODO: Proceed to merge/browser view
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
	// TODO: Implement browser navigation
	return m, nil
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

func (m Model) browserView() string {
	s := "\n  📖 Bookmark Browser\n\n"
	s += "  Browser view not yet implemented.\n\n"
	s += "  Press q to quit\n"
	return s
}
