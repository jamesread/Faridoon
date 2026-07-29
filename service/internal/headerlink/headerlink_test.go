package headerlink

import "testing"

func TestNormalizeTitle(t *testing.T) {
	got, err := NormalizeTitle("  Docs  ")
	if err != nil || got != "Docs" {
		t.Fatalf("got %q %v", got, err)
	}
	if _, err := NormalizeTitle("   "); err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestNormalizeURL(t *testing.T) {
	cases := []struct {
		in      string
		wantErr bool
	}{
		{"/about", false},
		{"https://example.com/docs", false},
		{"http://example.com", false},
		{"//evil.example", true},
		{"javascript:alert(1)", true},
		{"ftp://example.com", true},
		{"", true},
		{"  ", true},
	}
	for _, tc := range cases {
		_, err := NormalizeURL(tc.in)
		if tc.wantErr && err == nil {
			t.Fatalf("%q: expected error", tc.in)
		}
		if !tc.wantErr && err != nil {
			t.Fatalf("%q: unexpected error %v", tc.in, err)
		}
	}
}
