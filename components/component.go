package components

import t "github.com/gdamore/tcell/v2"

type Component interface {
	HandleEvent(ev t.Event) (consume bool)
	Render(screen t.Screen, hasFocus bool)
}

type ComponentPos struct{ x, y int }

func (p *ComponentPos) Pos() (int, int) { return p.x, p.y }
func (p *ComponentPos) Move(x, y int)   { p.x, p.y = x, y }

type ComponentSize struct{ w, h int }

func (p *ComponentSize) Size() (int, int) { return p.w, p.h }
func (p *ComponentSize) Resize(w, h int)  { p.w, p.h = w, h }
