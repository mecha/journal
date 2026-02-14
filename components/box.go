package components

import t "github.com/gdamore/tcell/v2"

type BoxProps struct {
	Title    string
	Borders  BorderSet
	Style    t.Style
	Children Component
}

func Box(r Renderer, props BoxProps) EventHandler {
	w, h := r.Size()
	x2, y2 := w-1, h-1

	r.Put(0, 0, props.Borders.RB, props.Style)
	r.Put(x2, 0, props.Borders.LB, props.Style)
	r.Put(0, y2, props.Borders.TR, props.Style)
	r.Put(x2, y2, props.Borders.TL, props.Style)

	for i := range w - 2 {
		r.Put(i+1, 0, props.Borders.LR, props.Style)
		r.Put(i+1, y2, props.Borders.LR, props.Style)
	}

	for i := range h - 2 {
		r.Put(0, i+1, props.Borders.TB, props.Style)
		r.Put(x2, i+1, props.Borders.TB, props.Style)
	}

	if len(props.Title) > 0 {
		r.PutStrStyled(2, 0, props.Title, props.Style)
	}

	if props.Children != nil {
		inside := r.SubRegion(Rect{Pos{1, 1}, Size{w - 2, h - 2}})
		return props.Children(inside)
	}

	return nil
}

type BorderSet struct{ LR, TB, RB, LB, TR, TL, LRB, TLR, TLB, TRB, TLRB string }

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
