package components

import (
	"journal-tui/render"
	"journal-tui/theme"
	"strings"

	t "github.com/gdamore/tcell/v2"
)

type ConfirmComponent struct {
	message       string
	yesButton     string
	noButton      string
	onChoiceFunc  func(accepted bool)
	buttonFocused int
}

var _ Component = (*ConfirmComponent)(nil)

func NewConfirmComponent(message string, onChoice func(accepted bool)) *ConfirmComponent {
	return &ConfirmComponent{
		message:       message,
		onChoiceFunc:  onChoice,
		yesButton:     "Yes",
		noButton:      "No",
		buttonFocused: 1,
	}
}

func (c *ConfirmComponent) YesButton(text string) {
	c.yesButton = text
}

func (c *ConfirmComponent) NoButton(text string) {
	c.noButton = text
}

func (c *ConfirmComponent) HandleEvent(ev t.Event) bool {
	yes, no := strings.ToLower(c.yesButton), strings.ToLower(c.noButton)

	switch ev := ev.(type) {
	default:
		return false
	case *t.EventKey:
		switch ev.Key() {
		default:
			return false
		case t.KeyRune:
			switch ev.Rune() {
			default:
				return false
			case rune(yes[0]):
				c.onChoiceFunc(true)
			case rune(no[0]):
				c.onChoiceFunc(false)
			}
		case t.KeyEsc:
			c.onChoiceFunc(false)
		case t.KeyEnter:
			c.onChoiceFunc(c.buttonFocused%2 == 0)
		case t.KeyLeft, t.KeyRight, t.KeyTab:
			c.buttonFocused = (c.buttonFocused + 1) % 2
		}
	}
	return true
}

func (c *ConfirmComponent) Render(screen t.Screen, hasFocus bool) {
	sw, sh := screen.Size()
	minWidth := len(c.yesButton) + len(c.noButton) + 2
	width := min(sw, max(40, minWidth))
	lines := render.WrapString(c.message, width-2)
	height := 3 + len(lines)
	x, y := (sw-width)/2, (sh-height)/2

	render.Box(screen, x, y, width, height, render.RoundedBorders, theme.BorderStyle(hasFocus))

	for i, line := range lines {
		screen.PutStr(x+1, y+1+i, line)
	}

	right := x + width - 1
	buttonY := y + height - 2

	noButtonStyle := theme.ButtonStyle(c.buttonFocused%2 == 1)
	noButtonText := "  " + c.noButton + "  "
	noButtonPos := right - 2 - len(noButtonText)

	yesButtonStyle := theme.ButtonStyle(c.buttonFocused%2 == 0)
	yesButtonText := "  " + c.yesButton + "  "
	yesButtonPos := noButtonPos - 1 - len(yesButtonText)

	screen.PutStrStyled(noButtonPos, buttonY, noButtonText, noButtonStyle)
	screen.PutStrStyled(noButtonPos+2, buttonY, c.noButton[:1], noButtonStyle.Underline(true))

	screen.PutStrStyled(yesButtonPos, buttonY, yesButtonText, yesButtonStyle)
	screen.PutStrStyled(yesButtonPos+2, buttonY, c.yesButton[:1], yesButtonStyle.Underline(true))
}
