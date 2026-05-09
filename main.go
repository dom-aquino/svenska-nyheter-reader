package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"svenska-nyheter-reader/internal/feed"
)

type articleItem struct {
	title    string
	category string
	content  string
}

func (a articleItem) Title() string       { return a.title }
func (a articleItem) Description() string { return a.category }
func (a articleItem) FilterValue() string { return a.title }

type viewState int

const (
	viewList viewState = iota
	viewArticle
)

type wordPos struct {
	line int
	word int
}

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Padding(0, 1)
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Padding(0, 1)
	cursorStyle = lipgloss.NewStyle().Reverse(true)
)

type model struct {
	list     list.Model
	viewport viewport.Model
	state    viewState
	words    [][]string
	cursor   wordPos
}

func (m model) Init() tea.Cmd { return nil }

func parseWords(content string) [][]string {
	lines := strings.Split(content, "\n")
	result := make([][]string, len(lines))
	for i, line := range lines {
		result[i] = strings.Fields(line)
	}
	return result
}

func renderArticle(words [][]string, cur wordPos) string {
	var sb strings.Builder
	for l, line := range words {
		for w, word := range line {
			if l == cur.line && w == cur.word {
				sb.WriteString(cursorStyle.Render(word))
			} else {
				sb.WriteString(word)
			}
			if w < len(line)-1 {
				sb.WriteString(" ")
			}
		}
		if l < len(words)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func (m *model) wordForward() {
	if len(m.words) == 0 {
		return
	}
	if m.cursor.word < len(m.words[m.cursor.line])-1 {
		m.cursor.word++
		return
	}
	for l := m.cursor.line + 1; l < len(m.words); l++ {
		if len(m.words[l]) > 0 {
			m.cursor.line = l
			m.cursor.word = 0
			return
		}
	}
}

func (m *model) wordBackward() {
	if len(m.words) == 0 {
		return
	}
	if m.cursor.word > 0 {
		m.cursor.word--
		return
	}
	for l := m.cursor.line - 1; l >= 0; l-- {
		if len(m.words[l]) > 0 {
			m.cursor.line = l
			m.cursor.word = len(m.words[l]) - 1
			return
		}
	}
}

func (m *model) lineDown() {
	for l := m.cursor.line + 1; l < len(m.words); l++ {
		if len(m.words[l]) > 0 {
			m.cursor.line = l
			if m.cursor.word >= len(m.words[l]) {
				m.cursor.word = len(m.words[l]) - 1
			}
			return
		}
	}
}

func (m *model) lineUp() {
	for l := m.cursor.line - 1; l >= 0; l-- {
		if len(m.words[l]) > 0 {
			m.cursor.line = l
			if m.cursor.word >= len(m.words[l]) {
				m.cursor.word = len(m.words[l]) - 1
			}
			return
		}
	}
}

func (m *model) scrollToCursor() {
	if m.cursor.line < m.viewport.YOffset {
		m.viewport.YOffset = m.cursor.line
	} else if m.cursor.line >= m.viewport.YOffset+m.viewport.Height {
		m.viewport.YOffset = m.cursor.line - m.viewport.Height + 1
	}
}

func (m *model) refreshArticle() {
	m.viewport.SetContent(renderArticle(m.words, m.cursor))
	m.scrollToCursor()
}

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
					m.words = parseWords(item.content)
					m.cursor = wordPos{}
					for l, line := range m.words {
						if len(line) > 0 {
							m.cursor.line = l
							break
						}
					}
					m.viewport.SetContent(renderArticle(m.words, m.cursor))
					m.viewport.GotoTop()
					m.state = viewArticle
					return m, nil
				}
			}
		case "w":
			if m.state == viewArticle {
				m.wordForward()
				m.refreshArticle()
				return m, nil
			}
		case "b":
			if m.state == viewArticle {
				m.wordBackward()
				m.refreshArticle()
				return m, nil
			}
		case "j", "down":
			if m.state == viewArticle {
				m.lineDown()
				m.refreshArticle()
				return m, nil
			}
		case "k", "up":
			if m.state == viewArticle {
				m.lineUp()
				m.refreshArticle()
				return m, nil
			}
		}
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 2
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
		footer := helpStyle.Render("w/b word · j/k row · esc back")
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
