package main

import (
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	c "github.com/mecha/journal/components"
	"github.com/mecha/journal/theme"

	t "github.com/gdamore/tcell/v2"
)

type App struct {
	journal       *Journal
	tagBrowser    *TagBrowser
	pwdInputState *c.InputState
	logViewer     *c.Text
	handler       c.EventHandler

	dayPicker c.Component
	preview   *Preview

	// state
	date           time.Time
	dayPickerState *DayPickerState
	previewState   *c.TextState
}

func CreateApp(journal *Journal) *App {
	app := &App{journal: journal}
	app.date = time.Now()
	// app.preview = CreatePreview(journal)
	app.dayPicker = CreateDayPicker(journal, app.preview)
	app.dayPickerState = &DayPickerState{gotoInput: &c.InputState{}}
	app.previewState = &c.TextState{}
	app.previewState.SetLines([]string{})
	app.tagBrowser = CreateTagBrowser(journal, app.showEntryPreview, app.resetPreview)
	app.logViewer = c.NewText([]string{})
	app.pwdInputState = &c.InputState{}
	return app
}

func (app *App) logWriter() io.Writer {
	return &AppLogWriter{app}
}

func (app *App) addLog(message string) {
	newLines := strings.Split(strings.TrimSuffix(message, "\n"), "\n")
	app.logViewer.AddLines(newLines)
	app.logViewer.ScrollToBottom()
}

func (app *App) showEntryPreview(date time.Time) {
	if app.journal.IsMounted() {
		entry, has, err := app.journal.GetEntry(date)
		switch {
		case err != nil:
			log.Println(err)
		case has:
			app.previewState.SetLines(strings.Split(entry, "\n"))
		default:
			app.previewState.SetLines([]string{"[No entry]"})
		}
	} else {
		app.previewState.SetLines([]string{"[Journal is locked]"})
	}
}

func (app *App) resetPreview() {
	app.showEntryPreview(app.date)
}

func (app *App) handlePasswordInput() {
	password := app.pwdInputState.Value
	app.pwdInputState.Value = ""
	app.pwdInputState.Cursor = 0

	if !app.journal.IsMounted() {
		err := app.journal.Mount(password)
		if err != nil {
			log.Println("failed to unlock journal; ", err)
			return
		}

		log.Println("Unlocked journal")
		screen.HideCursor()

		app.tagBrowser.UpdateTags()
		app.resetPreview()
	}
}

func (app *App) HandleEvent(ev t.Event) bool {
	if app.handler != nil && app.handler(ev) {
		return true
	}

	if focus.HandleEvent(ev) {
		return true
	}

	switch ev := ev.(type) {
	case *t.EventResize:
		screen.Sync()
		return true

	case *t.EventKey:
		switch key := ev.Key(); key {
		case t.KeyRune:
			switch ev.Rune() {
			case 'c':
				if focus.Is(app.logViewer) {
					app.logViewer.SetLines([]string{})
				}
			case 't':
				app.date = time.Now()
				return true
			}
		case t.KeyCtrlU, t.KeyPgUp:
			app.previewState.Scroll = app.previewState.Scroll.Add(0, -10)
			return true
		case t.KeyCtrlD, t.KeyPgDn:
			app.previewState.Scroll = app.previewState.Scroll.Add(0, 10)
			return true
		}
	}

	return false
}

const logsHeightSm = 6
const logsHeightLg = 14
const calendarWidth = 45
const calendarHeight = 15

var minSizeLocked = c.Size{W: 28, H: 3}
var minSizeUnlocked = c.Size{W: 64, H: 26}

func (app *App) Render(r c.Renderer, hasFocus bool) {
	app.handler = nil
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
		return
	}

	if !app.journal.IsMounted() {
		rect := c.CenterRect(r.GetRegion(), min(width, 40), 3)
		inner := c.DrawPanel(r.SubRegion(rect), "Password", theme.BordersFocus())
		app.handler = c.DrawInput(inner, app.pwdInputState, c.InputProps{
			OnEnter: app.handlePasswordInput,
		})
	} else {
		var logsHeight = logsHeightSm
		if focus.Is(app.logViewer) {
			logsHeight = min(14, logsHeightLg)
		} else {
			app.logViewer.ScrollToBottom()
		}

		mainRegion, helpRegion := r.SplitVertical(height - 1)
		topRegion, logsRegion := mainRegion.SplitVertical(height - logsHeight)
		leftRegion, previewRegion := topRegion.SplitHorizontal(calendarWidth)
		calRegion, tagsRegion := leftRegion.SplitVertical(calendarHeight)

		logsInner := c.DrawPanel(logsRegion, "[4]─Log", theme.Borders(focus.Is(app.logViewer)))
		app.logViewer.Render(logsInner, focus.Is(app.logViewer))

		dayPickerFocused := focus.Is(app.dayPicker)
		dayPickerHandler := DayPicker2(calRegion, app.dayPickerState, DayPickerProps{
			journal:  app.journal,
			hasFocus: dayPickerFocused,
			date:     app.date,
			OnChange: func(newValue time.Time) {
				app.date = newValue
				app.showEntryPreview(newValue)
			},
		})
		if dayPickerFocused {
			app.handler = dayPickerHandler
		}

		app.tagBrowser.Render(tagsRegion, focus.Is(app.tagBrowser))

		previewFocused := focus.Is(app.preview)
		insidePreview := c.DrawPanel(previewRegion, "[3]─Preview", theme.Borders(previewFocused))
		previewHandler := c.DrawText(insidePreview, app.previewState, c.TextProps{})
		if previewFocused {
			app.handler = previewHandler
		}

		helpRegion.PutStrStyled(0, 0, app.bottomBarText(), theme.Help())
	}
}

func (app *App) bottomBarText() string {
	switch focus.Current() {
	case app.dayPicker:
		return "Select day: ⬍/⬌ | Edit: <ENTER> | Delete: d | Today: t | Go to specific day: g | Next/Previous month: n/p | Exit: q"
	case app.tagBrowser:
		return "Select: ⬍ | View entries: <ENTER>"
	case app.logViewer:
		return "Select: ⬍ | Clear: c"
	case app.preview:
		return "Scroll: ⬍"
	}
	return ""
}

type AppLogWriter struct{ app *App }

func (w *AppLogWriter) Write(data []byte) (int, error) {
	w.app.addLog(string(data))
	return len(data), nil
}
