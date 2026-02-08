package components

import (
	"fmt"
	"journal-tui/theme"
	"strings"
	"time"

	t "github.com/gdamore/tcell/v2"
)

type Calendar struct {
	year            int
	month           int
	cursor          int
	firstIdx        int
	lastIdx         int
	numDays         int
	prevNumDays     int
	underlineDay    func(day, month, year int) bool
	onDayChangeFunc func(day, month, year int)
}

var _ Component = (*Calendar)(nil)

func NewCalendar() *Calendar {
	today := time.Now()
	c := &Calendar{
		year:  today.Year(),
		month: int(today.Month()),
	}
	c.analyzeMonth()
	c.Today()
	return c
}

func (c *Calendar) UnderlineDay(fn func(day, month, year int) bool) *Calendar {
	c.underlineDay = fn
	return c
}

func (c *Calendar) OnDayChanged(callback func(day, month, year int)) *Calendar {
	c.onDayChangeFunc = callback
	return c
}

func (c *Calendar) Current() (day int, month int, year int) {
	day, month, year = 1, c.month, c.year

	switch {
	case c.cursor < c.firstIdx:
		day = c.prevNumDays - c.firstIdx + c.cursor + 1
		month--
	case c.cursor > c.lastIdx:
		day = c.cursor - c.lastIdx
		month++
	default:
		day = 1 + c.cursor - c.firstIdx
	}

	month, year = normalizeMonthsAndYears(month, year)
	return day, month, year
}

func (c *Calendar) SetDay(day, month, year int) {
	c.month, c.year = month, year
	c.analyzeMonth()
	c.cursor = c.firstIdx + day - 1
	c.notifyDayChange()
}

func (c *Calendar) Today() {
	now := time.Now()
	c.SetDay(now.Day(), int(now.Month()), now.Year())
}

func (c *Calendar) PrevMonth() {
	c.month, c.year = normalizeMonthsAndYears(c.month-1, c.year)
	day, _, _ := c.Current()
	c.analyzeMonth()
	c.cursor = c.firstIdx + day - 1
	c.notifyDayChange()
}

func (c *Calendar) NextMonth() {
	c.month, c.year = normalizeMonthsAndYears(c.month+1, c.year)
	day, _, _ := c.Current()
	c.analyzeMonth()
	c.cursor = c.firstIdx + day - 1
	c.notifyDayChange()
}

func (c *Calendar) DayLeft() {
	c.cursor--
	if c.cursor < 0 {
		c.PrevMonth()
	}
	c.notifyDayChange()
}

func (c *Calendar) DayRight() {
	c.cursor++
	if c.cursor > 41 {
		c.NextMonth()
	}
	c.notifyDayChange()
}

func (c *Calendar) DayUp() {
	c.cursor -= 7
	if c.cursor < 0 {
		c.PrevMonth()
	}
	c.notifyDayChange()
}

func (c *Calendar) DayDown() {
	c.cursor += 7
	if c.cursor > 41 {
		c.NextMonth()
	}
	c.notifyDayChange()
}

func (c *Calendar) notifyDayChange() {
	if c.onDayChangeFunc != nil {
		c.onDayChangeFunc(c.Current())
	}
}

func (c *Calendar) HandleEvent(ev t.Event) (consume bool) {
	switch ev := ev.(type) {
	case *t.EventKey:
		switch ev.Key() {
		default:
			return false
		case t.KeyUp:
			c.DayUp()
		case t.KeyDown:
			c.DayDown()
		case t.KeyLeft:
			c.DayLeft()
		case t.KeyRight:
			c.DayRight()

		case t.KeyRune:
			switch ev.Rune() {
			default:
				return false
			case 't':
				c.Today()
			case 'n':
				c.NextMonth()
			case 'p':
				c.PrevMonth()
			case 'j':
				c.DayDown()
			case 'k':
				c.DayUp()
			case 'h':
				c.DayLeft()
			case 'l':
				c.DayRight()
			}
		}
	}

	return true
}

func (c *Calendar) analyzeMonth() {
	startDate := time.Date(c.year, time.Month(c.month), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0).AddDate(0, 0, -1)

	c.numDays = endDate.Day()
	c.firstIdx = (int(startDate.Weekday()) + 6) % 7
	c.lastIdx = c.firstIdx + c.numDays - 1

	c.prevNumDays = startDate.AddDate(0, 0, -1).Day()
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

	for row := range numRows - 1 {
		for col := range 7 {
			idx := col + (row * 7)
			day, month, year := 0, c.month, c.year

			switch {
			case idx < c.firstIdx:
				day = c.prevNumDays - c.firstIdx + 1 + idx
				month--
			case idx > c.lastIdx:
				day = idx - c.lastIdx
				month++
			default:
				day = idx - c.firstIdx + 1
			}

			month, year = normalizeMonthsAndYears(month, year)

			var dayStyle t.Style
			switch {
			case idx == c.cursor:
				dayStyle = theme.CalendarSelect()
			case idx < c.firstIdx || idx > c.lastIdx:
				dayStyle = theme.CalendarOutside()
			case day == today.Day() && month == int(today.Month()) && year == today.Year():
				dayStyle = theme.CalendarToday()
			default:
				dayStyle = theme.CalendarDay()
			}

			x := 1 + (col * (colWidth + 1))
			y := headerHeight + (row * (rowHeight + 1))

			renderer.PutStrStyled(x, y, "    ", dayStyle)

			if c.underlineDay != nil && c.underlineDay(day, month, year) {
				dayStyle = dayStyle.Underline(true)
			}
			renderer.PutStrStyled(x+1, y, fmt.Sprintf("%02d", day), dayStyle)
		}
	}
}

func normalizeMonthsAndYears(month, year int) (int, int) {
	switch {
	case month < 1:
		month = 12
		year--
	case month > 12:
		month = 1
		year++
	}
	return month, year
}
