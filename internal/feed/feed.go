package feed

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const FeedURL = "https://8sidor.se/feed/"

type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	Language    string `xml:"language"`
	Items       []Item `xml:"item"`
}

type Enclosure struct {
	Type string `xml:"type,attr"`
}

type Item struct {
	Title       string     `xml:"title"`
	Link        string     `xml:"link"`
	Description string     `xml:"description"`
	Content     string     `xml:"http://purl.org/rss/1.0/modules/content/ encoded"`
	PubDate     string     `xml:"pubDate"`
	Categories  []string   `xml:"category"`
	Creator     string     `xml:"http://purl.org/dc/elements/1.1/ creator"`
	Enclosure   *Enclosure `xml:"enclosure"`
}

var (
	tagRe   = regexp.MustCompile(`<[^>]+>`)
	spaceRe = regexp.MustCompile(`\n{3,}`)
)

// StripHTML converts HTML to plain text suitable for terminal display.
func StripHTML(s string) string {
	s = strings.ReplaceAll(s, "<br />", "\n")
	s = strings.ReplaceAll(s, "<br/>", "\n")
	s = strings.ReplaceAll(s, "<br>", "\n")
	s = strings.ReplaceAll(s, "</p>", "\n\n")
	s = tagRe.ReplaceAllString(s, "")
	s = spaceRe.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}

func Fetch(url string) (*RSS, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status %d", resp.StatusCode)
	}

	return Parse(resp.Body)
}

func Parse(r io.Reader) (*RSS, error) {
	var rss RSS
	if err := xml.NewDecoder(r).Decode(&rss); err != nil {
		return nil, fmt.Errorf("parse RSS: %w", err)
	}
	return &rss, nil
}
