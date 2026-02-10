package main

import (
	"io"
	"log"
	"strings"
	"time"

	c "github.com/mecha/journal/components"
	"github.com/mecha/journal/theme"

	t "github.com/gdamore/tcell/v2"
)

type App struct {
	journal        *Journal
	preview        *Preview
	dayPicker      *DayPicker
	tagBrowser     *TagBrowser
	passwordPrompt *c.InputPrompt
	logViewer      *c.Text
}

func CreateApp(journal *Journal) *App {
	app := &App{journal: journal}
	app.preview = CreatePreview(journal)
	app.dayPicker = CreateDayPicker(journal, app.preview)
	app.tagBrowser = CreateTagBrowser(journal, app.preview.Update, app.resetPreview)
	app.passwordPrompt = c.NewInputPrompt("Password", c.NewInput().SetMask('*'), app.handlePasswordInput)
	app.logViewer = c.NewText([]string{})
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

func (app *App) resetPreview() {
	app.preview.Update(app.dayPicker.calendar.Date())
}

func (app *App) promptForPassword() {
	focus.Push(app.passwordPrompt)
	app.tagBrowser.UpdateTags()
	app.resetPreview()
}

func (app *App) handlePasswordInput(input *c.Input, cancelled bool) {
	if cancelled {
		return
	}

	password := input.Value()
	input.SetValue("")

	if !app.journal.IsMounted() {
		err := app.journal.Mount(password)
		if err != nil {
			log.Println("failed to unlock journal; ", err)
			return
		}

		log.Println("Unlocked journal")
		focus.Pop()
		screen.HideCursor()

		app.tagBrowser.UpdateTags()
		app.resetPreview()
	}
}

func (app *App) HandleEvent(ev t.Event) bool {
	if focus.HandleEvent(ev) {
		return true
	}

	switch ev := ev.(type) {
	case *JournalUnmountEvent:
		app.promptForPassword()
		return true

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
				app.dayPicker.calendar.SetToday()
				return true
			}
		case t.KeyCtrlU, t.KeyPgUp:
			app.preview.text.ScrollUp(10)
			return true
		case t.KeyCtrlD, t.KeyPgDn:
			app.preview.text.ScrollDown(10)
			return true
		}
	}

	return false
}

const logsHeightSm = 6
const logsHeightLg = 14
const calendarWidth = 45
const calendarHeight = 15

func (app *App) Render(r c.Renderer, hasFocus bool) {
	var logsHeight = logsHeightSm
	if focus.Is(app.logViewer) {
		logsHeight = min(14, logsHeightLg)
	} else {
		app.logViewer.ScrollToBottom()
	}

	_, height := r.Size()
	mainRegion, helpRegion := r.SplitVertical(height - 1)
	topRegion, logsRegion := mainRegion.SplitVertical(height - logsHeight)
	leftRegion, previewRegion := topRegion.SplitHorizontal(calendarWidth)
	calRegion, tagsRegion := leftRegion.SplitVertical(calendarHeight)

	logsInner := c.DrawPanel(logsRegion, "[4]─Log", theme.Borders(focus.Is(app.logViewer)))
	app.logViewer.Render(logsInner, focus.Is(app.logViewer))

	app.dayPicker.Render(calRegion, focus.Is(app.dayPicker))
	app.tagBrowser.Render(tagsRegion, focus.Is(app.tagBrowser))
	app.preview.Render(previewRegion, focus.Is(app.preview))

	if focus.Is(app.passwordPrompt) {
		rect := c.CenterRect(r.GetRegion(), 40, 3)
		app.passwordPrompt.Render(r.SubRegion(rect), true)
	}

	helpRegion.PutStrStyled(0, 0, app.bottomBarText(), theme.Help())
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
	case app.passwordPrompt:
		return "Submit: <enter>"
	}
	return ""
}

type AppLogWriter struct{ app *App }

func (w *AppLogWriter) Write(data []byte) (int, error) {
	w.app.addLog(string(data))
	return len(data), nil
}

type JournalUnmountEvent struct{ when time.Time }

func (ev *JournalUnmountEvent) When() time.Time { return ev.when }
