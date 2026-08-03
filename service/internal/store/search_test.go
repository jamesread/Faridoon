package store

import (
	"strings"
	"testing"
)

func TestBuildBooleanQuery(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"ab", ""},
		{"foo", "+foo*"},
		{"  foo   bar ", "+foo* +bar*"},
		{"foo+bar", "+foobar*"},
		{"a bc def", "+def*"},
		{"日本語", "+日本語*"}, // 3 runes → FULLTEXT
		{"éé", ""},       // 2 runes (4 bytes) → LIKE fallback
		{"ééé", "+ééé*"}, // 3 runes → FULLTEXT
	}
	for _, tt := range tests {
		if got := buildBooleanQuery(tt.in); got != tt.want {
			t.Errorf("buildBooleanQuery(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestEscapeLikePattern(t *testing.T) {
	got := escapeLikePattern(`100%_off|x`)
	want := `100|%|_off||x`
	if got != want {
		t.Errorf("escapeLikePattern = %q, want %q", got, want)
	}
}

func TestApprovedQuoteFilterSQL(t *testing.T) {
	filter := approvedQuoteFilterSQL()
	if filter != "q.approval = 1" {
		t.Fatalf("approvedQuoteFilterSQL = %q", filter)
	}
	// Search WHERE clauses must always include the approved-only predicate.
	fulltextWhere := filter + " AND MATCH(q.content) AGAINST (? IN BOOLEAN MODE)"
	likeWhere := filter + ` AND q.content LIKE ? ESCAPE '|'`
	for _, where := range []string{fulltextWhere, likeWhere, filter} {
		if !strings.Contains(where, "q.approval = 1") {
			t.Fatalf("search filter missing approval check: %q", where)
		}
		if strings.Contains(where, "approval = 0") {
			t.Fatalf("search filter must not include pending quotes: %q", where)
		}
	}
}
