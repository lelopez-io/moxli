package bookmark

import "time"

// Collection represents a collection of bookmarks with metadata
type Collection struct {
	Version   string      `json:"version"`   // Collection format version
	Updated   time.Time   `json:"updated"`   // Last update timestamp
	Bookmarks []*Bookmark `json:"bookmarks"` // Array for Anybox compatibility
	Metadata  Metadata    `json:"metadata"`  // Collection-level metadata

	// Internal index for deduplication and fast lookups (not exported)
	urlIndex map[string]*Bookmark `json:"-"` // normalizedURL → first bookmark with that URL
}

// Metadata contains collection-level information
type Metadata struct {
	Source      string    `json:"source,omitempty"`      // Origin of this collection
	ImportedAt  time.Time `json:"importedAt,omitempty"`  // When imported
	TotalCount  int       `json:"totalCount"`            // Total bookmarks
	TagCount    int       `json:"tagCount,omitempty"`    // Unique tags
	FolderCount int       `json:"folderCount,omitempty"` // Unique folders
}

// NewCollection creates a new empty collection
func NewCollection() *Collection {
	return &Collection{
		Version:   "1.0",
		Updated:   time.Now(),
		Bookmarks: make([]*Bookmark, 0),
		Metadata:  Metadata{},
		urlIndex:  make(map[string]*Bookmark),
	}
}

// Clone creates a deep copy of the collection
func (c *Collection) Clone() *Collection {
	clone := &Collection{
		Version:   c.Version,
		Updated:   c.Updated,
		Bookmarks: make([]*Bookmark, len(c.Bookmarks)),
		Metadata:  c.Metadata,
		urlIndex:  make(map[string]*Bookmark),
	}

	for i, b := range c.Bookmarks {
		clone.Bookmarks[i] = b.Clone()
	}

	clone.buildURLIndex()
	return clone
}

// buildURLIndex constructs the internal URL index for fast lookups
func (c *Collection) buildURLIndex() {
	c.urlIndex = make(map[string]*Bookmark)
	for _, b := range c.Bookmarks {
		if b.NormalizedURL != "" {
			// Map to first occurrence (for duplicate detection)
			if _, exists := c.urlIndex[b.NormalizedURL]; !exists {
				c.urlIndex[b.NormalizedURL] = b
			}
		}
	}
}

// Add adds a bookmark to the collection and updates the index
func (c *Collection) Add(b *Bookmark) {
	c.Bookmarks = append(c.Bookmarks, b)
	c.Updated = time.Now()

	// Update URL index if normalized URL is set
	if b.NormalizedURL != "" {
		if _, exists := c.urlIndex[b.NormalizedURL]; !exists {
			c.urlIndex[b.NormalizedURL] = b
		}
	}
}

// FindByURL looks up a bookmark by normalized URL
func (c *Collection) FindByURL(normalizedURL string) (*Bookmark, bool) {
	// Build index if not already built
	if c.urlIndex == nil {
		c.buildURLIndex()
	}

	b, exists := c.urlIndex[normalizedURL]
	return b, exists
}

// UpdateMetadata recalculates collection metadata
func (c *Collection) UpdateMetadata() {
	c.Metadata.TotalCount = len(c.Bookmarks)
	c.Updated = time.Now()

	// Count unique tags and folders
	tags := make(map[string]bool)
	folders := make(map[string]bool)

	for _, b := range c.Bookmarks {
		for _, tagHierarchy := range b.Tags {
			for _, tag := range tagHierarchy {
				tags[tag] = true
			}
		}
		for _, folder := range b.Folder {
			folders[folder] = true
		}
	}

	c.Metadata.TagCount = len(tags)
	c.Metadata.FolderCount = len(folders)
}
