package quote

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var reversedLinkRE = regexp.MustCompile(`\(([^)\n]+)\)\[([^\]\n]+)\]`)

var markdownEngine = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	goldmark.WithRendererOptions(html.WithHardWraps()),
)

var markdownSanitizer = func() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	p.AddTargetBlankToFullyQualifiedLinks(true)
	p.RequireNoFollowOnLinks(true)
	return p
}()

// NormalizeInput applies newline normalization and Discord paste fixes before formatting.
func NormalizeInput(raw string) string {
	return FixDiscordLinebreaks(NormalizeNewlines(strings.TrimSpace(raw)))
}

func normalizeMarkdownLinks(text string) string {
	return reversedLinkRE.ReplaceAllString(text, "[$2]($1)")
}

// RenderMarkdown converts Markdown to sanitized HTML for quote display.
func RenderMarkdown(text string, inline bool) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	text = normalizeMarkdownLinks(text)
	var buf bytes.Buffer
	if err := markdownEngine.Convert([]byte(text), &buf); err != nil {
		return ""
	}
	out := markdownSanitizer.Sanitize(buf.String())
	if inline {
		out = strings.TrimSpace(out)
		out = strings.TrimPrefix(out, "<p>")
		out = strings.TrimSuffix(out, "</p>")
	}
	return out
}
