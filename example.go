package main

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ferama/bubble-texteditor/texteditor"
)

var style = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder())

const (
	width  = 34
	height = 20
)

func main() {
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type errMsg error

type model struct {
	texteditor texteditor.Model
	err        error
}

func initialModel() model {
	te := texteditor.New()
	te.Focus()
	te.SetSyntax("sql")
	te.SetValue(`select *
from tab1
where id=120 and f='breaking line test'
`)
	te.SetSize(width, height)
	style.Width(width).Height(height)
	return model{
		texteditor: te,
		err:        nil,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if m.texteditor.Focused() {
				m.texteditor.Blur()
			}
		case tea.KeyCtrlC:
			return m, tea.Quit
		default:
			if !m.texteditor.Focused() {
				cmd = m.texteditor.Focus()
				cmds = append(cmds, cmd)
			}
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.texteditor, cmd = m.texteditor.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {

	return fmt.Sprintf(
		"Type some text...\n\n%s\n\n%s",
		style.Render(m.texteditor.View()),
		"(ctrl+c to quit)",
	) + "\n\n"
}
