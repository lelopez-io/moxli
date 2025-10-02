package tui

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lelopez-io/moxli/internal/importer"
)

// FileFormat represents detected file formats
type FileFormat string

const (
	FormatUnknown FileFormat = "unknown"
	FormatAnybox  FileFormat = "anybox"
	FormatFirefox FileFormat = "firefox"
	FormatSafari  FileFormat = "safari"
)

// DiscoveredFile represents a file found during discovery
type DiscoveredFile struct {
	Path     string
	Format   FileFormat
	Selected bool
	IsBase   bool
}

// FileDiscovery handles file discovery and format detection
type FileDiscovery struct {
	files []*DiscoveredFile
}

// NewFileDiscovery creates a new file discovery instance
func NewFileDiscovery() *FileDiscovery {
	return &FileDiscovery{
		files: make([]*DiscoveredFile, 0),
	}
}

// DiscoverPath scans a path for bookmark files
func (fd *FileDiscovery) DiscoverPath(path string) error {
	// Expand home directory if present
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(home, path[2:])
	}

	// Clean path
	path = filepath.Clean(path)

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path not found: %w", err)
	}

	if info.IsDir() {
		// Scan directory for bookmark files
		return fd.scanDirectory(path)
	}

	// Single file - detect and add
	return fd.addFile(path)
}

// scanDirectory scans a directory for .json and .html files
func (fd *FileDiscovery) scanDirectory(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))

		// Only process .json and .html files
		if ext == ".json" || ext == ".html" {
			fullPath := filepath.Join(dirPath, name)
			if err := fd.addFile(fullPath); err != nil {
				// Log error but continue with other files
				continue
			}
		}
	}

	return nil
}

// addFile adds a file with format detection
func (fd *FileDiscovery) addFile(path string) error {
	format, err := detectFileFormat(path)
	if err != nil {
		return err
	}

	// Only add if we could detect the format
	if format == FormatUnknown {
		return fmt.Errorf("unknown file format")
	}

	fd.files = append(fd.files, &DiscoveredFile{
		Path:     path,
		Format:   format,
		Selected: false,
		IsBase:   false,
	})

	return nil
}

// Files returns all discovered files
func (fd *FileDiscovery) Files() []*DiscoveredFile {
	return fd.files
}

// Clear removes all discovered files
func (fd *FileDiscovery) Clear() {
	fd.files = make([]*DiscoveredFile, 0)
}

// detectFileFormat detects the format of a bookmark file
func detectFileFormat(path string) (FileFormat, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return FormatUnknown, fmt.Errorf("failed to read file: %w", err)
	}

	reader := bytes.NewReader(content)

	// Try Anybox (JSON)
	anyboxImporter := &importer.AnyboxImporter{}
	if anyboxImporter.Detect(reader) {
		return FormatAnybox, nil
	}

	// Reset reader
	if _, err := reader.Seek(0, 0); err != nil {
		return FormatUnknown, fmt.Errorf("failed to reset reader: %w", err)
	}

	// Try Firefox (HTML with timestamps)
	firefoxImporter := &importer.FirefoxImporter{}
	if firefoxImporter.Detect(reader) {
		return FormatFirefox, nil
	}

	// Reset reader
	if _, err := reader.Seek(0, 0); err != nil {
		return FormatUnknown, fmt.Errorf("failed to reset reader: %w", err)
	}

	// Try Safari (HTML without timestamps)
	safariImporter := &importer.SafariImporter{}
	if safariImporter.Detect(reader) {
		return FormatSafari, nil
	}

	return FormatUnknown, nil
}
