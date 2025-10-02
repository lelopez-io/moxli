package importer

import (
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lelopez-io/moxli/internal/bookmark"
	"golang.org/x/net/html"
)

// SafariImporter handles Safari Netscape Bookmark HTML format
type SafariImporter struct{}

// Parse reads Safari HTML and returns a collection with URLs only (no timestamps)
func (s *SafariImporter) Parse(r io.Reader) (*bookmark.Collection, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	collection := bookmark.NewCollection()
	collection.Metadata.Source = "safari"
	collection.Metadata.ImportedAt = time.Now()

	// Parse the HTML tree
	s.parseNode(doc, collection)

	collection.UpdateMetadata()
	return collection, nil
}

// parseNode recursively parses HTML nodes
func (s *SafariImporter) parseNode(n *html.Node, collection *bookmark.Collection) {
	if n.Type == html.ElementNode && n.Data == "a" {
		// Extract bookmark data from <A> tag
		b := s.extractBookmark(n)
		if b != nil {
			collection.Add(b)
		}
	}

	// Recursively process children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		s.parseNode(c, collection)
	}
}

// extractBookmark extracts bookmark data from an <A> tag
func (s *SafariImporter) extractBookmark(n *html.Node) *bookmark.Bookmark {
	var href string

	// Extract HREF attribute
	for _, attr := range n.Attr {
		if strings.ToUpper(attr.Key) == "HREF" {
			href = attr.Val
			break
		}
	}

	// Skip if no URL
	if href == "" {
		return nil
	}

	// Create minimal bookmark with URL only (Safari has no timestamps)
	b := &bookmark.Bookmark{
		ID:         uuid.New().String(),
		URL:        href,
		Source:     "safari",
		ImportedAt: time.Now(),
	}

	// Normalize URL for matching
	if err := bookmark.NormalizeBookmarkURL(b); err != nil {
		return nil
	}

	return b
}

// Detect checks if the content is Safari/Netscape Bookmark format without timestamps
func (s *SafariImporter) Detect(r io.Reader) bool {
	doc, err := html.Parse(r)
	if err != nil {
		return false
	}

	// Look for Netscape Bookmark format markers but WITHOUT ADD_DATE
	// (Firefox has ADD_DATE, Safari doesn't)
	hasNetscapeMarker := s.hasNetscapeMarker(doc)
	hasAddDate := s.hasAddDate(doc)

	// Safari: has Netscape marker but no ADD_DATE attributes
	return hasNetscapeMarker && !hasAddDate
}

// hasNetscapeMarker checks for Netscape Bookmark file markers
func (s *SafariImporter) hasNetscapeMarker(n *html.Node) bool {
	if n.Type == html.CommentNode && strings.Contains(n.Data, "DOCTYPE NETSCAPE-Bookmark-file") {
		return true
	}
	if n.Type == html.ElementNode && n.Data == "h1" {
		// Check for "Bookmarks" title (Safari-specific)
		if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
			text := strings.ToLower(strings.TrimSpace(n.FirstChild.Data))
			if text == "bookmarks" {
				return true
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if s.hasNetscapeMarker(c) {
			return true
		}
	}

	return false
}

// hasAddDate checks for ADD_DATE attribute (Firefox-specific)
func (s *SafariImporter) hasAddDate(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if strings.ToUpper(attr.Key) == "ADD_DATE" {
				return true
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if s.hasAddDate(c) {
			return true
		}
	}

	return false
}

// Source returns the source identifier
func (s *SafariImporter) Source() string {
	return "safari"
}
