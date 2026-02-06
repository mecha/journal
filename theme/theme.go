package theme

import t "github.com/gdamore/tcell/v2"

var (
	Default         = t.StyleDefault
	Border          = Default
	BorderFocus     = Border.Bold(true).Foreground(t.ColorGreen)
	ListSelect      = Default.Bold(true).Foreground(t.ColorBlack).Background(t.ColorBlue)
	Button          = Default.Bold(true)
	ButtonFocus     = Button.Foreground(t.ColorBlack).Background(t.ColorGreen)
	CalendarBorder  = Border.Foreground(t.ColorDimGray)
	CalendarDay     = Default
	CalendarHeader  = CalendarDay
	CalendarSelect  = Default.Bold(true).Foreground(t.ColorBlack).Background(t.ColorBlue)
	CalendarToday   = Default.Bold(true).Foreground(t.ColorYellow)
	CalendarOutside = Default.Foreground(t.ColorDimGray)
	CalendarDot     = CalendarDay.Foreground(t.ColorGreen)
	Help            = Default.Foreground(t.ColorAqua)
)

func BorderStyle(hasFocus bool) t.Style {
	if hasFocus {
		return BorderFocus
	} else {
		return Border
	}
}

func ButtonStyle(hasFocus bool) t.Style {
	if hasFocus {
		return ButtonFocus
	} else {
		return Button
	}
}
