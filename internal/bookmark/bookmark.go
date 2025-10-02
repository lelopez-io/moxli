package bookmark

import "time"

// Bookmark represents a single bookmark with all its metadata.
// The structure matches Anybox JSON export format for compatibility.
type Bookmark struct {
	// Core fields
	ID            string    `json:"id"`            // Internal UUID
	URL           string    `json:"url"`           // Original URL
	NormalizedURL string    `json:"-"`             // Normalized for deduplication (not exported)
	Title         string    `json:"title"`         // Page title
	Description   string    `json:"description"`   // Auto-extracted meta description

	// Organization
	Tags   [][]string `json:"tags"`   // Hierarchical: [["Security", "User Auth"], ["Infrastructure"]]
	Folder []string   `json:"folder"` // Path array: ["BookmarksBar", "Research", "Wagmo"]

	// User annotations
	Comment   string `json:"comment"`   // User's personal note
	Keyword   string `json:"keyword"`   // Shortcut/alias for quick access
	IsStarred bool   `json:"isStarred"` // Favorite flag

	// Timestamps
	DateAdded    time.Time `json:"dateAdded"`              // ISO8601 format
	LastModified time.Time `json:"lastModified,omitempty"` // Optional

	// Content (Anybox feature - may be empty)
	Article string `json:"article,omitempty"` // Full saved article text

	// Metadata for moxli
	Source     string    `json:"source,omitempty"`     // "anybox", "safari", "firefox"
	ImportedAt time.Time `json:"importedAt,omitempty"` // When imported into moxli
}

// Clone creates a deep copy of the bookmark
func (b *Bookmark) Clone() *Bookmark {
	clone := *b

	// Deep copy slices
	if b.Tags != nil {
		clone.Tags = make([][]string, len(b.Tags))
		for i, tag := range b.Tags {
			clone.Tags[i] = make([]string, len(tag))
			copy(clone.Tags[i], tag)
		}
	}

	if b.Folder != nil {
		clone.Folder = make([]string, len(b.Folder))
		copy(clone.Folder, b.Folder)
	}

	return &clone
}
