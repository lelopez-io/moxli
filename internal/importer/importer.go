package importer

import (
	"io"

	"github.com/lelopez-io/moxli/internal/bookmark"
)

// Importer defines the interface for bookmark importers
type Importer interface {
	// Parse reads bookmarks from a reader and returns a collection
	Parse(r io.Reader) (*bookmark.Collection, error)

	// Detect checks if the content matches this importer's format
	Detect(r io.Reader) bool

	// Source returns the source identifier (e.g., "anybox", "firefox", "safari")
	Source() string
}
