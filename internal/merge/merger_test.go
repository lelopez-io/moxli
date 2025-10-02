package merge

import (
	"testing"
	"time"

	"github.com/lelopez-io/moxli/internal/bookmark"
)

func TestMerger_Merge_TimestampReconciliation(t *testing.T) {
	// Base collection with recent dates (from Safari import)
	base := bookmark.NewCollection()
	base.Add(&bookmark.Bookmark{
		URL:           "https://example.com",
		NormalizedURL: "https://example.com",
		Title:         "Example",
		DateAdded:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), // Recent date
	})

	// Source collection with older authentic dates (from Firefox)
	source := bookmark.NewCollection()
	source.Add(&bookmark.Bookmark{
		URL:           "https://example.com",
		NormalizedURL: "https://example.com",
		DateAdded:     time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), // Older authentic date
	})

	merger := New(base, source)
	result, err := merger.Merge()
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	if len(result.Bookmarks) != 1 {
		t.Fatalf("len(Bookmarks) = %v, want 1", len(result.Bookmarks))
	}

	// Should prefer older date from source
	b := result.Bookmarks[0]
	want := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	if !b.DateAdded.Equal(want) {
		t.Errorf("DateAdded = %v, want %v", b.DateAdded, want)
	}

	// Other fields should remain from base
	if b.Title != "Example" {
		t.Errorf("Title = %v, want Example (from base)", b.Title)
	}
}

func TestMerger_Merge_IgnoresNewBookmarks(t *testing.T) {
	// Base with one bookmark
	base := bookmark.NewCollection()
	base.Add(&bookmark.Bookmark{
		URL:           "https://example.com",
		NormalizedURL: "https://example.com",
		Title:         "Example",
	})

	// Source with additional bookmark not in base
	source := bookmark.NewCollection()
	source.Add(&bookmark.Bookmark{
		URL:           "https://deleted.com", // Not in base - intentionally deleted
		NormalizedURL: "https://deleted.com",
		Title:         "Deleted",
	})

	merger := New(base, source)
	result, err := merger.Merge()
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// Should not add new bookmark from source
	if len(result.Bookmarks) != 1 {
		t.Errorf("len(Bookmarks) = %v, want 1 (should not add deleted bookmark)", len(result.Bookmarks))
	}

	if result.Bookmarks[0].URL != "https://example.com" {
		t.Error("Only base bookmark should be present")
	}
}

func TestMerger_Merge_ZeroTimestamps(t *testing.T) {
	// Base with zero timestamp
	base := bookmark.NewCollection()
	base.Add(&bookmark.Bookmark{
		URL:           "https://example.com",
		NormalizedURL: "https://example.com",
		DateAdded:     time.Time{}, // Zero time
	})

	// Source with valid timestamp
	source := bookmark.NewCollection()
	source.Add(&bookmark.Bookmark{
		URL:           "https://example.com",
		NormalizedURL: "https://example.com",
		DateAdded:     time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	})

	merger := New(base, source)
	result, err := merger.Merge()
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// Should use non-zero timestamp from source
	b := result.Bookmarks[0]
	want := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	if !b.DateAdded.Equal(want) {
		t.Errorf("DateAdded = %v, want %v (should use non-zero timestamp)", b.DateAdded, want)
	}
}

func TestMerger_Merge_LastModified(t *testing.T) {
	base := bookmark.NewCollection()
	base.Add(&bookmark.Bookmark{
		URL:           "https://example.com",
		NormalizedURL: "https://example.com",
		LastModified:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	})

	source := bookmark.NewCollection()
	source.Add(&bookmark.Bookmark{
		URL:           "https://example.com",
		NormalizedURL: "https://example.com",
		LastModified:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), // Older
	})

	merger := New(base, source)
	result, err := merger.Merge()
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// Should prefer older LastModified too
	b := result.Bookmarks[0]
	want := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	if !b.LastModified.Equal(want) {
		t.Errorf("LastModified = %v, want %v", b.LastModified, want)
	}
}

func TestMerger_Merge_MultipleSources(t *testing.T) {
	base := bookmark.NewCollection()
	base.Add(&bookmark.Bookmark{
		URL:           "https://example.com",
		NormalizedURL: "https://example.com",
		DateAdded:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	})

	source1 := bookmark.NewCollection()
	source1.Add(&bookmark.Bookmark{
		URL:           "https://example.com",
		NormalizedURL: "https://example.com",
		DateAdded:     time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
	})

	source2 := bookmark.NewCollection()
	source2.Add(&bookmark.Bookmark{
		URL:           "https://example.com",
		NormalizedURL: "https://example.com",
		DateAdded:     time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), // Oldest
	})

	merger := New(base, source1, source2)
	result, err := merger.Merge()
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// Should use oldest date across all sources
	b := result.Bookmarks[0]
	want := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	if !b.DateAdded.Equal(want) {
		t.Errorf("DateAdded = %v, want %v (oldest across all sources)", b.DateAdded, want)
	}
}

func TestMerger_Merge_NoModificationToBase(t *testing.T) {
	base := bookmark.NewCollection()
	base.Add(&bookmark.Bookmark{
		URL:           "https://example.com",
		NormalizedURL: "https://example.com",
		Title:         "Example",
		Tags:          [][]string{{"security", "auth"}},
		DateAdded:     time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	})

	source := bookmark.NewCollection()
	source.Add(&bookmark.Bookmark{
		URL:           "https://example.com",
		NormalizedURL: "https://example.com",
		DateAdded:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), // Newer - should be ignored
	})

	merger := New(base, source)
	result, err := merger.Merge()
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// Nothing should change since source has newer date
	b := result.Bookmarks[0]
	want := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	if !b.DateAdded.Equal(want) {
		t.Errorf("DateAdded = %v, want %v (should keep base date)", b.DateAdded, want)
	}
}
