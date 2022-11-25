package selector

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	style     = lipgloss.NewStyle()
	lineStyle = lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{Light: "#eee", Dark: "#333"}).
			Padding(0, 1, 0, 1)

	lineSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Padding(0, 1, 0, 1).
				Background(lipgloss.Color("#33aa66")).
				Foreground(lipgloss.Color("#000000"))
)

type IntellisenseItem struct {
	Value string
	Kind  string
}

type Model struct {
	// The focus status
	focused bool
	// viewport is the vertically-scrollable viewport of the multi-line text
	// input.
	viewport *viewport.Model

	offset int

	cursor int
	items  []IntellisenseItem
}

func New() Model {
	vp := viewport.New(20, 8)
	m := Model{
		focused:  true,
		viewport: &vp,
		cursor:   0,
		offset:   0,
	}

	return m

}

func (m *Model) Focus() tea.Cmd {
	m.focused = true
	return nil
}

// Deactivate focus
func (m *Model) Blur() tea.Cmd {
	m.focused = false
	return nil
}

func (m *Model) Reset() {
	m.cursor = 0
	m.items = make([]IntellisenseItem, 0)
}

// Reports focus status
func (m Model) Focused() bool {
	return m.focused
}

func (m *Model) SetOffset(offset int) {
	m.offset = offset
}

func (m *Model) SetItems(items []IntellisenseItem) {
	m.items = items
}

func (m Model) SelectedItem() *IntellisenseItem {
	if len(m.items) > 0 {
		return &m.items[m.cursor]
	}
	return nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focused {
		return *m, nil
	}
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			if m.cursor-1 >= 0 {
				m.cursor--
			}
		case tea.KeyDown:
			if m.cursor+1 < len(m.items) {
				m.cursor++
			}
			// case tea.KeyEnter:
			// 	m.focused = false
			// 	m.items = make([]IntellisenseItem, 0)
		}
	}

	vp, cmd := m.viewport.Update(msg)
	m.viewport = &vp
	cmds = append(cmds, cmd)

	return *m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if len(m.items) == 0 {
		return ""
	}
	sb := new(strings.Builder)
	sb.WriteString("\n")
	for r, i := range m.items {
		var out string
		if r == m.cursor {
			out = fmt.Sprintf("%s\n", lineSelectedStyle.Render(i.Value))
		} else {
			out = fmt.Sprintf("%s\n", lineStyle.Render(i.Value))
		}
		sb.WriteString(out)
	}

	m.viewport.SetContent(sb.String())
	blank := strings.Repeat(" ", m.offset)
	out := style.Render(m.viewport.View())

	withOffset := new(strings.Builder)
	for _, l := range strings.Split(out, "\n") {
		withOffset.WriteString(fmt.Sprintf("%s%s\n", blank, l))
	}
	return withOffset.String()
}
