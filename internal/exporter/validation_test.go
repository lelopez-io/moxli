package exporter

import (
	"testing"
	"time"

	"github.com/lelopez-io/moxli/internal/bookmark"
)

func TestValidateCollection_ValidBookmark(t *testing.T) {
	collection := bookmark.NewCollection()
	collection.Add(&bookmark.Bookmark{
		ID:            "test-id",
		URL:           "https://example.com",
		NormalizedURL: "https://example.com",
		Title:         "Example",
		Tags:          [][]string{{"security", "auth"}},
		Folder:        []string{"Work"},
		DateAdded:     time.Now(),
	})

	result := ValidateCollection(collection)
	if !result.Valid {
		t.Errorf("Valid bookmark should pass validation, got errors: %v", result.Errors)
	}

	if len(result.Errors) != 0 {
		t.Errorf("len(Errors) = %v, want 0", len(result.Errors))
	}
}

func TestValidateCollection_MissingURL(t *testing.T) {
	collection := bookmark.NewCollection()
	collection.Add(&bookmark.Bookmark{
		ID:    "test-id",
		URL:   "", // Missing
		Title: "Example",
	})

	result := ValidateCollection(collection)
	if result.Valid {
		t.Error("Bookmark with missing URL should fail validation")
	}

	if len(result.Errors) == 0 {
		t.Fatal("Expected validation errors")
	}

	// Should have error about missing URL
	foundURLError := false
	for _, err := range result.Errors {
		if err.Field == "url" && err.BookmarkID == "test-id" {
			foundURLError = true
			break
		}
	}
	if !foundURLError {
		t.Error("Expected URL validation error")
	}
}

func TestValidateCollection_InvalidURLScheme(t *testing.T) {
	collection := bookmark.NewCollection()
	collection.Add(&bookmark.Bookmark{
		ID:            "test-id",
		URL:           "ftp://example.com", // Invalid scheme
		NormalizedURL: "ftp://example.com",
	})

	result := ValidateCollection(collection)
	if result.Valid {
		t.Error("Bookmark with invalid URL scheme should fail validation")
	}

	foundSchemeError := false
	for _, err := range result.Errors {
		if err.Field == "url" && err.Message == "URL must start with http:// or https://" {
			foundSchemeError = true
			break
		}
	}
	if !foundSchemeError {
		t.Error("Expected URL scheme validation error")
	}
}

func TestValidateCollection_MissingNormalizedURL(t *testing.T) {
	collection := bookmark.NewCollection()
	collection.Add(&bookmark.Bookmark{
		ID:            "test-id",
		URL:           "https://example.com",
		NormalizedURL: "", // Missing
	})

	result := ValidateCollection(collection)
	if result.Valid {
		t.Error("Bookmark with missing normalized URL should fail validation")
	}

	foundNormError := false
	for _, err := range result.Errors {
		if err.Field == "normalizedURL" {
			foundNormError = true
			break
		}
	}
	if !foundNormError {
		t.Error("Expected normalizedURL validation error")
	}
}

func TestValidateCollection_EmptyTagGroup(t *testing.T) {
	collection := bookmark.NewCollection()
	collection.Add(&bookmark.Bookmark{
		ID:            "test-id",
		URL:           "https://example.com",
		NormalizedURL: "https://example.com",
		Tags:          [][]string{{}}, // Empty tag group
	})

	result := ValidateCollection(collection)
	if result.Valid {
		t.Error("Bookmark with empty tag group should fail validation")
	}

	foundTagError := false
	for _, err := range result.Errors {
		if err.Field == "tags[0]" {
			foundTagError = true
			break
		}
	}
	if !foundTagError {
		t.Error("Expected tag group validation error")
	}
}

func TestValidateCollection_EmptyTagValue(t *testing.T) {
	collection := bookmark.NewCollection()
	collection.Add(&bookmark.Bookmark{
		ID:            "test-id",
		URL:           "https://example.com",
		NormalizedURL: "https://example.com",
		Tags:          [][]string{{"security", ""}}, // Empty tag value
	})

	result := ValidateCollection(collection)
	if result.Valid {
		t.Error("Bookmark with empty tag value should fail validation")
	}

	foundTagError := false
	for _, err := range result.Errors {
		if err.Field == "tags[0][1]" {
			foundTagError = true
			break
		}
	}
	if !foundTagError {
		t.Error("Expected tag value validation error")
	}
}

func TestValidateCollection_EmptyFolderSegment(t *testing.T) {
	collection := bookmark.NewCollection()
	collection.Add(&bookmark.Bookmark{
		ID:            "test-id",
		URL:           "https://example.com",
		NormalizedURL: "https://example.com",
		Folder:        []string{"Work", ""}, // Empty segment
	})

	result := ValidateCollection(collection)
	if result.Valid {
		t.Error("Bookmark with empty folder segment should fail validation")
	}

	foundFolderError := false
	for _, err := range result.Errors {
		if err.Field == "folder[1]" {
			foundFolderError = true
			break
		}
	}
	if !foundFolderError {
		t.Error("Expected folder segment validation error")
	}
}

func TestValidateCollection_NilCollection(t *testing.T) {
	result := ValidateCollection(nil)
	if result.Valid {
		t.Error("Nil collection should fail validation")
	}

	if len(result.Errors) != 1 {
		t.Fatalf("len(Errors) = %v, want 1", len(result.Errors))
	}

	if result.Errors[0].Field != "collection" {
		t.Errorf("Error field = %v, want collection", result.Errors[0].Field)
	}
}

func TestValidateCollection_EmptyCollection(t *testing.T) {
	collection := bookmark.NewCollection()

	result := ValidateCollection(collection)
	if !result.Valid {
		t.Errorf("Empty collection should be valid, got errors: %v", result.Errors)
	}

	if len(result.Errors) != 0 {
		t.Errorf("len(Errors) = %v, want 0 for empty collection", len(result.Errors))
	}
}

func TestValidateCollection_MultipleErrors(t *testing.T) {
	collection := bookmark.NewCollection()
	collection.Add(&bookmark.Bookmark{
		ID:            "test-id",
		URL:           "", // Missing URL
		NormalizedURL: "", // Missing normalized URL
		Tags:          [][]string{{}}, // Empty tag group
		Folder:        []string{""}, // Empty folder segment
	})

	result := ValidateCollection(collection)
	if result.Valid {
		t.Error("Invalid bookmark should fail validation")
	}

	// Should have multiple errors
	if len(result.Errors) < 3 {
		t.Errorf("Expected at least 3 errors, got %d", len(result.Errors))
	}
}

func TestValidationError_Error(t *testing.T) {
	err := ValidationError{
		BookmarkID: "test-id",
		Field:      "url",
		Message:    "URL is required",
	}

	expected := "bookmark test-id: url - URL is required"
	if err.Error() != expected {
		t.Errorf("Error() = %v, want %v", err.Error(), expected)
	}
}

func TestValidationError_Error_NoBookmarkID(t *testing.T) {
	err := ValidationError{
		Field:   "collection",
		Message: "collection is nil",
	}

	expected := "collection - collection is nil"
	if err.Error() != expected {
		t.Errorf("Error() = %v, want %v", err.Error(), expected)
	}
}
