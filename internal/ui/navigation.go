package ui

import "strings"

func parseWords(content string) [][]string {
	lines := strings.Split(content, "\n")
	result := make([][]string, len(lines))
	for i, line := range lines {
		result[i] = strings.Fields(line)
	}
	return result
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
