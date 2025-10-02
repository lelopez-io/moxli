package importer

import (
	"strings"
	"testing"
)

func TestSafariImporter_Parse(t *testing.T) {
	htmlData := `<!DOCTYPE NETSCAPE-Bookmark-file-1>
<H1>Bookmarks</H1>
<DL><p>
	<DT><H3>Folder</H3>
	<DL><p>
		<DT><A HREF="https://example.com">Example Site</A>
	</DL><p>
</DL><p>`

	importer := &SafariImporter{}
	collection, err := importer.Parse(strings.NewReader(htmlData))
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

	if b.Source != "safari" {
		t.Errorf("Source = %v, want safari", b.Source)
	}

	// Safari has no timestamps - should be zero
	if !b.DateAdded.IsZero() {
		t.Error("DateAdded should be zero for Safari imports")
	}

	if !b.LastModified.IsZero() {
		t.Error("LastModified should be zero for Safari imports")
	}

	// NormalizedURL should be set
	if b.NormalizedURL == "" {
		t.Error("NormalizedURL should be set")
	}
}

func TestSafariImporter_Detect(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "valid safari html without ADD_DATE",
			input: `<!DOCTYPE NETSCAPE-Bookmark-file-1><H1>Bookmarks</H1><A HREF="https://example.com">Test</A>`,
			want:  true,
		},
		{
			name:  "netscape comment marker without ADD_DATE",
			input: `<!-- DOCTYPE NETSCAPE-Bookmark-file-1 --><A HREF="https://example.com">Test</A>`,
			want:  true,
		},
		{
			name:  "firefox html with ADD_DATE (should be false)",
			input: `<!DOCTYPE NETSCAPE-Bookmark-file-1><A HREF="https://example.com" ADD_DATE="123">Test</A>`,
			want:  false,
		},
		{
			name:  "generic html",
			input: `<html><body><a href="https://example.com">Test</a></body></html>`,
			want:  false,
		},
		{
			name:  "not html",
			input: `not html`,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			importer := &SafariImporter{}
			got := importer.Detect(strings.NewReader(tt.input))
			if got != tt.want {
				t.Errorf("Detect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafariImporter_Source(t *testing.T) {
	importer := &SafariImporter{}
	if importer.Source() != "safari" {
		t.Errorf("Source() = %v, want safari", importer.Source())
	}
}
