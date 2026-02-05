package components

import (
	"fmt"
	"journal-tui/journal"
	"journal-tui/render"
	"journal-tui/theme"
	"time"

	t "github.com/gdamore/tcell/v2"
)

type Calendar struct {
	ComponentPos
	journal         *journal.Journal
	year            int
	month           int
	cursor          int
	firstIdx        int
	lastIdx         int
	numDays         int
	prevNumDays     int
	onDayChangeFunc func(day, month, year int)
}

var _ Component = (*Calendar)(nil)

func NewCalendar(journal *journal.Journal, x, y int) *Calendar {
	today := time.Now()
	c := &Calendar{
		ComponentPos: ComponentPos{x, y},
		journal:      journal,
		year:         today.Year(),
		month:        int(today.Month()),
	}
	c.analyzeMonth()
	c.Today()
	return c
}

func (c *Calendar) DayUnderCursor() (day int, month int, year int) {
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

func (c *Calendar) Today() {
	now := time.Now()
	c.month, c.year = int(now.Month()), now.Year()
	c.analyzeMonth()
	c.cursor = c.firstIdx + now.Day() - 1
	c.notifyDayChange()
}

func (c *Calendar) PrevMonth() {
	c.month, c.year = normalizeMonthsAndYears(c.month-1, c.year)
	day, _, _ := c.DayUnderCursor()
	c.analyzeMonth()
	c.cursor = c.firstIdx + day - 1
	c.notifyDayChange()
}

func (c *Calendar) NextMonth() {
	c.month, c.year = normalizeMonthsAndYears(c.month+1, c.year)
	day, _, _ := c.DayUnderCursor()
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

func (c *Calendar) OnDayChanged(callback func(day, month, year int)) {
	c.onDayChangeFunc = callback
}

func (c *Calendar) notifyDayChange() {
	if c.onDayChangeFunc != nil {
		c.onDayChangeFunc(c.DayUnderCursor())
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
		case t.KeyEnter:
			c.journal.EditEntry(c.DayUnderCursor())

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

func (c *Calendar) Size() (int, int) { return 45, 15 }

func (c *Calendar) Render(screen t.Screen, hasFocus bool) {
	const (
		numCols      = 7
		numRows      = 7
		colWidth     = 5
		rowHeight    = 1
		headerHeight = 3
	)
	width, _ := c.Size()

	borderStyle := theme.BorderStyle(hasFocus)

	title := fmt.Sprintf("[1]─%s %d", time.Month(c.month).String(), c.year)
	render.Box(screen, c.x, c.y, width, 15, render.RoundedBorders, borderStyle)
	screen.PutStrStyled(c.x+2, c.y, title, borderStyle)
	render.BoxHorizontalDivider(screen, c.x, c.y+2, width, render.RoundedBorders, borderStyle)

	for i, header := range calenderHeaders {
		if len(header) < colWidth {
			screen.PutStr(c.x+(i*6)+2, c.y+1, header)
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

			dayStyle := t.StyleDefault
			switch {
			case idx == c.cursor:
				dayStyle = theme.CalendarSelect
			case idx < c.firstIdx || idx > c.lastIdx:
				dayStyle = theme.CalendarOutside
			case day == today.Day() && month == int(today.Month()) && year == today.Year():
				dayStyle = theme.CalendarToday
			default:
				dayStyle = theme.CalendarDay
			}

			hasEntry, _ := c.journal.HasEntry(day, month, year)
			// dayStyle = dayStyle.Underline(hasEntry)
			dayText := fmt.Sprintf("%02d", day)

			x := c.x + 2 + (col * (colWidth + 1))
			y := c.y + headerHeight + (row * (rowHeight + 1))

			screen.PutStrStyled(x, y, "    ", dayStyle)
			screen.PutStrStyled(x+1, y, dayText, dayStyle.Underline(hasEntry))
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
