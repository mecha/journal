package components

import (
	t "github.com/gdamore/tcell/v2"
)

type PanelProps struct {
	Title    string
	Style    t.Style
	Borders  BorderSet
	Children Children
}

func Panel(r Renderer, props PanelProps) EventHandler {
	w, h := r.Size()

	DrawBox(r, 0, 0, w, h, props.Borders, props.Style)

	if len(props.Title) > 0 {
		r.PutStrStyled(2, 0, props.Title, props.Style)
	}

	if props.Children == nil {
		return nil
	}

	inside := r.SubRegion(Rect{Pos{1, 1}, Size{w - 2, h - 2}})
	return props.Children(inside)
}
