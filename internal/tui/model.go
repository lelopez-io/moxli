package tui

import (
	"fmt"

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

	if hasSession {
		// Load the session to display info
		sess, err := sessionMgr.Load()
		if err != nil {
			return nil, fmt.Errorf("failed to load session: %w", err)
		}
		currentSession = sess
	}

	model := &Model{
		currentView:    WelcomeView,
		sessionMgr:     sessionMgr,
		hasSession:     hasSession,
		currentSession: currentSession,
	}

	// If no session exists, skip welcome and go straight to file selection
	if !hasSession {
		model.currentView = FileSelectionView
	}

	return model, nil
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
	// TODO: Implement file selection logic
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
	s := "\n  📁 File Selection\n\n"
	s += "  TODO: Implement file discovery and selection\n\n"
	s += "  Press q to quit\n"
	return s
}

func (m Model) browserView() string {
	s := "\n  📖 Bookmark Browser\n\n"
	if m.currentSession != nil {
		s += fmt.Sprintf("  Working: %s\n\n", m.currentSession.CurrentFile)
	}
	s += "  TODO: Implement bookmark browsing\n\n"
	s += "  Press q to quit\n"
	return s
}
