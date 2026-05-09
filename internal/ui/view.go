package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

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

func cursorScreenX(words [][]string, cur wordPos) int {
	if cur.line >= len(words) {
		return 0
	}
	x := 0
	for i := 0; i < cur.word && i < len(words[cur.line]); i++ {
		x += lipgloss.Width(words[cur.line][i]) + 1
	}
	return x
}

func (m model) cursorScreenY() int {
	return (m.cursor.line - m.viewport.YOffset) + 1
}

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

		left := ansi.Truncate(bLine, px, "")
		lWidth := lipgloss.Width(left)
		if lWidth < px {
			left += strings.Repeat(" ", px-lWidth)
		}

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
