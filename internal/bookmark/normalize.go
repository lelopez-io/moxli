package bookmark

import (
	"net/url"
	"strings"
)

// NormalizeURL converts a URL to its canonical form for deduplication.
// This ensures that variations like trailing slashes, protocol differences,
// and query parameter ordering don't cause duplicate entries.
func NormalizeURL(rawURL string) (string, error) {
	// Parse the URL
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// Normalize scheme (default to https if missing)
	if u.Scheme == "" {
		u.Scheme = "https"
	}
	u.Scheme = strings.ToLower(u.Scheme)

	// Lowercase host
	u.Host = strings.ToLower(u.Host)

	// Remove trailing slash from path (unless it's the root path)
	if u.Path != "/" {
		u.Path = strings.TrimSuffix(u.Path, "/")
	}

	// Sort query parameters for consistency
	q := u.Query()
	u.RawQuery = q.Encode()

	// Remove fragment (hash)
	u.Fragment = ""

	return u.String(), nil
}

// NormalizeBookmarkURL normalizes the URL field and sets NormalizedURL
func NormalizeBookmarkURL(b *Bookmark) error {
	normalized, err := NormalizeURL(b.URL)
	if err != nil {
		return err
	}
	b.NormalizedURL = normalized
	return nil
}
