package components

import t "github.com/gdamore/tcell/v2"

type Component interface {
	HandleEvent(ev t.Event) (consumed bool)
	Render(screen t.Screen, bounds Rect, hasFocus bool)
}

type RenderFunc func(screen t.Screen, bounds Rect, hasFocus bool)
type EventHandlerFunc func(ev t.Event) (consumed bool)

type Pos struct{ X, Y int }

func NewPos(x, y int) Pos { return Pos{x, y} }

func (p Pos) XY() (int, int)    { return p.X, p.Y }
func (p Pos) Pos() (int, int)   { return p.X, p.Y }
func (p Pos) Add(x, y int) Pos  { return Pos{p.X + x, p.Y + y} }
func (p Pos) AddPos(p2 Pos) Pos { return p.Add(p2.X, p2.Y) }

type Size struct{ W, H int }

func NewSize(w, h int) Size { return Size{w, h} }

func (s Size) WH() (int, int)      { return s.W, s.H }
func (s Size) Size() (int, int)    { return s.W, s.H }
func (s Size) Add(w, h int) Size   { return Size{s.W + w, s.H + h} }
func (s Size) AddPos(s2 Size) Size { return s.Add(s2.W, s2.H) }

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

func CenterFit(rect Rect, size Size) Rect {
	return Rect{
		Pos{rect.X + (rect.W-size.W)/2, rect.Y + (rect.H-size.H)/2},
		Size{size.W, size.H},
	}
}
