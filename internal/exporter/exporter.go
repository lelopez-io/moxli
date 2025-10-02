package exporter

import (
	"io"

	"github.com/lelopez-io/moxli/internal/bookmark"
)

// Exporter defines the interface for bookmark exporters
type Exporter interface {
	// Export writes a collection to a writer
	Export(w io.Writer, c *bookmark.Collection) error
}
