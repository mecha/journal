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
	journal        *Journal
	preview        *Preview
	calendar       *c.Calendar
	gotoPrompt     *c.InputPrompt
	gotoState      *c.InputState
	date           time.Time
	showGotoPrompt bool
	confirmDelete  bool
	deleteChoice   bool
	handler        c.EventHandler
}

func CreateDayPicker(journal *Journal, preview *Preview) *DayPicker {
	calendar := c.NewCalendar().
		UnderlineDay(func(date time.Time) bool {
			hasEntry, _ := journal.HasEntry(date)
			return hasEntry
		}).
		OnDayChanged(preview.Update)

	return &DayPicker{
		journal:   journal,
		preview:   preview,
		calendar:  calendar,
		date:      time.Now(),
		gotoState: &c.InputState{},
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

	if d.handler != nil && d.handler(ev) {
		return true
	}

	switch ev := ev.(type) {
	case *t.EventKey:
		switch ev.Key() {
		case t.KeyEsc:
			d.showGotoPrompt = false
			d.confirmDelete = false
			d.deleteChoice = false
			d.gotoState.Value = ""
			d.gotoState.Cursor = 0
			return true
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
				d.showGotoPrompt = true
				return true
			}
		}
	}

	return false
}

func (dp *DayPicker) Render(r c.Renderer, hasFocus bool) {
	title := fmt.Sprintf("[1]─%s %d", dp.date.Month().String(), dp.date.Year())
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

	switch {
	case dp.showGotoPrompt:
		panel := c.DrawPanel(popupRegion, "Go to (dd/mm/yyyy)", theme.BordersFocus())
		dp.handler = c.DrawInput(panel, dp.gotoState, c.InputProps{
			HideCursor: !hasFocus,
			OnEnter: func() {
				date, err := utils.ParseDayMonthYear(dp.gotoState.Value)
				if err == nil {
					dp.date = date
				} else {
					log.Println(err)
				}
				dp.gotoState.Value = ""
				dp.gotoState.Cursor = 0
				dp.showGotoPrompt = false
				screen.HideCursor()
				screen.PostEvent(NewRerenderEvent())
			},
		})

	case dp.confirmDelete:
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
					dp.preview.Update(dp.date)
					log.Printf("deleted entry: %s", dp.journal.EntryPath(dp.date))
				}
				dp.confirmDelete = false
			},
		})
	}
}

type DayPickerState struct {
	gotoInput      *c.InputState
	deleteChoice   bool
	showGotoPrompt bool
	confirmDelete  bool
}

type DayPickerProps struct {
	journal  *Journal
	hasFocus bool
	date     time.Time
	OnChange func(time.Time)
}

func DayPicker2(r c.Renderer, state *DayPickerState, props DayPickerProps) c.EventHandler {
	title := fmt.Sprintf("[1]─%s %d", props.date.Month().String(), props.date.Year())
	panelRegion := c.DrawPanel(r, title, theme.Borders(props.hasFocus))

	handler := c.DrawCalendar(panelRegion, c.CalendarProps{
		HasFocus: true,
		Selected: props.date,
		OnSelectDay: props.OnChange,
		UnderlineDays: func(t time.Time) bool {
			has, _ := props.journal.HasEntry(t)
			return has
		},
	})

	popupRegion := c.CenteredRegion(c.NewScreenRenderer(screen), 40, 3)

	switch {
	case state.showGotoPrompt:
		panel := c.DrawPanel(popupRegion, "Go to (dd/mm/yyyy)", theme.BordersFocus())
		handler = c.DrawInput(panel, state.gotoInput, c.InputProps{
			HideCursor: !props.hasFocus,
			OnEnter: func() {
				date, err := utils.ParseDayMonthYear(state.gotoInput.Value)
				if err == nil {
					props.date = date
				} else {
					log.Println(err)
				}
				state.gotoInput.Value = ""
				state.gotoInput.Cursor = 0
				state.showGotoPrompt = false
				screen.HideCursor()
				screen.PostEvent(NewRerenderEvent())
			},
		})

	case state.confirmDelete:
		handler = c.DrawConfirm(popupRegion, true, c.ConfirmProps{
			Message: "Are you sure you want to delete this journal entry?",
			Yes:     "Yes",
			No:      "No",
			Border:  theme.Borders(true, theme.Dialog()),
			Value:   state.deleteChoice,
			OnSelect: func(value bool) {
				state.deleteChoice = value
			},
			OnChoice: func(accepted bool) {
				if accepted {
					props.journal.DeleteEntry(props.date)
					// props.preview.Update(props.date)
					log.Printf("deleted entry: %s", props.journal.EntryPath(props.date))
				}
				state.confirmDelete = false
			},
		})
	}

	return func(ev t.Event) bool {
		if !props.journal.IsMounted() {
			return false
		}

		if handler != nil && handler(ev) {
			return true
		}

		switch ev := ev.(type) {
		case *t.EventKey:
			switch ev.Key() {
			case t.KeyEsc:
				state.showGotoPrompt = false
				state.confirmDelete = false
				state.deleteChoice = false
				state.gotoInput.Value = ""
				state.gotoInput.Cursor = 0
				return true
			case t.KeyEnter:
				err := props.journal.EditEntry(props.date)
				if err != nil {
					log.Print(err)
				}
				return true
			case t.KeyRune:
				switch ev.Rune() {
				case 'd':
					if has, _ := props.journal.HasEntry(props.date); has {
						state.confirmDelete = true
					}
					return true
				case 'g':
					state.showGotoPrompt = true
					return true
				}
			}
		}

		return false
	}
}
