package helper

import (
	"bytes"

	"github.com/yuin/goldmark"
)

// MarkdownToHTML converts a Markdown string to HTML using the goldmark library.
func MarkdownToHTML(markdown string) (string, error) {
	var buf bytes.Buffer
	md := goldmark.New()
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
