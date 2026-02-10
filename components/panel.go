package components

import (
	"github.com/mecha/journal/theme"

	t "github.com/gdamore/tcell/v2"
)

type Panel struct {
	title   string
	content Component
}

func NewPanel(title string, content Component) *Panel {
	return &Panel{title, content}
}

func (p *Panel) SetTitle(title string) *Panel {
	p.title = title
	return p
}

func (p *Panel) HandleEvent(ev t.Event) bool {
	return p.content.HandleEvent(ev)
}

func (p *Panel) Render(r Renderer, hasFocus bool) {
	w, h := r.Size()
	style := theme.Borders(hasFocus)

	DrawBox(r, 0, 0, w, h, BordersRound, style)

	if len(p.title) > 0 {
		r.PutStrStyled(2, 0, p.title, style)
	}

	region := r.SubRegion(Rect{Pos{1, 1}, Size{w - 2, h - 2}})
	p.content.Render(region, hasFocus)
}

func DrawPanel(r Renderer, title string, style t.Style) Renderer {
	w, h := r.Size()

	DrawBox(r, 0, 0, w, h, BordersRound, style)

	if len(title) > 0 {
		r.PutStrStyled(2, 0, title, style)
	}

	return r.SubRegion(Rect{Pos{1, 1}, Size{w - 2, h - 2}})
}
