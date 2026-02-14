package main

import (
	"log"
	"os"

	c "github.com/mecha/journal/components"

	t "github.com/gdamore/tcell/v2"
	"gopkg.in/fsnotify.v1"
)

const Version = "0.1.0"
const MinGCFSVersion = "2.6.1"

func main() {
	parseFlags()

	if err := checkGCFSVersion(MinGCFSVersion); err != nil {
		log.Fatal(err)
	}

	screen, err := t.NewScreen()
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

	triggerRender := func() {
		// any event will trigger a render, so we just use a time event
		ev := &t.EventTime{}
		ev.SetEventNow()
		screen.PostEvent(ev)
	}

	journal.onFSEvent = func(ev fsnotify.Event) {
		app.showEntryPreview(app.date)
		app.tagsList.update(journal)
		triggerRender()
	}

	journal.onUnmount = triggerRender

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
	handler := (c.EventHandler)(nil)

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
