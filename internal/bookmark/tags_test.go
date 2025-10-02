package bookmark

import "testing"

func TestNormalizeTag(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "spaces to dashes",
			input: "User Auth",
			want:  "user-auth",
		},
		{
			name:  "camelCase to kebab-case",
			input: "DevOps",
			want:  "dev-ops",
		},
		{
			name:  "multiple words with spaces",
			input: "GraphQL API",
			want:  "graph-ql-api",
		},
		{
			name:  "already lowercase",
			input: "infrastructure",
			want:  "infrastructure",
		},
		{
			name:  "multiple spaces collapsed",
			input: "Web  Development",
			want:  "web-development",
		},
		{
			name:  "leading and trailing spaces",
			input: "  Security  ",
			want:  "security",
		},
		{
			name:  "special characters removed",
			input: "API's & Tools",
			want:  "api-s-tools",
		},
		{
			name:  "consecutive dashes collapsed",
			input: "foo--bar",
			want:  "foo-bar",
		},
		{
			name:  "mixed case with acronyms",
			input: "XMLParser",
			want:  "xml-parser",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeTag(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeTags(t *testing.T) {
	b := &Bookmark{
		Tags: [][]string{
			{"Muses", "Portfolio"},
			{"Security", "User Auth"},
			{"DevOps"},
		},
	}

	NormalizeTags(b)

	want := [][]string{
		{"muses", "portfolio"},
		{"security", "user-auth"},
		{"dev-ops"},
	}

	if len(b.Tags) != len(want) {
		t.Fatalf("len(Tags) = %v, want %v", len(b.Tags), len(want))
	}

	for i := range b.Tags {
		if len(b.Tags[i]) != len(want[i]) {
			t.Errorf("len(Tags[%d]) = %v, want %v", i, len(b.Tags[i]), len(want[i]))
			continue
		}
		for j := range b.Tags[i] {
			if b.Tags[i][j] != want[i][j] {
				t.Errorf("Tags[%d][%d] = %v, want %v", i, j, b.Tags[i][j], want[i][j])
			}
		}
	}
}

func TestNormalizeTags_Nil(t *testing.T) {
	b := &Bookmark{
		Tags: nil,
	}

	NormalizeTags(b) // Should not panic

	if b.Tags != nil {
		t.Errorf("Tags should remain nil")
	}
}
