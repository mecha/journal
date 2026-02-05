package render

import (
	"journal-tui/theme"

	t "github.com/gdamore/tcell/v2"
)

func Panel(screen t.Screen, title string, x, y, w, h int, borders BorderMap, hasFocus bool) {
	style := theme.BorderStyle(hasFocus)
	Box(screen, x, y, w, h, borders, style)
	screen.PutStrStyled(x+2, y, title, style)
}
