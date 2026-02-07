package components

import (
	"journal-tui/utils"

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

func NewTextScroller(lines []string) *Text {
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

func (s *Text) Render(screen t.Screen, region Rect, hasFocus bool) {
	s.lastSize = region.Size
	s.SetScrollPos(s.scroll)

	topLine := max(0, s.scroll.Y)
	lastLine := min(len(s.lines), topLine+region.H)

	if topLine == lastLine {
		return
	}

	for i, line := range s.lines[topLine:lastLine] {
		if len(line) == 0 {
			continue
		}
		left := max(0, s.scroll.X)
		right := min(len(line), s.scroll.X+region.W)
		row := utils.FixedString(line[left:right], region.W, " ")
		screen.PutStrStyled(region.X, region.Y+i, row, s.style)
	}
}
