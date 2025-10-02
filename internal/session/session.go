package session

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// MergeRecord tracks a single merge operation
type MergeRecord struct {
	BaseFile    string    `yaml:"base_file"`    // Anybox JSON used as base
	SourceFiles []string  `yaml:"source_files"` // Firefox/Safari HTML files merged
	Date        time.Time `yaml:"date"`
	Enhanced    int       `yaml:"enhanced"` // Number of bookmarks with timestamps updated
}

// Session tracks the current working state
type Session struct {
	WorkingDir   string        `yaml:"working_dir"`
	CurrentFile  string        `yaml:"current_file"`  // Active JSON being edited
	LastModified time.Time     `yaml:"last_modified"`
	MergeHistory []MergeRecord `yaml:"merge_history"`
}

// Manager handles session persistence
type Manager struct {
	configDir string
}

// NewManager creates a new session manager
func NewManager() (*Manager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(home, ".moxli")
	return &Manager{configDir: configDir}, nil
}

// ensureConfigDir creates the config directory if it doesn't exist
func (m *Manager) ensureConfigDir() error {
	return os.MkdirAll(m.configDir, 0755)
}

// sessionPath returns the path to the session file
func (m *Manager) sessionPath() string {
	return filepath.Join(m.configDir, "session.yaml")
}

// Load loads the session from disk
func (m *Manager) Load() (*Session, error) {
	path := m.sessionPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No session exists
		}
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var session Session
	if err := yaml.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to parse session file: %w", err)
	}

	return &session, nil
}

// Save saves the session to disk
func (m *Manager) Save(session *Session) error {
	if err := m.ensureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	session.LastModified = time.Now()

	data, err := yaml.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to serialize session: %w", err)
	}

	path := m.sessionPath()
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// Clear removes the session file
func (m *Manager) Clear() error {
	path := m.sessionPath()
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return nil // Already cleared
		}
		return fmt.Errorf("failed to remove session file: %w", err)
	}
	return nil
}

// Exists checks if a session file exists
func (m *Manager) Exists() bool {
	path := m.sessionPath()
	_, err := os.Stat(path)
	return err == nil
}

// AddMergeRecord adds a merge operation to the session history
func (s *Session) AddMergeRecord(baseFile string, sourceFiles []string, enhanced int) {
	record := MergeRecord{
		BaseFile:    baseFile,
		SourceFiles: sourceFiles,
		Date:        time.Now(),
		Enhanced:    enhanced,
	}
	s.MergeHistory = append(s.MergeHistory, record)
}
