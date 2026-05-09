package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"svenska-nyheter-reader/internal/translate"
)

const maxContentWidth = 72

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

type articleItem struct {
	title    string
	category string
	content  string
}

func (a articleItem) Title() string       { return a.title }
func (a articleItem) Description() string { return a.category }
func (a articleItem) FilterValue() string { return a.title }

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
