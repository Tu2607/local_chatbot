package helper

import (
	"bytes"
	"encoding/base64"
	"html"
	"log"
	"net"
	"strings"

	"time"

	"math/rand"

	"local_chatbot/server/template"

	"github.com/oklog/ulid/v2"
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

func GenerateULID() string {
	entropy := rand.New(rand.NewSource(time.Now().UnixNano()))
	ms := ulid.Timestamp(time.Now())
	new_ulid, err := ulid.New(ms, entropy)
	if err != nil {
		log.Println("Error generating ULID:", err)
	}

	return new_ulid.String()
}

func ReverseSlice(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// A helper function that combine multiple chat message into one big text body with new lines
func CombineChatMessages(messages []template.Message) string {
	var combined strings.Builder
	for _, msg := range messages {
		combined.WriteString(msg.Role + ": ")
		combined.WriteString(msg.Content)
		combined.WriteString("\n")
	}
	return combined.String()
}

func CheckPort(port string) bool {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		// Port is already in use or not accessible
		// May need more logging here in the future
		return false
	}
	listener.Close()
	// Port is available
	return true
}
