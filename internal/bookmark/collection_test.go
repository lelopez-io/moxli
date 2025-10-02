package bookmark

import (
	"testing"
	"time"
)

func TestNewCollection(t *testing.T) {
	c := NewCollection()

	if c.Version != "1.0" {
		t.Errorf("Version = %v, want 1.0", c.Version)
	}

	if c.Bookmarks == nil {
		t.Error("Bookmarks should not be nil")
	}

	if c.urlIndex == nil {
		t.Error("urlIndex should not be nil")
	}
}

func TestCollection_Add(t *testing.T) {
	c := NewCollection()
	b := &Bookmark{
		URL:           "https://example.com",
		NormalizedURL: "https://example.com",
		Title:         "Example",
	}

	c.Add(b)

	if len(c.Bookmarks) != 1 {
		t.Errorf("len(Bookmarks) = %v, want 1", len(c.Bookmarks))
	}

	if c.Bookmarks[0] != b {
		t.Error("Added bookmark not found in collection")
	}
}

func TestCollection_FindByURL(t *testing.T) {
	c := NewCollection()
	b1 := &Bookmark{
		URL:           "https://example.com",
		NormalizedURL: "https://example.com",
		Title:         "Example",
	}
	b2 := &Bookmark{
		URL:           "https://test.com",
		NormalizedURL: "https://test.com",
		Title:         "Test",
	}

	c.Add(b1)
	c.Add(b2)

	// Find existing bookmark
	found, exists := c.FindByURL("https://example.com")
	if !exists {
		t.Error("Bookmark should exist")
	}
	if found != b1 {
		t.Error("Wrong bookmark returned")
	}

	// Try non-existent URL
	_, exists = c.FindByURL("https://notfound.com")
	if exists {
		t.Error("Non-existent bookmark should not be found")
	}
}

func TestCollection_UpdateMetadata(t *testing.T) {
	c := NewCollection()

	c.Add(&Bookmark{
		Tags:   [][]string{{"security", "auth"}, {"infrastructure"}},
		Folder: []string{"Work", "Projects"},
	})
	c.Add(&Bookmark{
		Tags:   [][]string{{"security", "oauth"}},
		Folder: []string{"Personal"},
	})

	c.UpdateMetadata()

	if c.Metadata.TotalCount != 2 {
		t.Errorf("TotalCount = %v, want 2", c.Metadata.TotalCount)
	}

	// Should have unique tags: security, auth, infrastructure, oauth = 4
	if c.Metadata.TagCount != 4 {
		t.Errorf("TagCount = %v, want 4", c.Metadata.TagCount)
	}

	// Should have unique folders: Work, Projects, Personal = 3
	if c.Metadata.FolderCount != 3 {
		t.Errorf("FolderCount = %v, want 3", c.Metadata.FolderCount)
	}
}

func TestCollection_Clone(t *testing.T) {
	original := NewCollection()
	original.Version = "1.0"
	original.Updated = time.Now()

	b := &Bookmark{
		URL:           "https://example.com",
		NormalizedURL: "https://example.com",
		Title:         "Example",
		Tags:          [][]string{{"security"}},
	}
	original.Add(b)

	clone := original.Clone()

	// Check deep copy
	if clone == original {
		t.Error("Clone should be a different object")
	}

	if clone.Version != original.Version {
		t.Error("Version not cloned correctly")
	}

	if len(clone.Bookmarks) != len(original.Bookmarks) {
		t.Error("Bookmarks not cloned correctly")
	}

	// Modify clone should not affect original
	clone.Bookmarks[0].Title = "Modified"
	if original.Bookmarks[0].Title == "Modified" {
		t.Error("Clone is not independent from original")
	}
}

func TestCollection_buildURLIndex(t *testing.T) {
	c := NewCollection()

	b1 := &Bookmark{
		URL:           "https://example.com",
		NormalizedURL: "https://example.com",
	}
	b2 := &Bookmark{
		URL:           "https://example.com/", // Same normalized URL
		NormalizedURL: "https://example.com",
	}

	c.Bookmarks = []*Bookmark{b1, b2}
	c.buildURLIndex()

	// Should map to first occurrence
	found, exists := c.urlIndex["https://example.com"]
	if !exists {
		t.Error("URL not in index")
	}
	if found != b1 {
		t.Error("Index should map to first occurrence")
	}
}
