package exporter

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/lelopez-io/moxli/internal/bookmark"
)

func TestAnyboxExporter_Export(t *testing.T) {
	collection := bookmark.NewCollection()
	collection.Add(&bookmark.Bookmark{
		ID:          "test-id",
		URL:         "https://example.com",
		Title:       "Example",
		Description: "Test description",
		Tags:        [][]string{{"security", "auth"}},
		Folder:      []string{"Work"},
		IsStarred:   true,
		Keyword:     "ex",
		Comment:     "Test comment",
		DateAdded:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	})

	exporter := &AnyboxExporter{Indent: false}
	var buf bytes.Buffer

	err := exporter.Export(&buf, collection)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	// Parse back to verify structure
	var bookmarks []*bookmark.Bookmark
	if err := json.Unmarshal(buf.Bytes(), &bookmarks); err != nil {
		t.Fatalf("Failed to parse exported JSON: %v", err)
	}

	if len(bookmarks) != 1 {
		t.Fatalf("len(bookmarks) = %v, want 1", len(bookmarks))
	}

	b := bookmarks[0]
	if b.URL != "https://example.com" {
		t.Errorf("URL = %v, want https://example.com", b.URL)
	}

	if b.Title != "Example" {
		t.Errorf("Title = %v, want Example", b.Title)
	}

	if !b.IsStarred {
		t.Error("IsStarred should be true")
	}

	if len(b.Tags) != 1 {
		t.Fatalf("len(Tags) = %v, want 1", len(b.Tags))
	}
}

func TestAnyboxExporter_Export_WithIndent(t *testing.T) {
	collection := bookmark.NewCollection()
	collection.Add(&bookmark.Bookmark{
		URL:   "https://example.com",
		Title: "Example",
	})

	exporter := &AnyboxExporter{Indent: true}
	var buf bytes.Buffer

	err := exporter.Export(&buf, collection)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	// Check that output contains newlines (indicating indentation)
	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("\n")) {
		t.Error("Expected indented output to contain newlines")
	}
}

func TestAnyboxExporter_Export_EmptyCollection(t *testing.T) {
	collection := bookmark.NewCollection()

	exporter := &AnyboxExporter{Indent: false}
	var buf bytes.Buffer

	err := exporter.Export(&buf, collection)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	// Should export empty array
	var bookmarks []*bookmark.Bookmark
	if err := json.Unmarshal(buf.Bytes(), &bookmarks); err != nil {
		t.Fatalf("Failed to parse exported JSON: %v", err)
	}

	if len(bookmarks) != 0 {
		t.Errorf("len(bookmarks) = %v, want 0", len(bookmarks))
	}
}
