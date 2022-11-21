package texteditor

import (
	"fmt"
	"io"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/lipgloss"
)

var cursorStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("1")).
	Bold(true)

// A FormatterFunc is a Formatter implemented as a function.
//
// Guards against iterator panics.
type FormatterFunc func(w io.Writer, style *chroma.Style, iterator chroma.Iterator, hasCursor bool, cursorColumn int) error

func (f FormatterFunc) Format(w io.Writer, s *chroma.Style, it chroma.Iterator, hasCursor bool, cursorColumn int) (err error) { // nolint
	defer func() {
		if perr := recover(); perr != nil {
			err = perr.(error)
		}
	}()
	return f(w, s, it, hasCursor, cursorColumn)
}

var customFormatter = FormatterFunc(customFormatterFunc)

// Clear the background colour.
func clearBackground(style *chroma.Style) *chroma.Style {
	builder := style.Builder()
	bg := builder.Get(chroma.Background)
	bg.Background = 0
	bg.NoInherit = true
	builder.AddEntry(chroma.Background, bg)
	style, _ = builder.Build()
	return style
}

func applyStyle(entry chroma.StyleEntry) string {
	out := ""

	if !entry.IsZero() {
		if entry.Bold == chroma.Yes {
			out += "\033[1m"
		}
		if entry.Underline == chroma.Yes {
			out += "\033[4m"
		}
		if entry.Italic == chroma.Yes {
			out += "\033[3m"
		}
		if entry.Colour.IsSet() {
			out += fmt.Sprintf("\033[38;2;%d;%d;%dm", entry.Colour.Red(), entry.Colour.Green(), entry.Colour.Blue())
		}
		if entry.Background.IsSet() {
			out += fmt.Sprintf("\033[48;2;%d;%d;%dm", entry.Background.Red(), entry.Background.Green(), entry.Background.Blue())
		}
	}
	return out
}

func customFormatterFunc(w io.Writer, style *chroma.Style, it chroma.Iterator, hasCursor bool, cursorColumn int) error {
	style = clearBackground(style)

	column := 0
	doneWithCursor := false
	for token := it(); token != chroma.EOF; token = it() {

		entry := style.Get(token.Type)

		fmt.Fprint(w, applyStyle(entry))

		if hasCursor && column+len(token.Value) > cursorColumn && !doneWithCursor {
			pos := cursorColumn - column
			tv := token.Value
			preCursor := tv[0:pos]
			cursor := tv[pos : pos+1]
			postCursor := tv[pos+1:]

			fmt.Fprint(w, preCursor)
			fmt.Fprint(w, cursorStyle.Render(cursor))

			// reapply style resetted from cursor
			fmt.Fprint(w, applyStyle(entry))
			fmt.Fprint(w, postCursor)
			doneWithCursor = true
		} else {
			fmt.Fprint(w, token.Value)
		}

		if !entry.IsZero() {
			fmt.Fprint(w, "\033[0m")
		}

		column += len(token.Value)
	}
	return nil
}

// highlight some text.
// Lexer, formatter and style may be empty, in which case a best-effort is made.
func highlight(w io.Writer, source, lexer, style string, hasCursor bool, cursorColumn int) error {
	// Determine lexer.
	l := lexers.Get(lexer)
	if l == nil {
		l = lexers.Analyse(source)
	}
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)

	f := customFormatter

	// Determine style.
	s := styles.Get(style)
	if s == nil {
		s = styles.Fallback
	}

	it, err := l.Tokenise(nil, source)
	if err != nil {
		return err
	}
	return f.Format(w, s, it, hasCursor, cursorColumn)
}
