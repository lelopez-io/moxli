package importer

import (
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lelopez-io/moxli/internal/bookmark"
	"golang.org/x/net/html"
)

// AnyboxHTMLImporter handles Anybox HTML export format
type AnyboxHTMLImporter struct{}

// Parse reads Anybox HTML and returns a collection
func (a *AnyboxHTMLImporter) Parse(r io.Reader) (*bookmark.Collection, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	collection := bookmark.NewCollection()
	collection.Metadata.Source = "anybox-html"
	collection.Metadata.ImportedAt = time.Now()

	// Parse the HTML tree
	a.parseNode(doc, collection)

	collection.UpdateMetadata()
	return collection, nil
}

// parseNode recursively parses HTML nodes
func (a *AnyboxHTMLImporter) parseNode(n *html.Node, collection *bookmark.Collection) {
	if n.Type == html.ElementNode && n.Data == "a" {
		// Extract bookmark data from <A> tag
		b := a.extractBookmark(n)
		if b != nil {
			collection.Add(b)
		}
	}

	// Recursively process children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		a.parseNode(c, collection)
	}
}

// extractBookmark extracts bookmark data from an <A> tag
func (a *AnyboxHTMLImporter) extractBookmark(n *html.Node) *bookmark.Bookmark {
	var href string
	var addDate int64
	var tags []string

	// Extract attributes
	for _, attr := range n.Attr {
		switch strings.ToUpper(attr.Key) {
		case "HREF":
			href = attr.Val
		case "ADD_DATE":
			addDate, _ = strconv.ParseInt(attr.Val, 10, 64)
		case "TAGS":
			// Split tags by comma
			if attr.Val != "" {
				tags = strings.Split(attr.Val, ",")
			}
		}
	}

	// Skip if no URL
	if href == "" {
		return nil
	}

	// Create bookmark with URL, timestamp, and tags
	b := &bookmark.Bookmark{
		ID:         uuid.New().String(),
		URL:        href,
		Source:     "anybox-html",
		ImportedAt: time.Now(),
	}

	// Add tags if present (convert to hierarchical format)
	if len(tags) > 0 {
		b.Tags = make([][]string, len(tags))
		for i, tag := range tags {
			b.Tags[i] = []string{strings.TrimSpace(tag)}
		}
	}

	// Convert Unix timestamp to time.Time
	if addDate > 0 {
		b.DateAdded = time.Unix(addDate, 0)
	}

	// Normalize URL for matching
	if err := bookmark.NormalizeBookmarkURL(b); err != nil {
		return nil
	}

	// Normalize tags
	bookmark.NormalizeTags(b)

	return b
}

// Detect checks if the content is Anybox HTML format
func (a *AnyboxHTMLImporter) Detect(r io.Reader) bool {
	doc, err := html.Parse(r)
	if err != nil {
		return false
	}

	// Anybox HTML has TAGS attributes but NO folder structure (H3 tags)
	// Firefox also supports TAGS but has nested H3 folder hierarchy
	hasTags := a.hasTAGSAttribute(doc)
	hasH3 := a.hasH3Tags(doc)

	// Anybox: has TAGS but no H3 folder structure
	return hasTags && !hasH3
}

// hasTAGSAttribute checks for TAGS attribute in anchor tags
func (a *AnyboxHTMLImporter) hasTAGSAttribute(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if strings.ToUpper(attr.Key) == "TAGS" {
				return true
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if a.hasTAGSAttribute(c) {
			return true
		}
	}

	return false
}

// hasH3Tags checks for H3 folder structure (Firefox-specific)
func (a *AnyboxHTMLImporter) hasH3Tags(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "h3" {
		return true
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if a.hasH3Tags(c) {
			return true
		}
	}

	return false
}

// Source returns the source identifier
func (a *AnyboxHTMLImporter) Source() string {
	return "anybox-html"
}
