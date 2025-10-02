package exporter

import (
	"encoding/json"
	"io"

	"github.com/lelopez-io/moxli/internal/bookmark"
)

// AnyboxExporter handles Anybox JSON export format
type AnyboxExporter struct {
	// Indent enables pretty-printing with indentation
	Indent bool
}

// Export writes the collection as Anybox JSON format
func (a *AnyboxExporter) Export(w io.Writer, c *bookmark.Collection) error {
	encoder := json.NewEncoder(w)
	if a.Indent {
		encoder.SetIndent("", "  ")
	}

	// Export bookmarks array (Anybox format is just an array)
	return encoder.Encode(c.Bookmarks)
}
