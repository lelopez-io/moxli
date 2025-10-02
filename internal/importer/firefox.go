package importer

import (
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/lelopez-io/moxli/internal/bookmark"
	"github.com/google/uuid"
	"golang.org/x/net/html"
)

// FirefoxImporter handles Firefox Netscape Bookmark HTML format
type FirefoxImporter struct{}

// Parse reads Firefox HTML and returns a collection with URLs and timestamps only
func (f *FirefoxImporter) Parse(r io.Reader) (*bookmark.Collection, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	collection := bookmark.NewCollection()
	collection.Metadata.Source = "firefox"
	collection.Metadata.ImportedAt = time.Now()

	// Parse the HTML tree
	f.parseNode(doc, collection)

	collection.UpdateMetadata()
	return collection, nil
}

// parseNode recursively parses HTML nodes
func (f *FirefoxImporter) parseNode(n *html.Node, collection *bookmark.Collection) {
	if n.Type == html.ElementNode && n.Data == "a" {
		// Extract bookmark data from <A> tag
		b := f.extractBookmark(n)
		if b != nil {
			collection.Add(b)
		}
	}

	// Recursively process children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		f.parseNode(c, collection)
	}
}

// extractBookmark extracts bookmark data from an <A> tag
func (f *FirefoxImporter) extractBookmark(n *html.Node) *bookmark.Bookmark {
	var href string
	var addDate, lastModified int64

	// Extract attributes
	for _, attr := range n.Attr {
		switch strings.ToUpper(attr.Key) {
		case "HREF":
			href = attr.Val
		case "ADD_DATE":
			addDate, _ = strconv.ParseInt(attr.Val, 10, 64)
		case "LAST_MODIFIED":
			lastModified, _ = strconv.ParseInt(attr.Val, 10, 64)
		}
	}

	// Skip if no URL
	if href == "" {
		return nil
	}

	// Create minimal bookmark with URL and timestamps only
	b := &bookmark.Bookmark{
		ID:     uuid.New().String(),
		URL:    href,
		Source: "firefox",
		ImportedAt: time.Now(),
	}

	// Convert Unix timestamps to time.Time
	if addDate > 0 {
		b.DateAdded = time.Unix(addDate, 0)
	}
	if lastModified > 0 {
		b.LastModified = time.Unix(lastModified, 0)
	}

	// Normalize URL for matching
	if err := bookmark.NormalizeBookmarkURL(b); err != nil {
		return nil
	}

	return b
}

// Detect checks if the content is Firefox Netscape Bookmark format
func (f *FirefoxImporter) Detect(r io.Reader) bool {
	doc, err := html.Parse(r)
	if err != nil {
		return false
	}

	// Look for Netscape Bookmark format markers
	return f.hasNetscapeMarkers(doc)
}

// hasNetscapeMarkers checks for Netscape Bookmark file markers
func (f *FirefoxImporter) hasNetscapeMarkers(n *html.Node) bool {
	if n.Type == html.CommentNode && strings.Contains(n.Data, "DOCTYPE NETSCAPE-Bookmark-file") {
		return true
	}
	if n.Type == html.ElementNode && n.Data == "a" {
		// Check for ADD_DATE attribute (Firefox-specific)
		for _, attr := range n.Attr {
			if strings.ToUpper(attr.Key) == "ADD_DATE" {
				return true
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if f.hasNetscapeMarkers(c) {
			return true
		}
	}

	return false
}

// Source returns the source identifier
func (f *FirefoxImporter) Source() string {
	return "firefox"
}
