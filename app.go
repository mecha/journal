package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	c "github.com/mecha/journal/components"
	"github.com/mecha/journal/theme"
	"github.com/mecha/journal/utils"

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
	pwdError  error
	logs      *c.TextState
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
		logs:     &c.TextState{},
		pwdInput: &c.InputState{},
	}
	app.preview.Lines = []string{}
	return app
}

func (app *App) showEntryPreview(date time.Time) {
	if app.journal.isMounted {
		entry, has, err := app.journal.GetEntry(date)
		switch {
		case err != nil:
			log.Println(err)
		case has:
			app.preview.Lines = strings.Split(entry, "\n")
		default:
			app.preview.Lines = []string{"[No entry]"}
		}
	} else {
		app.preview.Lines = []string{"[Journal is locked]"}
	}
}

func (app *App) handlePasswordInput() {
	password := app.pwdInput.Value
	app.pwdInput.Value = ""
	app.pwdInput.Cursor = 0
	app.pwdError = nil

	if !app.journal.isMounted {
		app.pwdError = app.journal.Mount(password)
		if app.pwdError != nil {
			log.Println("failed to unlock journal; ", app.pwdError)
			return
		}

		log.Println("Unlocked journal")

		app.tagsList.update(app.journal)
		app.showEntryPreview(app.date)
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
	if app.journal.isMounted {
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

	if !app.journal.isMounted {
		style := theme.BordersFocus()
		if app.pwdError != nil {
			style = style.Foreground(t.ColorOrangeRed)
		}

		rect := c.CenterRect(r.GetRegion(), min(width, 40), 3)
		handler := c.Box(r.SubRegion(rect), c.BoxProps{
			Title:   "Password",
			Borders: c.BordersRound,
			Style:   style,
			Children: func(r c.Renderer) c.EventHandler {
				return c.Input(r, c.InputProps{
					State: app.pwdInput,
					Mask:  "*",
				})
			},
		})

		logoRegion := r.SubRegion(c.Rect{
			Pos:  c.Pos{X: (width - LogoSize.W) / 2, Y: rect.Y - LogoSize.H},
			Size: LogoSize,
		})
		for i, line := range Logo {
			logoRegion.PutStrStyled(0, i, line, theme.Logo())
		}

		if app.pwdError != nil {
			errRect := c.NewRect(rect.X+1, rect.Y+rect.H, rect.W-2, 3)
			c.Text(r.SubRegion(errRect), c.TextProps{
				Style: style.Bold(true),
				State: &c.TextState{Lines: []string{app.pwdError.Error()}},
			})
		}

		return func(ev t.Event) bool {
			if handler != nil && handler(ev) {
				return true
			}
			switch ev := ev.(type) {
			case *t.EventKey:
				if ev.Key() == t.KeyEnter {
					app.handlePasswordInput()
					return true
				}
			}
			return false
		}
	} else {
		var logsHeight = logsHeightSm
		isLogsFocused := app.focus == FocusLogs
		if isLogsFocused {
			logsHeight = min(14, logsHeightLg)
		} else {
			app.logs.Scroll.Y = len(app.logs.Lines)
		}

		mainRegion, helpRegion := r.SplitVertical(height - 1)
		topRegion, logsRegion := mainRegion.SplitVertical(height - logsHeight)
		leftRegion, previewRegion := topRegion.SplitHorizontal(calendarWidth)
		calRegion, tagsRegion := leftRegion.SplitVertical(calendarHeight)

		logsHandler := c.Box(logsRegion, c.BoxProps{
			Title:   fmt.Sprintf("[4]─Logs (%d)", len(app.logs.Lines)),
			Borders: c.BordersRound,
			Style:   theme.Borders(isLogsFocused),
			Children: func(r c.Renderer) c.EventHandler {
				return c.Text(r, c.TextProps{State: app.logs})
			},
		})

		tagsHandler := TagsBrowser(tagsRegion, TagsProps{
			state:         app.tagsList,
			journal:       app.journal,
			hasFocus:      app.focus == FocusTags,
			onSelectRef:   app.showEntryPreview,
			onDeselectRef: func() { app.showEntryPreview(app.date) },
		})

		previewHandler := c.Box(previewRegion, c.BoxProps{
			Title:   "[3]─Preview",
			Borders: c.BordersRound,
			Style:   theme.Borders(app.focus == FocusPreview),
			Children: func(r c.Renderer) c.EventHandler {
				return c.Text(r, c.TextProps{State: app.preview})
			},
		})

		dayPickerHandler := DayPicker(calRegion, DayPickerProps{
			state:    app.dayPicker,
			journal:  app.journal,
			hasFocus: app.focus == FocusDayPicker,
			date:     app.date,
			OnChange: func(newValue time.Time) {
				app.date = newValue
				app.showEntryPreview(newValue)
			},
		})

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
							app.logs.Lines = []string{}
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
		text = "Select day: ⬍/⬌ | Edit: <ENTER> or e | Delete: d | Today: t | Go to specific day: g | Exit: q"
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
	w.app.logs.Lines = append(w.app.logs.Lines, newLines...)
	w.app.logs.Scroll = c.Pos{X: 0, Y: len(w.app.logs.Lines)}
	return len(data), nil
}

var Logo = []string{
	"  ░██                                                      ░██ ",
	"                                                           ░██ ",
	"  ░██  ░███████  ░██    ░██ ░██░████ ░████████   ░██████   ░██ ",
	"  ░██ ░██    ░██ ░██    ░██ ░███     ░██    ░██       ░██  ░██ ",
	"  ░██ ░██    ░██ ░██    ░██ ░██      ░██    ░██  ░███████  ░██ ",
	"  ░██ ░██    ░██ ░██   ░███ ░██      ░██    ░██ ░██   ░██  ░██ ",
	"  ░██  ░███████   ░█████░██ ░██      ░██    ░██  ░█████░██ ░██ ",
	"  ░██                                                          ",
	"░███                                                           ",
}
var LogoSize = c.Size{W: utils.MaxLength(Logo), H: len(Logo)}
