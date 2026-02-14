package components

import (
	"strings"

	t "github.com/gdamore/tcell/v2"
)

// The simplest component is a function that outputs via a renderer and returns
// a function for handling events.
type Component func(r Renderer) EventHandler

// Similar to the tcell `Screen` interface, focused on rendering
type Renderer interface {
	// Creates a new renderer that renders inside a particular region
	// of the current renderer.
	SubRegion(rect Rect) Renderer

	// Gets the region that the renderer renders to.
	GetRegion() Rect

	// Gets the renderer for the entire screen.
	GetScreen() Renderer

	// Creates 2 new renderers that split the renderer's region horizontally.
	SplitHorizontal(x int) (Renderer, Renderer)

	// Creates 2 new renderers that split the renderer's region vertically.
	SplitVertical(y int) (Renderer, Renderer)

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

var _ Renderer = (*ScreenRenderer)(nil)
var _ Renderer = (*RegionRenderer)(nil)

// Adapter to make a tcell Screen a Renderer
type ScreenRenderer struct{ t.Screen }

func NewScreenRenderer(screen t.Screen) Renderer {
	return &ScreenRenderer{screen}
}

func (r *ScreenRenderer) SubRegion(rect Rect) Renderer {
	return &RegionRenderer{r, rect}
}

func (r *ScreenRenderer) GetRegion() Rect {
	w, h := r.Size()
	return Rect{Pos{0, 0}, Size{w, h}}
}

func (r *ScreenRenderer) GetScreen() Renderer {
	return r
}

func (r *ScreenRenderer) SplitHorizontal(x int) (Renderer, Renderer) {
	left, right := r.GetRegion().SplitHorizontal(x)
	return r.SubRegion(left), r.SubRegion(right)
}

func (r *ScreenRenderer) SplitVertical(y int) (Renderer, Renderer) {
	top, bottom := r.GetRegion().SplitVertical(y)
	return r.SubRegion(top), r.SubRegion(bottom)
}

// A renderer that renders to a rectangular region of another renderer.
type RegionRenderer struct {
	Renderer
	Rect
}

func CenteredRegion(parent Renderer, w, h int) *RegionRenderer {
	pw, ph := parent.Size()
	rect := CenterRect(Rect{Pos{0, 0}, Size{pw, ph}}, w, h)
	return &RegionRenderer{parent, rect}
}

func (r *RegionRenderer) SubRegion(rect Rect) Renderer {
	return &RegionRenderer{r, rect}
}

func (r *RegionRenderer) GetRegion() Rect {
	return r.Rect
}

func (r *RegionRenderer) GetScreen() Renderer {
	return r.Renderer.GetScreen()
}

func (r *RegionRenderer) SplitHorizontal(x int) (Renderer, Renderer) {
	left, right := r.Rect.SplitHorizontal(x)
	return r.SubRegion(left), r.SubRegion(right)
}

func (r *RegionRenderer) SplitVertical(y int) (Renderer, Renderer) {
	top, bottom := r.Rect.SplitVertical(y)
	return r.SubRegion(top), r.SubRegion(bottom)
}

func (r *RegionRenderer) Fill(rune rune, style t.Style) {
	for dy := range r.Rect.H {
		r.Renderer.PutStrStyled(r.X, r.Y+dy, strings.Repeat(string(rune), r.Rect.W), style)
	}
}

func (r *RegionRenderer) Put(x int, y int, str string, style t.Style) (string, int) {
	return r.Renderer.Put(r.Rect.X+x, r.Rect.Y+y, str, style)
}

func (r *RegionRenderer) PutStr(x int, y int, str string) {
	r.Renderer.PutStr(r.Rect.X+x, r.Rect.Y+y, str)
}

func (r *RegionRenderer) PutStrStyled(x int, y int, str string, style t.Style) {
	r.Renderer.PutStrStyled(r.Rect.X+x, r.Rect.Y+y, str, style)
}

func (r *RegionRenderer) ShowCursor(x int, y int) {
	r.Renderer.ShowCursor(r.X+x, r.Y+y)
}

func (r *RegionRenderer) HideCursor() {
	r.Renderer.HideCursor()
}

func (r *RegionRenderer) SetCursorStyle(style t.CursorStyle, color ...t.Color) {
	r.Renderer.SetCursorStyle(style, color...)
}

func (r *RegionRenderer) Size() (width, height int) {
	return r.Rect.WH()
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

func (r Rect) SplitHorizontal(x int) (Rect, Rect) {
	left := Rect{r.Pos, Size{x, r.Size.H}}
	right := Rect{Pos{r.Pos.X + x, r.Pos.Y}, Size{r.Size.W - x, r.Size.H}}
	return left, right
}

func (r Rect) SplitVertical(y int) (Rect, Rect) {
	top := Rect{r.Pos, Size{r.Size.W, y}}
	bottom := Rect{Pos{r.Pos.X, r.Pos.Y + y}, Size{r.Size.W, r.Size.H - y}}
	return top, bottom
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
