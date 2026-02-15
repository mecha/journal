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

type DayPickerProps struct {
	state    *DayPickerState
	journal  *Journal
	hasFocus bool
	date     time.Time
	OnChange func(time.Time)
}

type DayPickerState struct {
	gotoInput        *c.InputState
	showGotoPrompt   bool
	showDelConfirm   bool
	delConfirmChoice bool
}

func DayPicker(r c.Renderer, props DayPickerProps) c.EventHandler {
	state := props.state
	return c.Box(r, c.BoxProps{
		Title:   fmt.Sprintf("[1]─%s %d", props.date.Month().String(), props.date.Year()),
		Borders: c.BordersRound,
		Style:   theme.Borders(props.hasFocus),
		Children: func(r c.Renderer) c.EventHandler {
			handler := c.Calendar(r, c.CalendarProps{
				BorderStyle: theme.Borders(props.hasFocus),
				Selected:    props.date,
				OnSelectDay: props.OnChange,
				UnderlineDays: func(t time.Time) bool {
					has, _ := props.journal.HasEntry(t)
					return has
				},
			})

			switch {
			case state.showGotoPrompt:
				region := c.CenteredRegion(r.GetScreen(), 25, 3)
				gotoHandler := c.Box(region, c.BoxProps{
					Title:   "Go to (dd/mm/yyyy)",
					Borders: c.BordersRound,
					Style:   theme.BordersFocus(),
					Children: func(r c.Renderer) c.EventHandler {
						return c.Input(r, c.InputProps{State: state.gotoInput})
					},
				})
				handler = func(ev t.Event) bool {
					switch ev := ev.(type) {
					case *t.EventKey:
						if ev.Key() == t.KeyEnter {
							date, err := utils.ParseDayMonthYear(state.gotoInput.Value)
							if err == nil {
								props.date = date
							} else {
								log.Println(err)
							}
							state.gotoInput.Value = ""
							state.gotoInput.Cursor = 0
							state.showGotoPrompt = false
							return true
						}
					}
					if gotoHandler != nil {
						return gotoHandler(ev)
					}
					return false
				}

			case state.showDelConfirm:
				region := c.CenteredRegion(r.GetScreen(), 40, 3)
				handler = c.Confirm(region, true, c.ConfirmProps{
					Message: "Are you sure you want to delete this journal entry?",
					Yes:     "Yes",
					No:      "No",
					Borders: c.BordersRound,
					Style:   theme.Borders(true, theme.Dialog()),
					Value:   state.delConfirmChoice,
					OnSelect: func(value bool) {
						state.delConfirmChoice = value
					},
					OnChoice: func(accepted bool) {
						if accepted {
							props.journal.DeleteEntry(props.date)
							log.Printf("deleted entry: %s", props.journal.EntryPath(props.date))
						}
						state.showDelConfirm = false
					},
				})
			}

			return func(ev t.Event) bool {
				if !props.journal.isMounted {
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
						state.showDelConfirm = false
						state.delConfirmChoice = false
						state.gotoInput.Value = ""
						state.gotoInput.Cursor = 0
						return true

					case t.KeyEnter:
						err := props.journal.EditEntry(props.date, false)
						if err != nil {
							log.Print(err)
						}
						return true

					case t.KeyRune:
						break

					default:
						return false
					}

					switch ev.Rune() {
					case 'd':
						if has, _ := props.journal.HasEntry(props.date); has {
							state.showDelConfirm = true
							state.delConfirmChoice = false
						}
						return true

					case 'g':
						state.showGotoPrompt = true
						return true

					case 'e':
						err := props.journal.EditEntry(props.date, true)
						if err != nil {
							log.Print(err)
						}
						return true
					}
				}

				return false
			}
		},
	})
}
