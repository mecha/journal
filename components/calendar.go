package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/mecha/journal/theme"

	t "github.com/gdamore/tcell/v2"
)

var calenderHeaders = []string{
	"Mon",
	"Tue",
	"Wed",
	"Thu",
	"Fri",
	"Sat",
	"Sun",
}

type CalendarProps struct {
	BorderStyle   t.Style
	Selected      time.Time
	UnderlineDays func(time.Time) bool
	OnSelectDay   func(time.Time)
}

func Calendar(renderer Renderer, props CalendarProps) func(t.Event) bool {
	const (
		numCols      = 7
		numRows      = 7
		colWidth     = 5
		rowHeight    = 1
		headerHeight = 2
	)

	w, _ := renderer.Size()
	b := BordersRound
	renderer.PutStrStyled(-1, 1, b.TRB+strings.Repeat(b.LR, w)+b.TLB, props.BorderStyle)

	for i, header := range calenderHeaders {
		if len(header) < colWidth {
			renderer.PutStr((i*6)+1, 0, header)
		}
	}

	year, month, day := props.Selected.Date()
	monthStart := time.Date(year, month, 1, 12, 0, 0, 0, time.Local)
	numDays := monthStart.AddDate(0, 1, -1).Day()
	firstIdx := (int(monthStart.Weekday()) + 6) % 7
	lastIdx := firstIdx + numDays
	cursor := firstIdx + day - 1

	today := time.Now()
	start := props.Selected.AddDate(0, 0, -cursor)

	for idx := range 42 {
		row, col := idx/7, idx%7
		date := start.AddDate(0, 0, idx)

		dayStyle := theme.CalendarDay()
		if idx < firstIdx || idx > lastIdx {
			dayStyle = theme.CalendarOutside(dayStyle)
		}
		if idx == cursor {
			dayStyle = theme.CalendarSelect(dayStyle)
		}
		year, month, day := date.Date()
		if day == today.Day() && month == today.Month() && year == today.Year() {
			dayStyle = theme.CalendarToday(dayStyle)
		}

		x := 1 + (col * (colWidth + 1))
		y := headerHeight + (row * (rowHeight + 1))

		renderer.PutStrStyled(x, y, "    ", dayStyle)

		if props.UnderlineDays != nil && props.UnderlineDays(date) {
			dayStyle = dayStyle.Underline(true)
		}
		renderer.PutStrStyled(x+1, y, fmt.Sprintf("%02d", day), dayStyle)
	}

	return func(ev t.Event) bool {
		switch ev := ev.(type) {
		case *t.EventKey:
			switch ev.Key() {
			default:
				return false
			case t.KeyUp:
				props.OnSelectDay(props.Selected.AddDate(0, 0, -7))
			case t.KeyDown:
				props.OnSelectDay(props.Selected.AddDate(0, 0, 7))
			case t.KeyLeft:
				props.OnSelectDay(props.Selected.AddDate(0, 0, -1))
			case t.KeyRight:
				props.OnSelectDay(props.Selected.AddDate(0, 0, 1))

			case t.KeyRune:
				switch ev.Rune() {
				default:
					return false
				case 't':
					props.OnSelectDay(time.Now())
				case 'n':
					props.OnSelectDay(props.Selected.AddDate(0, 1, 0))
				case 'p':
					props.OnSelectDay(props.Selected.AddDate(0, -1, 0))
				case 'j':
					props.OnSelectDay(props.Selected.AddDate(0, 0, 7))
				case 'k':
					props.OnSelectDay(props.Selected.AddDate(0, 0, -7))
				case 'h':
					props.OnSelectDay(props.Selected.AddDate(0, 0, -1))
				case 'l':
					props.OnSelectDay(props.Selected.AddDate(0, 0, 1))
				}
			}
		}

		return true
	}
}
