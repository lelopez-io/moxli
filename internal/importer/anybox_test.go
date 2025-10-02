package importer

import (
	"strings"
	"testing"
)

func TestAnyboxImporter_Parse(t *testing.T) {
	jsonData := `[
		{
			"url": "https://example.com",
			"title": "Example Site",
			"description": "An example",
			"tags": [["Security", "Auth"]],
			"folder": ["Work", "Projects"],
			"isStarred": true,
			"keyword": "ex",
			"comment": "Test comment",
			"dateAdded": "2025-01-01T00:00:00Z"
		}
	]`

	importer := &AnyboxImporter{}
	collection, err := importer.Parse(strings.NewReader(jsonData))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(collection.Bookmarks) != 1 {
		t.Fatalf("len(Bookmarks) = %v, want 1", len(collection.Bookmarks))
	}

	b := collection.Bookmarks[0]
	if b.URL != "https://example.com" {
		t.Errorf("URL = %v, want https://example.com", b.URL)
	}

	if b.Title != "Example Site" {
		t.Errorf("Title = %v, want Example Site", b.Title)
	}

	if b.Source != "anybox" {
		t.Errorf("Source = %v, want anybox", b.Source)
	}

	// Check tags were normalized
	if len(b.Tags) != 1 || len(b.Tags[0]) != 2 {
		t.Fatalf("Tags structure incorrect")
	}
	if b.Tags[0][0] != "security" || b.Tags[0][1] != "auth" {
		t.Errorf("Tags = %v, want [[security auth]]", b.Tags)
	}

	if !b.IsStarred {
		t.Error("IsStarred should be true")
	}
}

func TestAnyboxImporter_Detect(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "valid anybox json",
			input: `[{"url":"https://example.com","isStarred":false,"keyword":"","tags":[]}]`,
			want:  true,
		},
		{
			name:  "not json",
			input: `not json`,
			want:  false,
		},
		{
			name:  "empty array",
			input: `[]`,
			want:  false,
		},
		{
			name:  "generic json without anybox fields",
			input: `[{"url":"https://example.com","title":"Test"}]`,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			importer := &AnyboxImporter{}
			got := importer.Detect(strings.NewReader(tt.input))
			if got != tt.want {
				t.Errorf("Detect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnyboxImporter_Source(t *testing.T) {
	importer := &AnyboxImporter{}
	if importer.Source() != "anybox" {
		t.Errorf("Source() = %v, want anybox", importer.Source())
	}
}
