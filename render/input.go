package render

import (
	"strings"

	t "github.com/gdamore/tcell/v2"
)

func InputBox(screen t.Screen, title, content string, x, y, width int) {
	Box(screen, x, y, width, 3, RoundedBorders, t.StyleDefault)

	screen.PutStrStyled(x+2, y, title, t.StyleDefault)

	if len(content) > width-4 {
		content = content[:len(content)-width-4]
	}

	screen.PutStrStyled(x+2, y+1, content, t.StyleDefault)
	screen.ShowCursor(x+2+len(content), y+1)
}

func PasswordBox(screen t.Screen, title string, length, x, y, width int) {
	numStars := max(0, min(width-4, length))
	stars := strings.Repeat("*", numStars)

	InputBox(screen, title, stars, x, y, width)
}
