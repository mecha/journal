package components

import (
	"journal-tui/render"
	"journal-tui/theme"

	t "github.com/gdamore/tcell/v2"
)

type Panel struct {
	title   string
	content Component
	style   t.Style
}

func NewPanel(title string, content Component) *Panel {
	return &Panel{title, content, t.StyleDefault}
}

func (p *Panel) Style(style t.Style) *Panel {
	p.style = style
	return p
}

func (p *Panel) HandleEvent(ev t.Event) bool {
	return p.content.HandleEvent(ev)
}

func (p *Panel) Render(screen t.Screen, region Rect, hasFocus bool) {
	x, y, w, h := region.XYWH()
	style := theme.BorderStyle(hasFocus)

	render.Box(screen, x, y, w, h, render.RoundedBorders, style)

	if len(p.title) > 0 {
		screen.PutStrStyled(x+2, y, p.title, style)
	}

	p.content.Render(screen, NewRect(x+1, y+1, w-2, h-2), hasFocus)
}
