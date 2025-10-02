package exporter

import (
	"fmt"
	"strings"

	"github.com/lelopez-io/moxli/internal/bookmark"
)

// ValidationError represents an error found during bookmark validation
type ValidationError struct {
	BookmarkID string
	Field      string
	Message    string
}

func (e ValidationError) Error() string {
	if e.BookmarkID != "" {
		return fmt.Sprintf("bookmark %s: %s - %s", e.BookmarkID, e.Field, e.Message)
	}
	return fmt.Sprintf("%s - %s", e.Field, e.Message)
}

// ValidationResult contains the results of validating a collection
type ValidationResult struct {
	Valid  bool
	Errors []ValidationError
}

// ValidateCollection checks if a collection is ready for export
func ValidateCollection(c *bookmark.Collection) ValidationResult {
	result := ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	if c == nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "collection",
			Message: "collection is nil",
		})
		return result
	}

	// Validate each bookmark
	for _, b := range c.Bookmarks {
		errs := validateBookmark(b)
		if len(errs) > 0 {
			result.Valid = false
			result.Errors = append(result.Errors, errs...)
		}
	}

	return result
}

// validateBookmark checks if a single bookmark is valid
func validateBookmark(b *bookmark.Bookmark) []ValidationError {
	var errors []ValidationError

	// URL is required
	if b.URL == "" {
		errors = append(errors, ValidationError{
			BookmarkID: b.ID,
			Field:      "url",
			Message:    "URL is required",
		})
	}

	// URL should be properly formed
	if b.URL != "" && !strings.HasPrefix(b.URL, "http://") && !strings.HasPrefix(b.URL, "https://") {
		errors = append(errors, ValidationError{
			BookmarkID: b.ID,
			Field:      "url",
			Message:    "URL must start with http:// or https://",
		})
	}

	// NormalizedURL should be set (indicates URL normalization was performed)
	if b.NormalizedURL == "" && b.URL != "" {
		errors = append(errors, ValidationError{
			BookmarkID: b.ID,
			Field:      "normalizedURL",
			Message:    "normalized URL is missing (URL normalization required)",
		})
	}

	// Tags should not have empty hierarchies
	for i, tagGroup := range b.Tags {
		if len(tagGroup) == 0 {
			errors = append(errors, ValidationError{
				BookmarkID: b.ID,
				Field:      fmt.Sprintf("tags[%d]", i),
				Message:    "tag group is empty",
			})
		}
		// Check for empty tag values
		for j, tag := range tagGroup {
			if strings.TrimSpace(tag) == "" {
				errors = append(errors, ValidationError{
					BookmarkID: b.ID,
					Field:      fmt.Sprintf("tags[%d][%d]", i, j),
					Message:    "tag value is empty",
				})
			}
		}
	}

	// Folder path should not have empty segments
	for i, segment := range b.Folder {
		if strings.TrimSpace(segment) == "" {
			errors = append(errors, ValidationError{
				BookmarkID: b.ID,
				Field:      fmt.Sprintf("folder[%d]", i),
				Message:    "folder segment is empty",
			})
		}
	}

	return errors
}
