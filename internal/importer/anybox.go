package importer

import (
	"encoding/json"
	"io"
	"time"

	"github.com/lelopez-io/moxli/internal/bookmark"
)

// AnyboxImporter handles Anybox JSON export format
type AnyboxImporter struct{}

// Parse reads Anybox JSON and returns a collection
func (a *AnyboxImporter) Parse(r io.Reader) (*bookmark.Collection, error) {
	var bookmarks []*bookmark.Bookmark
	if err := json.NewDecoder(r).Decode(&bookmarks); err != nil {
		return nil, err
	}

	collection := bookmark.NewCollection()
	collection.Metadata.Source = "anybox"
	collection.Metadata.ImportedAt = time.Now()

	// Process each bookmark
	for _, b := range bookmarks {
		b.Source = "anybox"
		b.ImportedAt = time.Now()

		// Normalize URL for deduplication
		if err := bookmark.NormalizeBookmarkURL(b); err != nil {
			// Skip bookmarks with invalid URLs
			continue
		}

		// Normalize tags to lower-kebab-case
		bookmark.NormalizeTags(b)

		collection.Add(b)
	}

	collection.UpdateMetadata()
	return collection, nil
}

// Detect checks if the content is Anybox JSON format
func (a *AnyboxImporter) Detect(r io.Reader) bool {
	var bookmarks []json.RawMessage
	if err := json.NewDecoder(r).Decode(&bookmarks); err != nil {
		return false
	}

	// Check if it's an array with at least one object that has Anybox-specific fields
	if len(bookmarks) > 0 {
		var sample map[string]interface{}
		if err := json.Unmarshal(bookmarks[0], &sample); err != nil {
			return false
		}

		// Look for Anybox-specific fields
		_, hasIsStarred := sample["isStarred"]
		_, hasKeyword := sample["keyword"]
		_, hasTags := sample["tags"]

		return hasIsStarred || hasKeyword || hasTags
	}

	return false
}

// Source returns the source identifier
func (a *AnyboxImporter) Source() string {
	return "anybox"
}
