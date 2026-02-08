package components

import (
	"journal-tui/theme"
	"strings"

	t "github.com/gdamore/tcell/v2"
)

// A subset of the tcell.Screen interface, for just the rendering stuff
type Renderer interface {
	// Fill fills the screen with the given character and style.
	// The effect of filling the screen is not visible until Show
	// is called (or Sync).
	Fill(rune, t.Style)

	// Put writes the first grapheme of the given string with the
	// given style at the given coordinates. (Only the first grapheme
	// occupying either one or two cells is stored.) It returns the
	// remainder of the string, and the width displayed.
	Put(x int, y int, str string, style t.Style) (string, int)

	// PutStr writes a string starting at the given position, using the
	// default style. The content is clipped to the screen dimensions.
	PutStr(x int, y int, str string)

	// PutStrStyled writes a string starting at the given position, using
	// the given style. The content is clipped to the screen dimensions.
	PutStrStyled(x int, y int, str string, style t.Style)

	// ShowCursor is used to display the cursor at a given location.
	// If the coordinates -1, -1 are given or are otherwise outside the
	// dimensions of the screen, the cursor will be hidden.
	ShowCursor(x int, y int)

	// HideCursor is used to hide the cursor.  It's an alias for
	// ShowCursor(-1, -1).sim
	HideCursor()

	// SetCursorStyle is used to set the cursor style.  If the style
	// is not supported (or cursor styles are not supported at all),
	// then this will have no effect.  Color will be changed if supplied,
	// and the terminal supports doing so.
	SetCursorStyle(t.CursorStyle, ...t.Color)

	// Size returns the screen size as width, height.  This changes in
	// response to a call to Clear or Flush.
	Size() (width, height int)
}

type Pos struct{ X, Y int }

func (p Pos) XY() (int, int)    { return p.X, p.Y }
func (p Pos) Pos() (int, int)   { return p.X, p.Y }
func (p Pos) Add(x, y int) Pos  { return Pos{p.X + x, p.Y + y} }
func (p Pos) AddPos(p2 Pos) Pos { return p.Add(p2.X, p2.Y) }

type Size struct{ W, H int }

func (s Size) WH() (int, int)       { return s.W, s.H }
func (s Size) Size() (int, int)     { return s.W, s.H }
func (s Size) Add(w, h int) Size    { return Size{s.W + w, s.H + h} }
func (s Size) AddSize(s2 Size) Size { return s.Add(s2.W, s2.H) }

type Rect struct {
	Pos
	Size
}

func NewRect(x, y, w, h int) Rect {
	return Rect{Pos{x, y}, Size{w, h}}
}

func (r Rect) XYWH() (int, int, int, int) {
	x, y := r.Pos.XY()
	w, h := r.Size.WH()
	return x, y, w, h
}

func CenterRect(rect Rect, w, h int) Rect {
	return Rect{
		Pos{
			rect.X + (rect.W-w)/2,
			rect.Y + (rect.H-h)/2,
		},
		Size{w, h},
	}
}

var _ Renderer = (*Region)(nil)

// A renderer that renders to a rectangular region of another renderer.
type Region struct {
	Renderer
	Rect
}

func RegionFrom(parent Renderer, rect Rect) *Region {
	return &Region{parent, rect}
}

func CenteredRegion(parent Renderer, w, h int) *Region {
	pw, ph := parent.Size()
	rect := CenterRect(Rect{Pos{0, 0}, Size{pw, ph}}, w, h)
	return &Region{parent, rect}
}

func (r *Region) Fill(rune rune, style t.Style) {
	for dy := range r.Rect.H {
		r.Renderer.PutStrStyled(r.X, r.Y+dy, strings.Repeat(string(rune), r.Rect.W), style)
	}
}

func (r *Region) Put(x int, y int, str string, style t.Style) (string, int) {
	return r.Renderer.Put(r.Rect.X+x, r.Rect.Y+y, str, style)
}

func (r *Region) PutStr(x int, y int, str string) {
	r.Renderer.PutStr(r.Rect.X+x, r.Rect.Y+y, str)
}

func (r *Region) PutStrStyled(x int, y int, str string, style t.Style) {
	r.Renderer.PutStrStyled(r.Rect.X+x, r.Rect.Y+y, str, style)
}

func (r *Region) ShowCursor(x int, y int) {
	r.Renderer.ShowCursor(r.X+x, r.Y+y)
}

func (r *Region) HideCursor() {
	r.Renderer.HideCursor()
}

func (r *Region) SetCursorStyle(style t.CursorStyle, color ...t.Color) {
	r.Renderer.SetCursorStyle(style, color...)
}

func (r *Region) Size() (width, height int) {
	return r.Rect.WH()
}

type BorderSet struct {
	LR, TB, RB, LB, TR, TL, LRB, TLR, TLB, TRB, TLRB string
}

var BordersRound = BorderSet{
	LR:   "─",
	TB:   "│",
	RB:   "╭",
	LB:   "╮",
	TR:   "╰",
	TL:   "╯",
	LRB:  "┬",
	TLR:  "┴",
	TRB:  "├",
	TLB:  "┤",
	TLRB: "┼",
}

var BordersSquare = BorderSet{
	LR:   "─",
	TB:   "│",
	RB:   "┌",
	LB:   "┐",
	TR:   "└",
	TL:   "┘",
	LRB:  "┬",
	TLR:  "┴",
	TRB:  "├",
	TLB:  "┤",
	TLRB: "┼",
}

func DrawBox(r Renderer, x, y, w, h int, borders BorderSet, style t.Style) {
	x2, y2 := x+w-1, y+h-1

	r.Put(x, y, borders.RB, style)
	r.Put(x2, y, borders.LB, style)
	r.Put(x, y2, borders.TR, style)
	r.Put(x2, y2, borders.TL, style)

	for i := range w - 2 {
		r.Put(x+i+1, y, borders.LR, style)
		r.Put(x+i+1, y2, borders.LR, style)
	}

	for i := range h - 2 {
		r.Put(x, y+i+1, borders.TB, style)
		r.Put(x2, y+i+1, borders.TB, style)
	}
}

func DrawButton(r Renderer, x, y int, text string, underline rune, hasFocus bool) int {
	btnStyle := theme.Button(hasFocus)
	fullText := "  " + text + "  "

	r.PutStrStyled(x, y, fullText, btnStyle)

	if underline > 0 {
		i := strings.IndexRune(text, underline)
		if i >= 0 {
			r.PutStrStyled(x+2+i, y, string(underline), btnStyle.Underline(true))
		}
	}

	return len(fullText)
}
