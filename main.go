package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"svenska-nyheter-reader/internal/feed"
)

type articleItem struct {
	title   string
	category string
	content string
}

func (a articleItem) Title() string       { return a.title }
func (a articleItem) Description() string { return a.category }
func (a articleItem) FilterValue() string { return a.title }

type viewState int

const (
	viewList viewState = iota
	viewArticle
)

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Padding(0, 1)
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Padding(0, 1)
)

type model struct {
	list     list.Model
	viewport viewport.Model
	state    viewState
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.state == viewArticle {
				m.state = viewList
				return m, nil
			}
			return m, tea.Quit
		case "esc":
			if m.state == viewArticle {
				m.state = viewList
				return m, nil
			}
		case "enter":
			if m.state == viewList {
				if item, ok := m.list.SelectedItem().(articleItem); ok {
					m.viewport.SetContent(item.content)
					m.viewport.GotoTop()
					m.state = viewArticle
					return m, nil
				}
			}
		}
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 2 // header + footer
	}

	if m.state == viewArticle {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.state == viewArticle {
		item := m.list.SelectedItem().(articleItem)
		header := headerStyle.Render(item.title)
		footer := helpStyle.Render("↑/↓ · esc back")
		return header + "\n" + m.viewport.View() + "\n" + footer
	}
	return m.list.View()
}

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

	p := tea.NewProgram(model{list: l, viewport: viewport.New(0, 0)}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
