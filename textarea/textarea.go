package textarea

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	minHeight     = 1
	minWidth      = 2
	defaultHeight = 6
	defaultWidth  = 40
	maxHeight     = 99
	maxWidth      = 500
)

// KeyMap is the key bindings for different actions within the textarea.
type KeyMap struct {
	CharacterBackward       key.Binding
	CharacterForward        key.Binding
	DeleteCharacterBackward key.Binding
	InsertNewline           key.Binding
}

var DefaultKeyMap = KeyMap{
	CharacterForward:        key.NewBinding(key.WithKeys("right", "ctrl+f")),
	CharacterBackward:       key.NewBinding(key.WithKeys("left", "ctrl+b")),
	InsertNewline:           key.NewBinding(key.WithKeys("enter", "ctrl+m")),
	DeleteCharacterBackward: key.NewBinding(key.WithKeys("backspace", "ctrl+h")),
}

type Model struct {
	focused bool

	// Underlying text value.
	value [][]rune

	// Cursor column.
	col int
	// Cursor row.
	row int

	// KeyMap encodes the keybindings recognized by the widget.
	KeyMap KeyMap

	// viewport is the vertically-scrollable viewport of the multi-line text
	// input.
	viewport *viewport.Model
}

func New() Model {
	vp := viewport.New(0, 0)

	m := Model{
		focused:  true,
		viewport: &vp,

		col: 0,
		row: 0,

		KeyMap: DefaultKeyMap,

		value: make([][]rune, minHeight, maxHeight),
	}

	vp.Width = defaultWidth
	vp.Height = defaultHeight
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Focus() tea.Cmd {
	m.focused = true
	return nil
}

func (m *Model) Blur() tea.Cmd {
	m.focused = false
	return nil
}

func (m Model) Focused() bool {
	return m.focused
}

func (m *Model) splitLine(row, col int) {
	// To perform a split, take the current line and keep the content before
	// the cursor, take the content after the cursor and make it the content of
	// the line underneath, and shift the remaining lines down by one
	head, tailSrc := m.value[row][:col], m.value[row][col:]
	tail := make([]rune, len(tailSrc))
	copy(tail, tailSrc)

	m.value = append(m.value[:row+1], m.value[row:]...)

	m.value[row] = head
	m.value[row+1] = tail

	m.col = 0
	m.row++
}

// mergeLineAbove merges the current line the cursor is on with the line above.
func (m *Model) mergeLineAbove(row int) {
	if row <= 0 {
		return
	}

	m.col = len(m.value[row-1])
	m.row = m.row - 1

	// To perform a merge, we will need to combine the two lines and then
	m.value[row-1] = append(m.value[row-1], m.value[row]...)

	// Shift all lines up by one
	for i := row; i < len(m.value)-1; i++ {
		m.value[i] = m.value[i+1]
	}

	// And, remove the last line
	if len(m.value) > 0 {
		m.value = m.value[:len(m.value)-1]
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	// var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.CharacterForward):
		case key.Matches(msg, m.KeyMap.CharacterBackward):
		case key.Matches(msg, m.KeyMap.DeleteCharacterBackward):
			m.col = clamp(m.col, 0, len(m.value[m.row]))
			if m.col <= 0 {
				m.mergeLineAbove(m.row)
				break
			}
			if len(m.value[m.row]) > 0 {
				m.value[m.row] = append(m.value[m.row][:max(0, m.col-1)], m.value[m.row][m.col:]...)
				// if m.col > 0 {
				// 	m.SetCursor(m.col - 1)
				// }
			}
		case key.Matches(msg, m.KeyMap.InsertNewline):
			if len(m.value) >= maxHeight {
				return m, nil
			}
			m.col = clamp(m.col, 0, len(m.value[m.row]))
			m.splitLine(m.row, m.col)
		default:
			m.col = min(m.col, len(m.value[m.row]))
			m.value[m.row] = append(m.value[m.row][:m.col], append(msg.Runes, m.value[m.row][m.col:]...)...)
			m.col = clamp(m.col+len(msg.Runes), 0, len(m.value[m.row]))
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	style := lipgloss.NewStyle().Background(lipgloss.Color("1"))

	sb := new(strings.Builder)

	for ir, r := range m.value {
		needCursor := false
		haveCursor := false
		if ir == m.row {
			needCursor = true
		}
		for ic, c := range r {
			if ic == m.col && needCursor {
				sb.WriteString(style.Render(string(c)))
				haveCursor = true
			} else {
				sb.WriteRune(c)
			}
		}
		if needCursor && !haveCursor {
			sb.WriteString(style.Render(" "))
		}
		sb.WriteString("\n")
	}

	m.viewport.SetContent(sb.String())
	return m.viewport.View()
}

func clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
