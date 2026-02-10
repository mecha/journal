package main

import (
	"log"
	"os"

	c "github.com/mecha/journal/components"

	t "github.com/gdamore/tcell/v2"
	"gopkg.in/fsnotify.v1"
)

const Version = "0.1.0"

var (
	screen t.Screen
	focus  *c.FocusManager
)

func main() {
	parseFlags()

	var err error
	screen, err = t.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	if err = screen.Init(); err != nil {
		log.Fatal(err)
	}

	journal, err := NewJournal(Flags.path, Flags.mntPath, Flags.idleTimeout)
	if err != nil {
		log.Fatal(err)
	}

	app := CreateApp(journal)

	journal.OnFSEvent(func(ev fsnotify.Event) {
		app.preview.Update(app.dayPicker.calendar.Date())
		app.tagBrowser.UpdateTags()
		screen.PostEvent(&t.EventTime{})
	})

	journal.OnUnmount(func() {
		app.promptForPassword()
		screen.PostEvent(&t.EventTime{})
	})

	log.SetOutput(app.logWriter())

	focus = c.NewFocusManager(
		[]c.Component{
			app.dayPicker,
			app.tagBrowser,
			app.preview,
			app.logViewer,
		})

	focus.SwitchTo(app.dayPicker)
	focus.Push(app.passwordPrompt)

	defer func() {
		err := recover()
		log.SetOutput(os.Stdout)
		screen.Fini()
		if err != nil {
			log.Fatal(err)
		}
	}()

	renderer := c.NewScreenRenderer(screen)

	for {
		ev := screen.PollEvent()

		if !app.HandleEvent(ev) {
			if ev, isKey := ev.(*t.EventKey); isKey && (ev.Key() == t.KeyCtrlC || ev.Rune() == 'q') {
				journal.Unmount()
				return
			}
		}

		screen.Clear()
		app.Render(renderer, true)
		screen.Show()
	}
}
