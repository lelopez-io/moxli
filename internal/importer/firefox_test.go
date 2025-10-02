package importer

import (
	"strings"
	"testing"
)

func TestFirefoxImporter_Parse(t *testing.T) {
	htmlData := `<!DOCTYPE NETSCAPE-Bookmark-file-1>
<H1>Bookmarks Menu</H1>
<DL><p>
	<DT><H3 ADD_DATE="1736021373">Folder</H3>
	<DL><p>
		<DT><A HREF="https://example.com" ADD_DATE="1581232315" LAST_MODIFIED="1698071705">Example Site</A>
	</DL><p>
</DL><p>`

	importer := &FirefoxImporter{}
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

	if b.Source != "firefox" {
		t.Errorf("Source = %v, want firefox", b.Source)
	}

	// Check timestamps were parsed
	if b.DateAdded.IsZero() {
		t.Error("DateAdded should not be zero")
	}

	if b.LastModified.IsZero() {
		t.Error("LastModified should not be zero")
	}

	// NormalizedURL should be set
	if b.NormalizedURL == "" {
		t.Error("NormalizedURL should be set")
	}
}

func TestFirefoxImporter_Detect(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "valid firefox html with ADD_DATE",
			input: `<!DOCTYPE NETSCAPE-Bookmark-file-1><A HREF="https://example.com" ADD_DATE="123">Test</A>`,
			want:  true,
		},
		{
			name:  "netscape comment marker",
			input: `<!-- DOCTYPE NETSCAPE-Bookmark-file-1 --><A HREF="https://example.com">Test</A>`,
			want:  true,
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
			importer := &FirefoxImporter{}
			got := importer.Detect(strings.NewReader(tt.input))
			if got != tt.want {
				t.Errorf("Detect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFirefoxImporter_Source(t *testing.T) {
	importer := &FirefoxImporter{}
	if importer.Source() != "firefox" {
		t.Errorf("Source() = %v, want firefox", importer.Source())
	}
}
