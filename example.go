package main

import (
	"fmt"
	"log"

	"github.com/alecthomas/chroma/v2"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ferama/bubble-texteditor/texteditor"
	"github.com/ferama/bubble-texteditor/texteditor/selector"
)

var style = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder())

const (
	width  = 40
	height = 10
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

func intellisense(t chroma.Token) []selector.IntellisenseItem {
	resEmpty := make([]selector.IntellisenseItem, 0)
	resFull := append(resEmpty,
		selector.IntellisenseItem{
			Value: "table1",
			Kind:  "table",
		},
		selector.IntellisenseItem{
			Value: "table2",
			Kind:  "table",
		},
		selector.IntellisenseItem{
			Value: "table3",
			Kind:  "table",
		},
		selector.IntellisenseItem{
			Value: "table4",
			Kind:  "table",
		},
		selector.IntellisenseItem{
			Value: "table5",
			Kind:  "table",
		},
	)

	var res []selector.IntellisenseItem

	switch t.Type {
	case chroma.Keyword:
		switch t.Value {
		case "select":
			res = resFull
		default:
			res = resEmpty
		}
	default:
		res = resEmpty
	}

	fmt.Printf("\033[2K\r%s: %s", t.Type, t.Value)
	return res
}

func initialModel() model {
	te := texteditor.New()
	te.Focus()
	te.SetSyntax("sql")
	// te.SetValue(`selec`)
	te.SetValue(`select *
from tab1
where id=120 and f='èasd'`)
	te.SetSize(width, height)
	te.Intellisense = intellisense
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
		"\n%s\n\n%s",
		style.Render(m.texteditor.View()),
		"(ctrl+c to quit)",
	) + "\n\n"
}
