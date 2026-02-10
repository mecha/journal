package components

import (
	"strings"

	"github.com/mecha/journal/theme"
	"github.com/mecha/journal/utils"

	t "github.com/gdamore/tcell/v2"
)

type Confirm struct {
	message       string
	yesButton     string
	noButton      string
	onChoiceFunc  func(accepted bool)
	buttonFocused int
}

var _ Component = (*Confirm)(nil)

func NewConfirm(message string, onChoice func(accepted bool)) *Confirm {
	return &Confirm{
		message:       message,
		onChoiceFunc:  onChoice,
		yesButton:     "Yes",
		noButton:      "No",
		buttonFocused: 1,
	}
}

func (c *Confirm) YesButton(text string) {
	c.yesButton = text
}

func (c *Confirm) NoButton(text string) {
	c.noButton = text
}

func (c *Confirm) HandleEvent(ev t.Event) bool {
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

func (c *Confirm) Render(r Renderer, hasFocus bool) {
	bw, bh := r.Size()

	minWidth := len(c.yesButton) + len(c.noButton) + 2
	width := min(bw, max(40, minWidth))

	lines := utils.WrapString(c.message, width-2)
	height := 3 + len(lines)

	x, y := (bw-width)/2, (bh-height)/2

	region := CenteredRegion(r, width, height)

	region.Fill(' ', theme.Dialog())
	DrawBox(r, x, y, width, height, BordersRound, theme.Borders(hasFocus, theme.Dialog()))

	for i, line := range lines {
		r.PutStr(x+1, y+1+i, line)
	}

	right := x + width - 1
	buttonY := y + height - 2

	noBtnX := right - 2 - len(c.noButton) - 4
	yesBtnX := noBtnX - 1 - len(c.yesButton) - 4

	DrawButton(r, noBtnX, buttonY, c.noButton, rune(c.noButton[0]), c.buttonFocused%2 == 1)
	DrawButton(r, yesBtnX, buttonY, c.yesButton, rune(c.yesButton[0]), c.buttonFocused%2 == 0)
}
