package main

import (
	"fmt"
	"log"
	"time"

	c "journal-tui/components"
	j "journal-tui/journal"
	"journal-tui/utils"

	t "github.com/gdamore/tcell/v2"
)

type DayPicker struct {
	journal       *j.Journal
	preview       *PreviewComp
	calendar      *c.Calendar
	confirmDelete *c.Confirm
	gotoPrompt    *c.InputPrompt
}

func CreateDayPicker(journal *j.Journal, preview *PreviewComp) *DayPicker {
	calendar := c.NewCalendar().
		UnderlineDay(func(day, month, year int) bool {
			hasEntry, _ := journal.HasEntry(day, month, year)
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
					day, month, year, err := utils.ParseDayMonthYear(value)
					if err == nil {
						calendar.SetDay(day, month, year)
					} else {
						log.Println(err)
					}
				}
				input.SetValue("")
				Screen.HideCursor()
				Focus.Pop()
				renderScreen()
			}),
		confirmDelete: c.NewConfirm("Are you sure you want to delete this journal entry?", func(accepted bool) {
			if accepted {
				day, month, year := calendar.Current()
				journal.DeleteEntry(day, month, year)
				log.Printf("deleted entry: %s", Journal.EntryPath(day, month, year))
				preview.Update(day, month, year)
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
			d.journal.EditEntry(d.calendar.Current())
			d.preview.Update(d.calendar.Current())
			return true
		case t.KeyRune:
			switch ev.Rune() {
			case 'd':
				if has, _ := d.journal.HasEntry(d.calendar.Current()); has {
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
	_, month, year := dp.calendar.Current()
	title := fmt.Sprintf("[1]─%s %d", time.Month(month).String(), year)
	panelRegion := c.DrawPanel(r, title, t.StyleDefault, hasFocus)

	dp.calendar.Render(panelRegion, hasFocus)

	popupRegion := c.CenteredRegion(Screen, 40, 3)

	if Focus.Is(dp.gotoPrompt) {
		dp.gotoPrompt.Render(popupRegion, true)
	}

	if Focus.Is(dp.confirmDelete) {
		dp.confirmDelete.Render(popupRegion, true)
	}
}
