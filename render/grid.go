package render

import (
	"strings"

	t "github.com/gdamore/tcell/v2"
)

func Grid(screen t.Screen, x, y, cols, rows, colWidth, rowHeight int, borders BorderMap, style t.Style) {
	width, height := 1+cols*(1+colWidth), 1+rows*(1+rowHeight)

	hLine := strings.Repeat(borders.LR, colWidth)

	screen.PutStrStyled(x, y, borders.RB, style)
	screen.PutStrStyled(x, y+height-1, borders.TR, style)

	for c := range cols {
		cx := x + 1 + (colWidth+1)*c
		if c > 0 {
			screen.PutStrStyled(cx-1, y, borders.LRB, style)
			screen.PutStrStyled(cx-1, y+height-1, borders.TLR, style)
		}
		screen.PutStrStyled(cx, y, hLine, style)
		screen.PutStrStyled(cx, y+height-1, hLine, style)

		for r := range rows {
			ry := y + (rowHeight+1)*r
			screen.PutStrStyled(cx, ry, hLine+borders.TLRB, style)

			if c > 0 {
				for i := range rowHeight {
					screen.PutStrStyled(cx-1, ry+i+1, borders.TB, style)
				}
			}
		}
	}

	screen.PutStrStyled(x+width-1, y, borders.LB, style)
	screen.PutStrStyled(x+width-1, y+height-1, borders.TL, style)

	for r := range rows {
		ry := y + 1 + (rowHeight+1)*r
		if r > 0 {
			screen.PutStrStyled(x, ry-1, borders.TRB, style)
			screen.PutStrStyled(x+width-1, ry-1, borders.TLB, style)
		}
		for i := range rowHeight {
			screen.PutStrStyled(x, ry+i, borders.TB, style)
			screen.PutStrStyled(x+width-1, ry+i, borders.TB, style)
		}
	}
}
