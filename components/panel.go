package components

import (
	"journal-tui/theme"

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

	p.content.Render(RegionFrom(r, Rect{Pos{1, 1}, Size{w - 2, h - 2}}), hasFocus)
}

func DrawPanel(r Renderer, title string, style t.Style, hasFocus bool) *Region {
	w, h := r.Size()
	borderStyle := theme.Borders(hasFocus, style)

	DrawBox(r, 0, 0, w, h, BordersRound, borderStyle)

	if len(title) > 0 {
		r.PutStrStyled(2, 0, title, borderStyle)
	}

	return RegionFrom(r, Rect{Pos{1, 1}, Size{w - 2, h - 2}})
}
