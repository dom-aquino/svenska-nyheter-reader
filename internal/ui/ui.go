package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"svenska-nyheter-reader/internal/feed"
)

func Run(rss *feed.RSS, apiKey string) error {
	var items []list.Item
	for _, item := range rss.Channel.Items {
		if item.Enclosure != nil {
			continue
		}
		content := feed.StripHTML(item.Content)
		if content == "" {
			content = feed.StripHTML(item.Description)
		}
		items = append(items, articleItem{
			title:    item.Title,
			category: item.Categories[0],
			content:  content,
		})
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = rss.Channel.Title

	p := tea.NewProgram(
		model{list: l, viewport: viewport.New(0, 0), apiKey: apiKey},
		tea.WithAltScreen(),
	)
	_, err := p.Run()
	return err
}
