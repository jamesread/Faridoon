package quote

import (
	"strings"
	"testing"
)

func TestNormalizeMarkdownLinks(t *testing.T) {
	got := normalizeMarkdownLinks("(https://en.wikipedia.org/wiki/Arthur_Ashe)[Arthur Ashe]")
	want := "[Arthur Ashe](https://en.wikipedia.org/wiki/Arthur_Ashe)"
	if got != want {
		t.Fatalf("got=%q want=%q", got, want)
	}
}

func TestRenderMarkdownBold(t *testing.T) {
	got := RenderMarkdown("**bold**", false)
	if !strings.Contains(got, "<strong>bold</strong>") {
		t.Fatalf("got=%q", got)
	}
}

func TestRenderMarkdownStripsScript(t *testing.T) {
	got := RenderMarkdown("<script>alert(1)</script>", false)
	if strings.Contains(strings.ToLower(got), "<script") {
		t.Fatalf("got=%q", got)
	}
}

func TestRenderMarkdownReversedLink(t *testing.T) {
	got := RenderMarkdown("(https://en.wikipedia.org/wiki/Arthur_Ashe)[Arthur Ashe]", true)
	if !strings.Contains(got, `href="https://en.wikipedia.org/wiki/Arthur_Ashe"`) {
		t.Fatalf("got=%q", got)
	}
	if !strings.Contains(got, "Arthur Ashe") {
		t.Fatalf("got=%q", got)
	}
}

func TestNormalizeInputFixesDiscordPaste(t *testing.T) {
	raw := "[\n12:34\n]\nalice\n:\nhello"
	got := NormalizeInput(raw)
	if got != "alice: hello" {
		t.Fatalf("got=%q", got)
	}
}

func TestNormalizeInputFixesDiscordPasteCRLF(t *testing.T) {
	raw := "[\r\n12:34\r\n]\r\nalice\r\n:\r\nhello"
	got := NormalizeInput(raw)
	if got != "alice: hello" {
		t.Fatalf("got=%q", got)
	}
}

func TestNormalizeInputFixesMultipleDiscordPastes(t *testing.T) {
	raw := "[\n12:34\n]\nalice\n:\nhello\n[\n12:35\n]\nbob\n:\nworld"
	got := NormalizeInput(raw)
	if got != "alice: hello\nbob: world" {
		t.Fatalf("got=%q", got)
	}
}
