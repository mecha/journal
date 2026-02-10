package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/mecha/journal/theme"

	t "github.com/gdamore/tcell/v2"
)

var _ Component = (*Calendar)(nil)

type Calendar struct {
	date            time.Time
	cursor          int
	firstIdx        int
	lastIdx         int
	underlineDay    func(time.Time) bool
	onDayChangeFunc func(time.Time)
}

func NewCalendar() *Calendar {
	c := &Calendar{}
	c.SetDate(dateOnly(time.Now()))
	return c
}

func (c *Calendar) UnderlineDay(fn func(time.Time) bool) *Calendar {
	c.underlineDay = fn
	return c
}

func (c *Calendar) OnDayChanged(callback func(time.Time)) *Calendar {
	c.onDayChangeFunc = callback
	return c
}

func (c *Calendar) Date() time.Time {
	return c.date
}

func (c *Calendar) SetDate(date time.Time) {
	c.date = date

	year, month, day := c.date.Date()
	monthStart := makeDate(year, month, 1)
	numDays := monthStart.AddDate(0, 1, -1).Day()

	c.firstIdx = (int(monthStart.Weekday()) + 6) % 7
	c.lastIdx = c.firstIdx + numDays
	c.cursor = c.firstIdx + day - 1

	if c.onDayChangeFunc != nil {
		c.onDayChangeFunc(c.date)
	}
}

func (c *Calendar) SetToday() {
	year, month, day := time.Now().Date()
	c.SetDate(time.Date(year, month, day, 12, 0, 0, 0, time.Local))
}

func (c *Calendar) HandleEvent(ev t.Event) (consume bool) {
	switch ev := ev.(type) {
	case *t.EventKey:
		switch ev.Key() {
		default:
			return false
		case t.KeyUp:
			c.SetDate(c.date.AddDate(0, 0, -7))
		case t.KeyDown:
			c.SetDate(c.date.AddDate(0, 0, 7))
		case t.KeyLeft:
			c.SetDate(c.date.AddDate(0, 0, -1))
		case t.KeyRight:
			c.SetDate(c.date.AddDate(0, 0, 1))

		case t.KeyRune:
			switch ev.Rune() {
			default:
				return false
			case 't':
				c.SetToday()
			case 'n':
				c.SetDate(c.date.AddDate(0, 1, 0))
			case 'p':
				c.SetDate(c.date.AddDate(0, -1, 0))
			case 'j':
				c.SetDate(c.date.AddDate(0, 0, 7))
			case 'k':
				c.SetDate(c.date.AddDate(0, 0, -7))
			case 'h':
				c.SetDate(c.date.AddDate(0, 0, -1))
			case 'l':
				c.SetDate(c.date.AddDate(0, 0, 1))
			}
		}
	}

	return true
}

var calenderHeaders = []string{
	"Mon",
	"Tue",
	"Wed",
	"Thu",
	"Fri",
	"Sat",
	"Sun",
}

func (c *Calendar) Render(renderer Renderer, hasFocus bool) {
	const (
		numCols      = 7
		numRows      = 7
		colWidth     = 5
		rowHeight    = 1
		headerHeight = 2
	)

	w, _ := renderer.Size()
	b := BordersRound
	bs := theme.Borders(hasFocus)
	renderer.PutStrStyled(-1, 1, b.TRB+strings.Repeat(b.LR, w)+b.TLB, bs)

	for i, header := range calenderHeaders {
		if len(header) < colWidth {
			renderer.PutStr((i*6)+1, 0, header)
		}
	}

	today := time.Now()
	start := c.date.AddDate(0, 0, -c.cursor)

	for idx := range 42 {
		row, col := idx/7, idx%7
		date := start.AddDate(0, 0, idx)

		dayStyle := theme.CalendarDay()
		if idx < c.firstIdx || idx > c.lastIdx {
			dayStyle = theme.CalendarOutside(dayStyle)
		}
		if idx == c.cursor {
			dayStyle = theme.CalendarSelect(dayStyle)
		}
		year, month, day := date.Date()
		if day == today.Day() && month == today.Month() && year == today.Year() {
			dayStyle = theme.CalendarToday(dayStyle)
		}

		x := 1 + (col * (colWidth + 1))
		y := headerHeight + (row * (rowHeight + 1))

		renderer.PutStrStyled(x, y, "    ", dayStyle)

		if c.underlineDay != nil && c.underlineDay(date) {
			dayStyle = dayStyle.Underline(true)
		}
		renderer.PutStrStyled(x+1, y, fmt.Sprintf("%02d", day), dayStyle)
	}
}

func makeDate(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 12, 0, 0, 0, time.Local)
}

func dateOnly(t time.Time) time.Time {
	return makeDate(t.Date())
}
