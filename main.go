package main

import (
	"io"
	"log"
	"os"
	"strings"
	"time"

	"journal-tui/components"
	"journal-tui/journal"
	"journal-tui/render"
	"journal-tui/theme"

	t "github.com/gdamore/tcell/v2"
)

func main() {
	parseFlags()

	screen, err := t.NewScreen()
	if err != nil {
		log.Fatal(err)
	}

	if err = screen.Init(); err != nil {
		log.Fatal(err)
	}

	journal := journal.NewJournal(Flags.path, Flags.mntPath, Flags.idleTimeout)

	var app *App

	app = &App{
		screen:        screen,
		journal:       journal,
		calendar:      components.NewCalendar(journal, 0, 3),
		tags:          components.NewTagList(journal, 0, 18),
		logs:          components.NewListComponent("[3]─Log", 0, 0, 0, 0),
		preview:       components.NewMarkdownComponent("[4]-Preview", ""),
		passwordInput: components.NewInputComponent("Password", func(value string) { app.mountJournal(value) }).SetMask('*'),
		confirmDelete: components.NewConfirmComponent("Are you sure you want to delete this journal entry?", func(accepted bool) {
			if accepted {
				journal.DeleteEntry(app.calendar.DayUnderCursor())
			}
			app.setFocus(app.calendar)
		}),
		help: components.NewTextComponent("", theme.Help, " "),
	}
	app.Resize(screen.Size())
	app.setFocus(app.passwordInput)

	app.calendar.OnDayChanged(app.updatePreview)
	journal.OnUnmount(app.onJournalUnmonut)

	// write log messages to the logs component
	log.SetOutput(&LogWriter{
		func(msg string) {
			line := strings.ReplaceAll(msg, "\n", " | ")
			app.logs.AddItem(line)
			app.Render()
		},
	})

	go func() {
		for range time.NewTicker(3 * time.Second).C {
			app.updateTags()
		}
	}()

	for {
		ev := screen.PollEvent()
		app.HandleEvent(ev)
		app.Render()
	}
}

const MountWaitTime = time.Millisecond * 150

type App struct {
	screen        t.Screen
	journal       *journal.Journal
	calendar      *components.Calendar
	tags          *components.TagList
	preview       *components.MarkdownComponent
	passwordInput *components.InputComponent
	confirmDelete *components.ConfirmComponent
	logs          *components.ListComponent
	help          *components.TextComponent
	focusedComp   components.Component
}

func (app *App) Resize(width, height int) {
	app.logs.Move(0, height-6)
	app.logs.Resize(width, 5)

	app.tags.Move(0, 18)
	app.tags.Resize(45, height-18-6)

	app.preview.Move(45, 0)
	app.preview.Resize(width-45, height-6)

	app.help.Move(0, height-1)
	app.help.Resize(width, 1)

	passwordWidth := min(width, 40)
	app.passwordInput.Resize(passwordWidth, 3)
	app.passwordInput.Move((width-passwordWidth)/2, (height-3)/2)
}

func (app *App) HandleEvent(ev t.Event) bool {
	switch ev := ev.(type) {
	case *t.EventResize:
		app.Resize(ev.Size())
		app.screen.Sync()

	case *t.EventKey:
		switch ev.Key() {
		case t.KeyRune:
			switch ev.Rune() {
			case '1':
				app.setFocus(app.calendar)
				return true
			case '2':
				app.setFocus( app.tags)
				return true
			case '3':
				app.setFocus( app.logs)
				return true
			case '4':
				app.setFocus(app.preview)
				return true
			case 'd':
				if app.focusedComp == app.calendar {
					if hasEntry, _ := app.journal.HasEntry(app.calendar.DayUnderCursor()); hasEntry {
						app.setFocus(app.confirmDelete)
						return true
					}
				}
			case 'q':
				app.quit(nil)
			}
		case t.KeyCtrlC:
			app.quit(nil)
		case t.KeyTab:
			switch app.focusedComp {
			case app.calendar:
				app.setFocus(app.tags)
			case app.tags:
				app.setFocus(app.logs)
			case app.logs:
				app.setFocus(app.preview)
			case app.preview:
				app.setFocus(app.calendar)
			}
		}
	}

	consumed := app.focusedComp.HandleEvent(ev)
	return consumed
}

func (app *App) setFocus(comp components.Component) {
	app.focusedComp = comp
	app.updateHelpText()
}

func (app *App) Render() {
	app.screen.Clear()

	if !app.journal.IsMounted() {
		app.setFocus(app.passwordInput)
	}

	render.Box(app.screen, 0, 0, 45, 3, render.RoundedBorders, theme.Border)
	app.screen.PutStr(1, 1, "Journal v0.1.0")

	children := []components.Component{app.calendar, app.tags, app.logs, app.preview, app.help}
	for _, comp := range children {
		app.screen.HideCursor()
		comp.Render(app.screen, app.focusedComp == comp)
	}

	switch app.focusedComp {
	case app.passwordInput:
		app.passwordInput.Render(app.screen, true)
	case app.confirmDelete:
		app.confirmDelete.Render(app.screen, true)
	}

	app.screen.Show()
}

func (app *App) updateTags() {
	changed := app.tags.RefreshTags()
	if changed {
		app.tags.Render(app.screen, app.focusedComp == app.tags)
		app.screen.Show()
	}
}

func (app *App) updatePreview(day, month, year int) {
	if !app.journal.IsMounted() {
		return
	}

	hasEntry, err := app.journal.HasEntry(day, month, year)
	if err != nil {
		log.Println(err)
		return
	}

	if hasEntry {
		filepath := app.journal.EntryPath(day, month, year)
		content, err := os.ReadFile(filepath)
		if err != nil {
			log.Println(err)
		} else {
			app.preview.SetContent(string(content))
		}
	} else {
		app.preview.SetContent("[No entry]")
	}
}

func (app *App) updateHelpText() {
	switch app.focusedComp {
	case app.calendar:
		app.help.SetText("Select day: ⬍/⬌ | Edit: <ENTER> | Delete: d | Today: t | Next/Previous month: n/p | Exit: q")
	case app.tags:
		app.help.SetText("Select: ⬍ | View entries: <ENTER>")
	case app.logs:
		app.help.SetText("Select: ⬍ | Clear: c")
	case app.preview:
		app.help.SetText("Scroll: ⬍")
	case app.confirmDelete:
		app.help.SetText("Delete: y | Keep: n/<ESC>")
	case app.passwordInput:
		app.help.SetText("Submit: <enter>")
	}
}

func (app *App) mountJournal(password string) {
	if app.journal.IsMounted() {
		return
	}

	err := app.journal.Mount(password)
	if err != nil {
		log.Println(err)
		return
	}

	app.setFocus(app.calendar)
	app.passwordInput.SetValue("")
	app.screen.HideCursor()

	time.Sleep(MountWaitTime)
	app.updateTags()
	app.updatePreview(app.calendar.DayUnderCursor())
	app.Render()
}

func (app *App) onJournalUnmonut() {
	app.preview.SetContent("")
	app.tags.RefreshTags()
	app.setFocus(app.passwordInput)
	app.Render()
}

func (app *App) quit(reason error) {
	log.SetOutput(os.Stdout)
	app.screen.Fini()

	if reason != nil {
		log.Println(reason)
	}

	err := app.journal.Unmount()
	if err != nil {
		log.Println(err)
	}

	os.Exit(0)
}

type LogWriter struct {
	callback func(msg string)
}

var _ io.Writer = (*LogWriter)(nil)

func (w *LogWriter) Write(data []byte) (int, error) {
	msg := strings.TrimSpace(string(data))
	w.callback(msg)
	return len(data), nil
}
