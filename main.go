package main

import (
	"fmt"
	"os"

	"svenska-nyheter-reader/internal/feed"
	"svenska-nyheter-reader/internal/ui"
)

func main() {
	rss, err := feed.Fetch(feed.FeedURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if err := ui.Run(rss, os.Getenv("DEEPL_API_KEY")); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
