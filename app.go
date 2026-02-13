package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	c "github.com/mecha/journal/components"
	"github.com/mecha/journal/theme"

	t "github.com/gdamore/tcell/v2"
)

type App struct {
	journal   *Journal
	focus     int
	date      time.Time
	dayPicker *DayPickerState
	tagsList  *TagsState
	preview   *c.TextState
	pwdInput  *c.InputState
	logsState *c.TextState
}

const (
	FocusDayPicker = iota
	FocusTags
	FocusPreview
	FocusLogs
)

func CreateApp(journal *Journal) *App {
	app := &App{
		journal:   journal,
		focus:     FocusDayPicker,
		date:      time.Now(),
		dayPicker: &DayPickerState{gotoInput: &c.InputState{}},
		preview:   &c.TextState{},
		tagsList: &TagsState{
			tags:    []string{},
			refs:    []time.Time{},
			tagList: &c.ListState[string]{},
			refList: &c.ListState[time.Time]{},
		},
		logsState: &c.TextState{},
		pwdInput:  &c.InputState{},
	}
	app.preview.SetLines([]string{})
	return app
}

func (app *App) showPreview(date time.Time) {
	if app.journal.IsMounted() {
		entry, has, err := app.journal.GetEntry(date)
		switch {
		case err != nil:
			log.Println(err)
		case has:
			app.preview.SetLines(strings.Split(entry, "\n"))
		default:
			app.preview.SetLines([]string{"[No entry]"})
		}
	} else {
		app.preview.SetLines([]string{"[Journal is locked]"})
	}
}

func (app *App) resetPreview() {
	app.showPreview(app.date)
}

func (app *App) updateTags() {
	app.tagsList.update(app.journal)
}

func (app *App) handlePasswordInput() {
	password := app.pwdInput.Value
	app.pwdInput.Value = ""
	app.pwdInput.Cursor = 0

	if !app.journal.IsMounted() {
		err := app.journal.Mount(password)
		if err != nil {
			log.Println("failed to unlock journal; ", err)
			return
		}

		log.Println("Unlocked journal")
		screen.HideCursor()

		app.updateTags()
		app.resetPreview()
	}
}

const logsHeightSm = 6
const logsHeightLg = 14
const calendarWidth = 45
const calendarHeight = 15

var minSizeLocked = c.Size{W: 28, H: 3}
var minSizeUnlocked = c.Size{W: 64, H: 26}

func DrawApp(r c.Renderer, app *App) c.EventHandler {
	width, height := r.Size()

	var minSize c.Size
	if app.journal.IsMounted() {
		minSize = minSizeUnlocked
	} else {
		minSize = minSizeLocked
	}

	if width < minSize.W || height < minSize.H {
		line1 := "Terminal is too small."
		line2 := fmt.Sprintf("Current size: %d x %d", width, height)
		line3 := "Must be at least: 64 x 26"
		x, y := max(0, (width-len(line3))/2), max(0, (height-3)/2)
		r.PutStr(x, y, line1)
		r.PutStr(x, y+1, line2)
		r.PutStr(x, y+2, line3)
		return nil
	}

	if !app.journal.IsMounted() {
		rect := c.CenterRect(r.GetRegion(), min(width, 40), 3)
		inner := c.DrawPanel(r.SubRegion(rect), "Password", theme.BordersFocus())

		return c.DrawInput(inner, app.pwdInput, c.InputProps{
			OnEnter: app.handlePasswordInput,
		})
	} else {
		var logsHeight = logsHeightSm
		isLogsFocused := app.focus == FocusLogs
		if isLogsFocused {
			logsHeight = min(14, logsHeightLg)
		} else {
			app.logsState.ScrollToBottom()
		}

		mainRegion, helpRegion := r.SplitVertical(height - 1)
		topRegion, logsRegion := mainRegion.SplitVertical(height - logsHeight)
		leftRegion, previewRegion := topRegion.SplitHorizontal(calendarWidth)
		calRegion, tagsRegion := leftRegion.SplitVertical(calendarHeight)

		insideLogs := c.DrawPanel(logsRegion, "[4]─Log", theme.Borders(isLogsFocused))
		logsHandler := c.DrawText(insideLogs, app.logsState, c.TextProps{})

		dayPickerHandler := DayPicker2(calRegion, app.dayPicker, DayPickerProps{
			journal:  app.journal,
			hasFocus: app.focus == FocusDayPicker,
			date:     app.date,
			OnChange: func(newValue time.Time) {
				app.date = newValue
				app.showPreview(newValue)
			},
		})

		tagsHandler := DrawTags(tagsRegion, app.tagsList, TagsProps{
			journal:  app.journal,
			hasFocus: app.focus == FocusTags,
		})

		insidePreview := c.DrawPanel(previewRegion, "[3]─Preview", theme.Borders(app.focus == FocusPreview))
		previewHandler := c.DrawText(insidePreview, app.preview, c.TextProps{})

		DrawHelp(helpRegion, app.focus)

		return func(ev t.Event) bool {
			switch app.focus {
			case FocusDayPicker:
				if dayPickerHandler(ev) {
					return true
				}
			case FocusTags:
				if tagsHandler(ev) {
					return true
				}
			case FocusPreview:
				if previewHandler(ev) {
					return true
				}
			case FocusLogs:
				if logsHandler(ev) {
					return true
				}
			}

			switch ev := ev.(type) {
			case *t.EventKey:
				switch key := ev.Key(); key {
				case t.KeyRune:
					switch ev.Rune() {
					case '1':
						app.focus = FocusDayPicker
					case '2':
						app.focus = FocusTags
					case '3':
						app.focus = FocusPreview
					case '4':
						app.focus = FocusLogs
					case 'c':
						if app.focus == FocusLogs {
							app.logsState.SetLines([]string{})
						}
					case 't':
						app.date = time.Now()
						return true
					}
				case t.KeyTab:
					app.focus = (app.focus + 1) % 4
				case t.KeyBacktab:
					app.focus = (app.focus + 3) % 4
				case t.KeyCtrlU, t.KeyPgUp:
					app.preview.Scroll = app.preview.Scroll.Add(0, -10)
					return true
				case t.KeyCtrlD, t.KeyPgDn:
					app.preview.Scroll = app.preview.Scroll.Add(0, 10)
					return true
				}
			}

			return false
		}
	}
}

func DrawHelp(r c.Renderer, focus int) {
	text := ""
	switch focus {
	case FocusDayPicker:
		text = "Select day: ⬍/⬌ | Edit: <ENTER> | Delete: d | Today: t | Go to specific day: g | Next/Previous month: n/p | Exit: q"
	case FocusTags:
		text = "Select: ⬍ | View entries: <ENTER>"
	case FocusPreview:
		text = "Scroll: ⬍"
	case FocusLogs:
		text = "Select: ⬍ | Clear: c"
	}

	r.PutStrStyled(0, 0, text, theme.Help())
}

type AppLogWriter struct{ app *App }

func (w *AppLogWriter) Write(data []byte) (int, error) {
	newLines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	w.app.logsState.AddLines(newLines)
	w.app.logsState.ScrollToBottom()
	return len(data), nil
}
