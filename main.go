package main

import (
	"log"
	"os"
	"time"

	c "journal-tui/components"

	t "github.com/gdamore/tcell/v2"
)

const Version = "0.1.0"

var (
	screen  t.Screen
	journal *Journal
	focus   *c.FocusManager
	app     *App
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

	journal = NewJournal(Flags.path, Flags.mntPath, Flags.idleTimeout)

	app = CreateApp(journal)

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
		recover()
		log.SetOutput(os.Stdout)
		journal.OnUnmount(nil)
		journal.Unmount()
	}()

	go func() {
		for range time.NewTicker(3 * time.Second).C {
			app.tagBrowser.UpdateTags()
		}
	}()

	for {
		ev := screen.PollEvent()
		app.HandleEvent(ev)
		renderScreen()
	}
}

func renderScreen() {
	screen.Clear()
	app.Render(c.NewScreenRenderer(screen), true)
	screen.Show()
}

func quit(reason error) {
	log.SetOutput(os.Stdout)
	screen.Fini()

	if reason != nil {
		log.Println(reason)
	}

	err := journal.Unmount()
	if err != nil {
		log.Println(err)
	}

	os.Exit(0)
}
