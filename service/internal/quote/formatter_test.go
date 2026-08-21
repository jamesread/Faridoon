package quote

import "testing"

func TestFormatUsernameLineCount(t *testing.T) {
	got := NewFormatter().Format(1, "alice: hello\nbob: world", "2024-01-01", 3, true, "")
	if len(got.Lines) != 2 {
		t.Fatalf("lines=%d", len(got.Lines))
	}
}

func TestFormatUsernameLineAlice(t *testing.T) {
	got := NewFormatter().Format(1, "alice: hello\nbob: world", "2024-01-01", 3, true, "")
	if got.Lines[0].Username != "alice" || got.Lines[0].UsernameColor != 1 {
		t.Fatalf("line0=%+v", got.Lines[0])
	}
}

func TestFormatUsernameLineBob(t *testing.T) {
	got := NewFormatter().Format(1, "alice: hello\nbob: world", "2024-01-01", 3, true, "")
	if got.Lines[1].Username != "bob" || got.Lines[1].UsernameColor != 2 {
		t.Fatalf("line1=%+v", got.Lines[1])
	}
}

func TestNormalizeNewlines(t *testing.T) {
	if NormalizeNewlines("a\r\nb") != "a\nb" {
		t.Fatal("normalize failed")
	}
}

func TestFormatSignatureQuote(t *testing.T) {
	got := NewFormatter().Format(1, "alice: hello\nbob: world\n- Oscar Wilde", "2024-01-01", 0, true, "")
	if got.FormatStyle != FormatStyleSignatureQuote {
		t.Fatalf("formatStyle=%q", got.FormatStyle)
	}
	if got.SignatureAuthor != "Oscar Wilde" {
		t.Fatalf("signatureAuthor=%q", got.SignatureAuthor)
	}
	if len(got.Lines) != 2 {
		t.Fatalf("lines=%d", len(got.Lines))
	}
}

func TestFormatSignatureQuoteRequiresNewlineBeforeAuthor(t *testing.T) {
	got := NewFormatter().Format(1, "- Oscar Wilde", "2024-01-01", 0, true, "")
	if got.FormatStyle != "" {
		t.Fatalf("formatStyle=%q", got.FormatStyle)
	}
	if len(got.Lines) != 1 || got.Lines[0].Content != "- Oscar Wilde" {
		t.Fatalf("lines=%+v", got.Lines)
	}
}

func TestFormatSignatureQuoteRequiresDashPrefix(t *testing.T) {
	got := NewFormatter().Format(1, "hello\nOscar Wilde", "2024-01-01", 0, true, "")
	if got.FormatStyle != "" {
		t.Fatalf("formatStyle=%q", got.FormatStyle)
	}
}
