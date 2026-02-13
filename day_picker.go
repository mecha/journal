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
