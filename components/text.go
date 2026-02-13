package components

import (
	"strings"

	"github.com/mecha/journal/utils"

	t "github.com/gdamore/tcell/v2"
)

var _ Component = (*Text)(nil)

type Text struct {
	scroll   Pos
	lines    []string
	lastSize Size
	maxLen   int
	style    t.Style
}

func NewText(lines []string) *Text {
	s := &Text{}
	s.SetLines(lines)
	return s
}

func (s *Text) SetLines(lines []string) *Text {
	s.lines = lines
	s.maxLen = utils.MaxLength(lines)
	return s
}

func (s *Text) AddLines(lines []string) *Text {
	s.maxLen = max(s.maxLen, utils.MaxLength(lines))
	s.lines = append(s.lines, lines...)
	return s
}

func (s *Text) SetStyle(style t.Style) *Text {
	s.style = style
	return s
}

func (s *Text) SetScrollPos(pos Pos) *Text {
	s.scroll.X = max(0, min(s.maxLen-s.lastSize.W, pos.X))
	s.scroll.Y = max(0, min(len(s.lines)-s.lastSize.H, pos.Y))
	return s
}

func (s *Text) Writer() *TextWriter {
	return &TextWriter{s, nil}
}

func (s *Text) ScrollBy(x, y int) { s.SetScrollPos(s.scroll.Add(x, y)) }
func (s *Text) ScrollUp(n int)    { s.ScrollBy(0, -n) }
func (s *Text) ScrollDown(n int)  { s.ScrollBy(0, n) }
func (s *Text) ScrollLeft(n int)  { s.ScrollBy(-n, 0) }
func (s *Text) ScrollRight(n int) { s.ScrollBy(n, 0) }
func (s *Text) ScrollToBottom()   { s.SetScrollPos(Pos{s.scroll.X, len(s.lines)}) }

func (s *Text) HandleEvent(ev t.Event) bool {
	switch ev := ev.(type) {
	default:
		return false
	case *t.EventKey:
		switch ev.Key() {
		default:
			return false
		case t.KeyRune:
			switch ev.Rune() {
			default:
				return false
			case 'h':
				s.ScrollLeft(1)
			case 'j':
				s.ScrollDown(1)
			case 'k':
				s.ScrollUp(1)
			case 'l':
				s.ScrollRight(1)
			}
		case t.KeyLeft:
			s.ScrollLeft(1)
		case t.KeyDown:
			s.ScrollDown(1)
		case t.KeyUp:
			s.ScrollUp(1)
		case t.KeyRight:
			s.ScrollRight(1)
		}
	}

	return true
}

func (s *Text) Render(r Renderer, hasFocus bool) {
	width, height := r.Size()
	s.lastSize = Size{width, height}
	s.SetScrollPos(s.scroll)

	topLine := max(0, s.scroll.Y)
	lastLine := min(len(s.lines), topLine+height)

	if topLine == lastLine {
		return
	}

	for i, line := range s.lines[topLine:lastLine] {
		if len(line) == 0 {
			continue
		}
		left := max(0, s.scroll.X)
		right := min(len(line), s.scroll.X+width)
		row := utils.FixedString(line[left:right], width, " ")
		r.PutStrStyled(0, i, row, s.style)
	}
}

type TextWriter struct {
	component *Text
	callback  func()
}

func (w *TextWriter) Write(data []byte) (int, error) {
	newLines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	w.component.AddLines(newLines)
	w.component.ScrollToBottom()
	if w.callback != nil {
		w.callback()
	}
	return len(data), nil
}

func (w *TextWriter) OnWrite(callback func()) {
	w.callback = callback
}

type TextState struct {
	Scroll Pos
	lines  []string
	maxLen int
}

func (s *TextState) AddLines(lines []string) {
	s.lines = append(s.lines, lines...)
	s.maxLen = max(s.maxLen, utils.MaxLength(lines))
}

func (s *TextState) SetLines(lines []string) {
	s.lines = lines
	s.maxLen = utils.MaxLength(lines)
}

func (s *TextState) ScrollToBottom() {
	s.Scroll = Pos{s.Scroll.X, len(s.lines)}
}

type TextProps struct {
	Style t.Style
}

func DrawText(r Renderer, state *TextState, props TextProps) EventHandler {
	width, height := r.Size()
	setScroll := func(pos Pos) {
		state.Scroll.X = max(0, min(state.maxLen-width, pos.X))
		state.Scroll.Y = max(0, min(len(state.lines)-height, pos.Y))
	}
	setScroll(state.Scroll)

	topLine := max(0, state.Scroll.Y)
	lastLine := min(len(state.lines), topLine+height)

	for i, line := range state.lines[topLine:lastLine] {
		if len(line) == 0 {
			continue
		}
		left := max(0, state.Scroll.X)
		right := min(len(line), state.Scroll.X+width)
		row := utils.FixedString(line[left:right], width, " ")
		r.PutStrStyled(0, i, row, props.Style)
	}

	return func(ev t.Event) bool {
		switch ev := ev.(type) {
		default:
			return false
		case *t.EventKey:
			switch ev.Key() {
			default:
				return false
			case t.KeyRune:
				switch ev.Rune() {
				default:
					return false
				case 'h':
					setScroll(state.Scroll.Add(-1, 0))
				case 'j':
					setScroll(state.Scroll.Add(0, 1))
				case 'k':
					setScroll(state.Scroll.Add(0, -1))
				case 'l':
					setScroll(state.Scroll.Add(1, 0))
				case ',':
					setScroll(state.Scroll.Add(0, -10))
				case '.':
					setScroll(state.Scroll.Add(0, 10))
				}
			case t.KeyLeft:
				setScroll(state.Scroll.Add(-1, 0))
			case t.KeyDown:
				setScroll(state.Scroll.Add(0, 1))
			case t.KeyUp:
				setScroll(state.Scroll.Add(0, -1))
			case t.KeyRight:
				setScroll(state.Scroll.Add(1, 0))
			}
		}

		return true
	}
}
