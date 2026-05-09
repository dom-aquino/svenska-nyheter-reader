package main

import (
	"fmt"
	"os"

	"svenska-nyheter-reader/internal/feed"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type articleItem struct {
	title string
	desc  string
}

func (a articleItem) Title() string       { return a.title }
func (a articleItem) Description() string { return a.desc }
func (a articleItem) FilterValue() string { return a.title }

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string { return m.list.View() }

func main() {
	rss, err := feed.Fetch(feed.FeedURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	var items []list.Item
	for _, item := range rss.Channel.Items {
		if item.Enclosure != nil {
			continue
		}
		items = append(items, articleItem{
			title: item.Title,
			desc:  item.Categories[0],
		})
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = rss.Channel.Title

	p := tea.NewProgram(model{list: l}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
