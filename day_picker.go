package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	c "journal-tui/components"
	"journal-tui/theme"

	t "github.com/gdamore/tcell/v2"
)

type DayPicker struct {
	journal       *Journal
	preview       *Preview
	calendar      *c.Calendar
	confirmDelete *c.Confirm
	gotoPrompt    *c.InputPrompt
}

func CreateDayPicker(journal *Journal, preview *Preview) *DayPicker {
	calendar := c.NewCalendar().
		UnderlineDay(func(date time.Time) bool {
			hasEntry, _ := journal.HasEntry(date)
			return hasEntry
		}).
		OnDayChanged(preview.Update)

	return &DayPicker{
		journal:  journal,
		preview:  preview,
		calendar: calendar,
		gotoPrompt: c.NewInputPrompt(
			"Go to (dd/mm/yyyy)",
			c.NewInput(),
			func(input *c.Input, cancelled bool) {
				if !cancelled {
					value := input.Value()
					date, err := ParseDayMonthYear(value)
					if err == nil {
						calendar.SetDate(date)
					} else {
						log.Println(err)
					}
				}
				input.SetValue("")
				Screen.HideCursor()
				Focus.Pop()
				RenderScreen()
			}),
		confirmDelete: c.NewConfirm("Are you sure you want to delete this journal entry?", func(accepted bool) {
			if accepted {
				date := calendar.Date()
				journal.DeleteEntry(date)
				log.Printf("deleted entry: %s", journal.EntryPath(date))
				preview.Update(date)
			}
			Focus.Pop()
		}),
	}
}

func (d *DayPicker) HandleEvent(ev t.Event) bool {
	if !d.journal.IsMounted() {
		return false
	}

	if Focus.Is(d.confirmDelete) && d.confirmDelete.HandleEvent(ev) {
		return true
	}

	if Focus.Is(d.gotoPrompt) && d.gotoPrompt.HandleEvent(ev) {
		return true
	}

	switch ev := ev.(type) {
	case *t.EventKey:
		switch ev.Key() {
		case t.KeyEnter:
			date := d.calendar.Date()
			d.journal.EditEntry(date)
			return true
		case t.KeyRune:
			switch ev.Rune() {
			case 'd':
				if has, _ := d.journal.HasEntry(d.calendar.Date()); has {
					Focus.Push(d.confirmDelete)
				}
				return true
			case 'g':
				Focus.Push(d.gotoPrompt)
				return true
			}
		}
	}

	return d.calendar.HandleEvent(ev)
}

func (dp *DayPicker) Render(r c.Renderer, hasFocus bool) {
	date := dp.calendar.Date()
	title := fmt.Sprintf("[1]─%s %d", date.Month().String(), date.Year())
	panelRegion := c.DrawPanel(r, title, theme.Borders(hasFocus))

	dp.calendar.Render(panelRegion, hasFocus)

	popupRegion := c.CenteredRegion(Screen, 40, 3)

	if Focus.Is(dp.gotoPrompt) {
		dp.gotoPrompt.Render(popupRegion, true)
	}

	if Focus.Is(dp.confirmDelete) {
		dp.confirmDelete.Render(popupRegion, true)
	}
}

func ParseDayMonthYear(s string) (time.Time, error) {
	parts := strings.Split(s, "/")
	if len(parts) != 3 {
		return time.Time{}, errors.New("invalid date: must be <day>/<month>/<year>")
	}

	nums := [3]int{0, 0, 0}
	for i, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil {
			return time.Time{}, errors.New("invalid date: \"" + part + "\" is not a number")
		}
		nums[i] = num
	}

	return time.Date(nums[2], time.Month(nums[1]), nums[0], 0, 0, 0, 0, time.Local), nil
}
