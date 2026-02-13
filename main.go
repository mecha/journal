package main

import (
	"log"
	"os"

	c "github.com/mecha/journal/components"

	t "github.com/gdamore/tcell/v2"
	"gopkg.in/fsnotify.v1"
)

const Version = "0.1.0"

var screen t.Screen

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
		app.showPreview(app.date)
		app.updateTags()
		screen.PostEvent(NewRerenderEvent())
	})

	journal.OnUnmount(func() {
		screen.PostEvent(NewRerenderEvent())
	})

	log.SetOutput(&AppLogWriter{app})

	defer func() {
		err := recover()
		log.SetOutput(os.Stdout)
		screen.Fini()
		if err != nil {
			log.Fatal(err)
		}
	}()

	renderer := c.NewScreenRenderer(screen)

	var handler c.EventHandler = nil

	for {
		ev := screen.PollEvent()

		if handler == nil || !handler(ev) {
			switch ev := ev.(type) {
			case *t.EventResize:
				screen.Sync()
			case *t.EventKey:
				if ev.Key() == t.KeyCtrlC || ev.Rune() == 'q' {
					journal.Unmount()
					return
				}
			}
		}

		screen.Clear()
		screen.HideCursor()
		handler = DrawApp(renderer, app)
		screen.Show()
	}
}

// Creates a simple time event, used to trigger a re-render.
func NewRerenderEvent() t.Event {
	ev := &t.EventTime{}
	ev.SetEventNow()
	return ev
}
