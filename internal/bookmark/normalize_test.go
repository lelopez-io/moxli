package bookmark

import "testing"

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "basic URL",
			input: "https://example.com/path",
			want:  "https://example.com/path",
		},
		{
			name:  "trailing slash removed",
			input: "https://example.com/path/",
			want:  "https://example.com/path",
		},
		{
			name:  "root path keeps slash",
			input: "https://example.com/",
			want:  "https://example.com/",
		},
		{
			name:  "lowercase host",
			input: "https://Example.COM/path",
			want:  "https://example.com/path",
		},
		{
			name:  "missing scheme defaults to https",
			input: "example.com/path",
			want:  "https://example.com/path",
		},
		{
			name:  "http scheme preserved",
			input: "http://example.com/path",
			want:  "http://example.com/path",
		},
		{
			name:  "query params sorted",
			input: "https://example.com?z=1&a=2",
			want:  "https://example.com?a=2&z=1",
		},
		{
			name:  "fragment removed",
			input: "https://example.com/path#section",
			want:  "https://example.com/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NormalizeURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeBookmarkURL(t *testing.T) {
	b := &Bookmark{
		URL: "https://Example.COM/path/",
	}

	err := NormalizeBookmarkURL(b)
	if err != nil {
		t.Fatalf("NormalizeBookmarkURL() error = %v", err)
	}

	want := "https://example.com/path"
	if b.NormalizedURL != want {
		t.Errorf("NormalizedURL = %v, want %v", b.NormalizedURL, want)
	}
}
