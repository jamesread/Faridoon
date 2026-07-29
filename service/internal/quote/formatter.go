package quote

import (
	"html"
	"regexp"
	"strings"
)

var lineRE = regexp.MustCompile(`(?i)^[\]\[\(\):\d ]*<?[&+@~]{0,1}([\w\- ]+)[:>] (.*)`)

type Line struct {
	Content       string
	Username      string
	UsernameColor int
}

type Formatted struct {
	Content            string
	Created            string
	SyntaxHighlighting string
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
	return Formatted{
		ID: id, Content: content, Created: created, VoteCount: voteCount,
		Approved: approved, SyntaxHighlighting: syntax, Lines: f.parseLines(content),
	}
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

func FixDiscordLinebreaks(content string) string {
	re := regexp.MustCompile(`(?i)\[\n(?P<date>\d\d:\d\d)\n\]\n(?P<username>[\w\d_]+)\n:\n`)
	return re.ReplaceAllString(content, "${username}: ")
}

func NormalizeNewlines(content string) string {
	return strings.ReplaceAll(content, "\r\n", "\n")
}
