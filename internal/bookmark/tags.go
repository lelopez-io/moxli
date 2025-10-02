package bookmark

import (
	"regexp"
	"strings"
	"unicode"
)

// NormalizeTag converts a tag to lower-kebab-case format.
// Examples:
//   - "User Auth" → "user-auth"
//   - "DevOps" → "dev-ops"
//   - "GraphQL API" → "graph-ql-api"
func NormalizeTag(tag string) string {
	// 1. Trim whitespace
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return ""
	}

	// 2. Insert dash before uppercase letters in camelCase
	tag = insertDashBeforeCaps(tag)

	// 3. Replace non-alphanumeric characters with dashes
	tag = replaceNonAlphaNum(tag)

	// 4. Convert to lowercase
	tag = strings.ToLower(tag)

	// 5. Collapse multiple consecutive dashes
	tag = collapseRepeatedDashes(tag)

	// 6. Trim leading/trailing dashes
	tag = strings.Trim(tag, "-")

	return tag
}

// NormalizeTags normalizes all tags in a bookmark's tag hierarchy.
// Each level of the hierarchy is normalized separately to preserve parent-child relationships.
func NormalizeTags(b *Bookmark) {
	if b.Tags == nil {
		return
	}

	for i, tagHierarchy := range b.Tags {
		for j, tag := range tagHierarchy {
			b.Tags[i][j] = NormalizeTag(tag)
		}
	}
}

// insertDashBeforeCaps inserts a dash before uppercase letters in camelCase words
// Example: "DevOps" → "Dev-Ops"
func insertDashBeforeCaps(s string) string {
	var result strings.Builder
	runes := []rune(s)

	for i, r := range runes {
		// Insert dash before uppercase if:
		// 1. Not the first character
		// 2. Previous character is lowercase
		// 3. OR next character is lowercase (for sequences like "XMLParser" → "XML-Parser")
		if i > 0 && unicode.IsUpper(r) {
			prev := runes[i-1]
			if unicode.IsLower(prev) {
				result.WriteRune('-')
			} else if i < len(runes)-1 && unicode.IsLower(runes[i+1]) {
				result.WriteRune('-')
			}
		}
		result.WriteRune(r)
	}

	return result.String()
}

// replaceNonAlphaNum replaces non-alphanumeric characters with dashes
func replaceNonAlphaNum(s string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9-]+`)
	return re.ReplaceAllString(s, "-")
}

// collapseRepeatedDashes collapses multiple consecutive dashes into a single dash
func collapseRepeatedDashes(s string) string {
	re := regexp.MustCompile(`-+`)
	return re.ReplaceAllString(s, "-")
}
