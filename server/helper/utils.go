package helper

import (
	"bytes"
	"encoding/base64"
	"html"

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

func HtmlOrCurlResponse(isHTML bool, response string) string {
	if isHTML {
		htmlOutput, err := MarkdownToHTML(response)
		if err != nil {
			return "Error converting Markdown to HTML: " + html.EscapeString(err.Error())
		}
		return htmlOutput
	}
	return response
}

func EncodeByteSliceToBase64(images []byte) string {
	base64Str := base64.StdEncoding.EncodeToString(images)
	return base64Str
}
