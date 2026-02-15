package theme

import t "github.com/gdamore/tcell/v2"

var (
	Logo = func(s ...t.Style) t.Style {
		return extend(s).Foreground(t.ColorGreen)
	}
	Input = func(s ...t.Style) t.Style {
		return extend(s)
	}
	Dialog = func(s ...t.Style) t.Style {
		return extend(s)
	}
	BordersNormal = func(s ...t.Style) t.Style {
		return extend(s)
	}
	BordersFocus = func(s ...t.Style) t.Style {
		return extend(s).Bold(true).Foreground(t.ColorGreen)
	}
	ListNormal = func(s ...t.Style) t.Style {
		return extend(s)
	}
	ListSelect = func(s ...t.Style) t.Style {
		return extend(s).Bold(true).Foreground(t.ColorBlack).Background(t.ColorBlue)
	}
	ButtonNormal = func(s ...t.Style) t.Style {
		return extend(s).Bold(true)
	}
	ButtonFocus = func(s ...t.Style) t.Style {
		return ButtonNormal(s...).Foreground(t.ColorBlack).Background(t.ColorGreen)
	}
	CalendarDay = func(s ...t.Style) t.Style {
		return extend(s)
	}
	CalendarSelect = func(s ...t.Style) t.Style {
		return extend(s).Foreground(t.ColorBlack).Background(t.ColorBlue)
	}
	CalendarToday = func(s ...t.Style) t.Style {
		return extend(s).Bold(true).Foreground(t.ColorGold)
	}
	CalendarOutside = func(s ...t.Style) t.Style {
		return extend(s).Foreground(t.ColorDimGray)
	}
	Help = func(s ...t.Style) t.Style {
		return extend(s).Foreground(t.ColorAqua)
	}
)

func extend(base []t.Style) t.Style {
	if len(base) > 0 {
		return base[0]
	}
	return t.StyleDefault
}

func Borders(hasFocus bool, base ...t.Style) t.Style {
	if hasFocus {
		return BordersFocus(base...)
	} else {
		return BordersNormal(base...)
	}
}

func Button(hasFocus bool, base ...t.Style) t.Style {
	if hasFocus {
		return ButtonFocus(base...)
	} else {
		return ButtonNormal(base...)
	}
}

func ListItem(isSelected bool, base ...t.Style) t.Style {
	if isSelected {
		return ListSelect(base...)
	} else {
		return ListNormal(base...)
	}
}
