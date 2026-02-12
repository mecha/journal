package main

import (
	"fmt"
	"log"
	"time"

	c "github.com/mecha/journal/components"
	"github.com/mecha/journal/theme"
	"github.com/mecha/journal/utils"

	t "github.com/gdamore/tcell/v2"
)

type DayPicker struct {
	journal       *Journal
	preview       *Preview
	calendar      *c.Calendar
	gotoPrompt    *c.InputPrompt
	date          time.Time
	confirmDelete bool
	deleteChoice  bool
	handler       c.EventHandler
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
		date:     time.Now(),
		gotoPrompt: c.NewInputPrompt(
			"Go to (dd/mm/yyyy)",
			c.NewInput(),
			func(input *c.Input, cancelled bool) {
				if !cancelled {
					value := input.Value()
					date, err := utils.ParseDayMonthYear(value)
					if err == nil {
						calendar.SetDate(date)
					} else {
						log.Println(err)
					}
				}
				input.SetValue("")
				focus.Pop()
				screen.HideCursor()
				screen.PostEvent(NewRerenderEvent())
			}),
	}
}

func (d *DayPicker) HandleEvent(ev t.Event) bool {
	if !d.journal.IsMounted() {
		return false
	}

	if focus.Is(d.gotoPrompt) && d.gotoPrompt.HandleEvent(ev) {
		return true
	}

	switch ev := ev.(type) {
	case *t.EventKey:
		switch ev.Key() {
		case t.KeyEnter:
			err := d.journal.EditEntry(d.date)
			if err != nil {
				log.Print(err)
			}
			return true
		case t.KeyRune:
			switch ev.Rune() {
			case 'd':
				if has, _ := d.journal.HasEntry(d.date); has {
					d.confirmDelete = true
				}
				return true
			case 'g':
				focus.Push(d.gotoPrompt)
				return true
			}
		}
	}

	if d.handler != nil {
		return d.handler(ev)
	}
	return false
}

func (dp *DayPicker) Render(r c.Renderer, hasFocus bool) {
	date := dp.calendar.Date()
	title := fmt.Sprintf("[1]─%s %d", date.Month().String(), date.Year())
	panelRegion := c.DrawPanel(r, title, theme.Borders(hasFocus))

	dp.handler = c.DrawCalendar(panelRegion, c.CalendarProps{
		HasFocus: true,
		Selected: dp.date,
		OnSelectDay: func(value time.Time) {
			dp.date = value
			dp.preview.Update(dp.date)
		},
		UnderlineDays: func(t time.Time) bool {
			has, _ := dp.journal.HasEntry(t)
			return has
		},
	})

	popupRegion := c.CenteredRegion(c.NewScreenRenderer(screen), 40, 3)

	if focus.Is(dp.gotoPrompt) {
		dp.gotoPrompt.Render(popupRegion, true)
	}

	if dp.confirmDelete {
		dp.handler = c.DrawConfirm(popupRegion, true, c.ConfirmProps{
			Message: "Are you sure you want to delete this journal entry?",
			Yes:     "Yes",
			No:      "No",
			Border:  theme.Borders(true, theme.Dialog()),
			Value:   dp.deleteChoice,
			OnSelect: func(value bool) {
				dp.deleteChoice = value
			},
			OnChoice: func(accepted bool) {
				if accepted {
					dp.journal.DeleteEntry(dp.date)
					dp.preview.Update(date)
					log.Printf("deleted entry: %s", dp.journal.EntryPath(date))
				}
				dp.confirmDelete = false
			},
		})
	}
}
