package merge

import (
	"github.com/lelopez-io/moxli/internal/bookmark"
)

// Merger handles base-centric merge operations
type Merger struct {
	base    *bookmark.Collection
	sources []*bookmark.Collection
}

// New creates a new merger with a base collection and source collections
func New(base *bookmark.Collection, sources ...*bookmark.Collection) *Merger {
	return &Merger{
		base:    base,
		sources: sources,
	}
}

// Merge performs the base-centric merge with timestamp-only enhancement
func (m *Merger) Merge() (*bookmark.Collection, error) {
	// Clone the base to avoid modifying original
	result := m.base.Clone()

	// Ensure URL index is built for fast lookups
	// First lookup will trigger build if needed
	_, _ = result.FindByURL("")

	// Track enhancement statistics
	enhanced := 0

	// Process each source collection
	for _, source := range m.sources {
		for _, sourceBookmark := range source.Bookmarks {
			// Look up bookmark in base by normalized URL
			baseBookmark, exists := result.FindByURL(sourceBookmark.NormalizedURL)
			if exists {
				// URL exists in base - enhance with timestamps only
				if m.enhanceTimestamps(baseBookmark, sourceBookmark) {
					enhanced++
				}
			}
			// If URL not in base, skip (intentionally deleted/not included)
		}
	}

	result.UpdateMetadata()
	return result, nil
}

// enhanceTimestamps updates base bookmark with older timestamps from source
// Returns true if any timestamp was updated
func (m *Merger) enhanceTimestamps(base, source *bookmark.Bookmark) bool {
	updated := false

	// Prefer oldest non-zero DateAdded
	if !source.DateAdded.IsZero() {
		if base.DateAdded.IsZero() || source.DateAdded.Before(base.DateAdded) {
			base.DateAdded = source.DateAdded
			updated = true
		}
	}

	// Prefer oldest non-zero LastModified
	if !source.LastModified.IsZero() {
		if base.LastModified.IsZero() || source.LastModified.Before(base.LastModified) {
			base.LastModified = source.LastModified
			updated = true
		}
	}

	return updated
}
