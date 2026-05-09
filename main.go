package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"svenska-nyheter-reader/internal/feed"
	"svenska-nyheter-reader/internal/translate"
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

type translationMsg struct {
	word        string
	translation string
}

const maxContentWidth = 72

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Padding(0, 1)
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Padding(0, 1)
	cursorStyle = lipgloss.NewStyle().Reverse(true)
	popupStyle  = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 1)
)

type model struct {
	list            list.Model
	viewport        viewport.Model
	state           viewState
	words           [][]string
	cursor          wordPos
	translation     string
	showTranslation bool
	apiKey          string
	width           int
	height          int
	leftPad         int
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

func currentWord(m model) string {
	if len(m.words) == 0 || m.cursor.line >= len(m.words) {
		return ""
	}
	line := m.words[m.cursor.line]
	if m.cursor.word >= len(line) {
		return ""
	}
	return line[m.cursor.word]
}

func indentLines(s string, n int) string {
	if n <= 0 {
		return s
	}
	pad := strings.Repeat(" ", n)
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = pad + l
	}
	return strings.Join(lines, "\n")
}

// cursorScreenX returns the visual column of the current word on its line.
func cursorScreenX(words [][]string, cur wordPos) int {
	if cur.line >= len(words) {
		return 0
	}
	x := 0
	for i := 0; i < cur.word && i < len(words[cur.line]); i++ {
		x += lipgloss.Width(words[cur.line][i]) + 1 // word + space
	}
	return x
}

// cursorScreenY returns the screen row of the cursor line (0-based, accounting for header).
func (m model) cursorScreenY() int {
	return (m.cursor.line - m.viewport.YOffset) + 1 // +1 for header
}

// overlayAt renders popup on top of base at position (px, py).
func overlayAt(base, popup string, px, py int) string {
	baseLines := strings.Split(base, "\n")
	popupLines := strings.Split(popup, "\n")

	result := make([]string, len(baseLines))
	copy(result, baseLines)

	for i, pLine := range popupLines {
		y := py + i
		if y < 0 || y >= len(result) {
			continue
		}
		pWidth := lipgloss.Width(pLine)
		bLine := result[y]

		// Left: base truncated to px columns.
		left := ansi.Truncate(bLine, px, "")
		lWidth := lipgloss.Width(left)
		if lWidth < px {
			left += strings.Repeat(" ", px-lWidth)
		}

		// Right: plain-text remainder after the popup ends.
		stripped := []rune(ansi.Strip(bLine))
		right := ""
		if end := px + pWidth; end < len(stripped) {
			right = string(stripped[end:])
		}

		result[y] = left + pLine + right
	}
	return strings.Join(result, "\n")
}

func buildPopup(word, translation string) string {
	var body string
	if translation == "" {
		body = word + "\n·"
	} else {
		body = word + "\n" + translation
	}
	return popupStyle.Render(body)
}

func fetchTranslation(word, apiKey string) tea.Cmd {
	if apiKey == "" || word == "" {
		return nil
	}
	return func() tea.Msg {
		result, err := translate.Word(word, apiKey)
		if err != nil {
			return translationMsg{word: word}
		}
		return translationMsg{word: word, translation: result}
	}
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
	case translationMsg:
		if msg.word == currentWord(m) {
			m.translation = msg.translation
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.state == viewArticle {
				m.showTranslation = false
				m.state = viewList
				return m, nil
			}
			return m, tea.Quit
		case "esc":
			if m.state == viewArticle {
				if m.showTranslation {
					m.showTranslation = false
					return m, nil
				}
				m.state = viewList
				return m, nil
			}
		case " ":
			if m.state == viewArticle {
				if m.showTranslation {
					m.showTranslation = false
					return m, nil
				}
				m.showTranslation = true
				m.translation = ""
				return m, fetchTranslation(currentWord(m), m.apiKey)
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
					m.translation = ""
					m.showTranslation = false
					m.viewport.SetContent(renderArticle(m.words, m.cursor))
					m.viewport.GotoTop()
					m.state = viewArticle
					return m, nil
				}
			}
		case "w":
			if m.state == viewArticle {
				m.showTranslation = false
				m.translation = ""
				m.wordForward()
				m.refreshArticle()
				return m, nil
			}
		case "b":
			if m.state == viewArticle {
				m.showTranslation = false
				m.translation = ""
				m.wordBackward()
				m.refreshArticle()
				return m, nil
			}
		case "j", "down":
			if m.state == viewArticle {
				m.showTranslation = false
				m.translation = ""
				m.lineDown()
				m.refreshArticle()
				return m, nil
			}
		case "k", "up":
			if m.state == viewArticle {
				m.showTranslation = false
				m.translation = ""
				m.lineUp()
				m.refreshArticle()
				return m, nil
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		vw := msg.Width
		if vw > maxContentWidth {
			vw = maxContentWidth
		}
		m.leftPad = (msg.Width - vw) / 2
		m.list.SetSize(vw, msg.Height)
		m.viewport.Width = vw
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
		footer := helpStyle.Render("space translate · w/b word · j/k row · esc back")
		base := indentLines(header+"\n"+m.viewport.View()+"\n"+footer, m.leftPad)

		if m.showTranslation {
			popup := buildPopup(currentWord(m), m.translation)
			popupLines := strings.Split(popup, "\n")
			popupH := len(popupLines)
			popupW := lipgloss.Width(popupLines[0])

			px := cursorScreenX(m.words, m.cursor) + m.leftPad
			py := m.cursorScreenY() + 1

			if py+popupH >= m.height {
				py = m.cursorScreenY() - popupH
			}
			if px+popupW > m.width {
				px = m.width - popupW
			}
			if px < 0 {
				px = 0
			}

			return overlayAt(base, popup, px, py)
		}

		return base
	}
	return indentLines(m.list.View(), m.leftPad)
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

	p := tea.NewProgram(
		model{list: l, viewport: viewport.New(0, 0), apiKey: os.Getenv("DEEPL_API_KEY")},
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
