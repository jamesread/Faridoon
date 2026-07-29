package headerlink

import (
	"fmt"
	"net/url"
	"strings"
)

func NormalizeTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return "", fmt.Errorf("title required")
	}
	if len(title) > 64 {
		return "", fmt.Errorf("title too long")
	}
	return title, nil
}

func NormalizeURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("URL required")
	}
	if len(raw) > 2048 {
		return "", fmt.Errorf("URL too long")
	}
	if strings.HasPrefix(raw, "/") {
		return normalizeRelativeURL(raw)
	}
	return normalizeAbsoluteURL(raw)
}

func normalizeRelativeURL(raw string) (string, error) {
	if strings.ContainsAny(raw, " \t\n\r") {
		return "", fmt.Errorf("URL must not contain whitespace")
	}
	if strings.HasPrefix(raw, "//") {
		return "", fmt.Errorf("URL must be a path or http(s) URL")
	}
	return raw, nil
}

func normalizeAbsoluteURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("URL is not valid")
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", fmt.Errorf("URL scheme must be http or https")
	}
	if strings.TrimSpace(u.Host) == "" {
		return "", fmt.Errorf("URL host is required")
	}
	return raw, nil
}
