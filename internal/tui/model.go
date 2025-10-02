package tui

import (
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

// Model is the main application state
type Model struct {
	// Current view
	currentView View

	// Session management
	sessionMgr *session.Manager
	// session    *session.Session // TODO: Will be used when loading sessions

	// Application state
	// collection *bookmark.Collection // TODO: Will be used when browsing bookmarks
	width  int
	height int

	// Error state
	// err error // TODO: Will be used for error handling
}

// NewModel creates a new TUI model
func NewModel() (*Model, error) {
	sessionMgr, err := session.NewManager()
	if err != nil {
		return nil, err
	}

	return &Model{
		currentView: WelcomeView,
		sessionMgr:  sessionMgr,
	}, nil
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

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

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
	return "Moxli - Bookmark Manager\n\nPress q to quit"
}

func (m Model) fileSelectionView() string {
	return "File Selection View\n\nPress q to quit"
}

func (m Model) browserView() string {
	return "Browser View\n\nPress q to quit"
}
