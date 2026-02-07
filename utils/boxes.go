package utils

import (
	"strings"

	t "github.com/gdamore/tcell/v2"
)

func Box(screen t.Screen, x, y, w, h int, borders BorderSet, style t.Style) {
	x2, y2 := x+w-1, y+h-1

	screen.Put(x, y, borders.RB, style)
	screen.Put(x2, y, borders.LB, style)
	screen.Put(x, y2, borders.TR, style)
	screen.Put(x2, y2, borders.TL, style)

	for i := range w - 2 {
		screen.Put(x+i+1, y, borders.LR, style)
		screen.Put(x+i+1, y2, borders.LR, style)
	}

	for i := range h - 2 {
		screen.Put(x, y+i+1, borders.TB, style)
		screen.Put(x2, y+i+1, borders.TB, style)
	}
}

func BoxHorizontalDivider(screen t.Screen, x, y, w int, borders BorderSet, style t.Style) {
	screen.PutStrStyled(x, y, borders.TRB+strings.Repeat(borders.LR, w-2)+borders.TLB, style)
}

func BoxVerticalDivider(screen t.Screen, x, y, h int, borders BorderSet, style t.Style) {
	screen.PutStrStyled(x, y, borders.LRB, style)
	for i := range h - 2 {
		screen.PutStrStyled(x, y+i+1, borders.TB, style)
	}
	screen.PutStrStyled(x, y+h-1, borders.TLR, style)
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
