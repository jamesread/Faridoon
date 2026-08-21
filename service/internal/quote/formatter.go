package quote

import (
	"html"
	"regexp"
	"strings"
)

var lineRE = regexp.MustCompile(`(?i)^[\]\[\(\):\d ]*<?[&+@~]{0,1}([\w\- ]+)[:>] (.*)`)

const FormatStyleSignatureQuote = "signatureQuote"

type Line struct {
	Content       string
	Username      string
	UsernameColor int
}

type Formatted struct {
	Content            string
	Created            string
	SyntaxHighlighting string
	FormatStyle        string
	SignatureAuthor    string
	Lines              []Line
	ID                 int
	VoteCount          int
	Approved           bool
}

type Formatter struct {
	usernameColors map[string]int
	colorIndex     int
}

func NewFormatter() *Formatter {
	return &Formatter{}
}

func (f *Formatter) Format(id int, raw, created string, voteCount int, approved bool, syntax string) Formatted {
	f.colorIndex = 1
	f.usernameColors = map[string]int{}
	content := html.UnescapeString(strings.ReplaceAll(raw, `\'`, "'"))
	content = NormalizeNewlines(content)
	formatStyle, signatureAuthor, body := detectSignatureQuote(content)
	return Formatted{
		ID: id, Content: content, Created: created, VoteCount: voteCount,
		Approved: approved, SyntaxHighlighting: syntax, FormatStyle: formatStyle,
		SignatureAuthor: signatureAuthor, Lines: f.parseLines(body),
	}
}

func detectSignatureQuote(content string) (formatStyle, signatureAuthor, body string) {
	lines := strings.Split(content, "\n")
	if len(lines) < 2 {
		return "", "", content
	}
	lines = trimTrailingEmptyLines(lines)
	if len(lines) < 2 {
		return "", "", content
	}
	author, ok := signatureAuthorFromLine(lines[len(lines)-1])
	if !ok {
		return "", "", content
	}
	return FormatStyleSignatureQuote, author, strings.Join(lines[:len(lines)-1], "\n")
}

func trimTrailingEmptyLines(lines []string) []string {
	for len(lines) >= 2 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func signatureAuthorFromLine(lastLine string) (author string, ok bool) {
	if !strings.HasPrefix(lastLine, "- ") {
		return "", false
	}
	author = strings.TrimSpace(lastLine[2:])
	return author, author != ""
}

func (f *Formatter) parseLines(content string) []Line {
	var lines []Line
	for _, line := range strings.Split(content, "\n") {
		lines = append(lines, f.parseLine(line))
	}
	return lines
}

func (f *Formatter) parseLine(line string) Line {
	out := Line{Content: line}
	m := lineRE.FindStringSubmatch(line)
	if len(m) != 3 {
		return out
	}
	msg := strings.TrimSpace(strings.ReplaceAll(m[2], "<br />", ""))
	if msg == "" {
		return out
	}
	out.Username = m[1]
	out.UsernameColor = f.colorFor(m[1])
	out.Content = m[2]
	return out
}

func (f *Formatter) colorFor(username string) int {
	if c, ok := f.usernameColors[username]; ok {
		return c
	}
	c := f.colorIndex
	f.colorIndex++
	f.usernameColors[username] = c
	return c
}

var discordLinebreakRE = regexp.MustCompile(`(?i)\[\n(?P<date>\d\d:\d\d)\n\]\n(?P<username>[\w\d_]+)\n:\n`)

func FixDiscordLinebreaks(content string) string {
	return discordLinebreakRE.ReplaceAllString(content, "${username}: ")
}

func NormalizeNewlines(content string) string {
	return strings.ReplaceAll(content, "\r\n", "\n")
}
