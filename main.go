package main

import (
	"log"
	"os"
	"time"

	c "journal-tui/components"
	j "journal-tui/journal"
	"journal-tui/theme"

	t "github.com/gdamore/tcell/v2"
)

const Version = "0.1.0"

var (
	Screen  t.Screen
	Journal *j.Journal
	Focus   *c.FocusManager
	Layout  *c.Layout
)

func main() {
	parseFlags()

	Screen = initScreen()
	Journal = j.NewJournal(Flags.path, Flags.mntPath, Flags.idleTimeout)

	titlePanel := c.NewPanel("", c.NewText([]string{"Journal v" + Version}))

	preview := CreatePreview(Journal)
	dayPicker := CreateDayPicker(Journal, preview)

	resetPreview := func() {
		preview.Update(dayPicker.calendar.Current())
	}

	tagBrowser := CreateTagBrowser(Journal, preview.Update, resetPreview)
	logsPanel := CreateLogs()

	passwordInput := c.NewFocusToggle(
		c.NewInputPrompt("Password",
			c.NewInput().SetMask('*'),
			func(input *c.Input, cancelled bool) {
				if cancelled {
					return
				}

				password := input.Value()
				input.SetValue("")

				if !Journal.IsMounted() {
					err := Journal.Mount(password)
					if err != nil {
						log.Println("failed to unlock journal; ", err)
						return
					}

					log.Println("Unlocked journal")
					Focus.Pop()
					Screen.HideCursor()

					tagBrowser.UpdateTags()
					resetPreview()
					renderScreen()
				}
			},
		),
	)

	helpbar := c.NewText([]string{}).SetStyle(theme.Help())
	helpMap := map[c.Component]string{
		dayPicker:     "Select day: ⬍/⬌ | Edit: <ENTER> | Delete: d | Today: t | Go to specific day: g | Next/Previous month: n/p | Exit: q",
		tagBrowser:    "Select: ⬍ | View entries: <ENTER>",
		logsPanel:     "Select: ⬍ | Clear: c",
		preview:       "Scroll: ⬍",
		passwordInput: "Submit: <enter>",
	}

	Layout = c.NewLayout(
		func(r c.Renderer, hasFocus bool) []c.LayoutTile {
			w, h := r.Size()
			region := c.NewRect(0, 0, w, h)

			const titleHeight = 3
			const calWidth = 45
			const calHeight = 15
			const helpHeight = 1

			logsH := 6
			if Focus.Current() == logsPanel {
				logsH = min(14, h)
			}

			tagsH := h - titleHeight - calHeight - logsH - helpHeight
			previewH := h - logsH - helpHeight
			inputRect := c.CenterRect(region, min(w, 40), 3)

			return []c.LayoutTile{
				c.NewLayoutTile(c.NewRect(0, 0, calWidth, titleHeight), titlePanel),
				c.NewLayoutTile(c.NewRect(0+calWidth, 0, w-calWidth, previewH), preview),
				c.NewLayoutTile(c.NewRect(0, h-logsH-1, w, logsH), logsPanel),
				c.NewLayoutTile(c.NewRect(0, 3, calWidth, calHeight), dayPicker),
				c.NewLayoutTile(c.NewRect(0, 18, calWidth, tagsH), tagBrowser),
				c.NewLayoutTile(c.NewRect(0, h-helpHeight, w, helpHeight), helpbar),
				c.NewLayoutTile(inputRect, passwordInput),
			}
		},
	).WithFocus(func() c.Component { return Focus.Current() })

	Focus = c.NewFocusManager(
		Layout,
		[]c.Component{
			dayPicker,
			tagBrowser,
			preview,
			logsPanel,
		}).
		OnFocusChanged(func(current c.Component) {
			helpbar.SetLines([]string{helpMap[current]})
		})

	Focus.SwitchTo(dayPicker)
	Focus.Push(passwordInput)

	Journal.OnUnmount(func() {
		Focus.Push(passwordInput)
		tagBrowser.UpdateTags()
		resetPreview()
		renderScreen()
	})

	defer func() {
		recover()
		log.SetOutput(os.Stdout)
		Journal.Unmount()
	}()

	go func() {
		for range time.NewTicker(3 * time.Second).C {
			tagBrowser.UpdateTags()
		}
	}()

	for {
		ev := Screen.PollEvent()

		switch ev := ev.(type) {
		case *t.EventResize:
			Screen.Sync()

		case *t.EventKey:
			switch key := ev.Key(); key {
			case t.KeyRune:
				switch ev.Rune() {
				case 'q':
					quit(nil)
				}
			case t.KeyCtrlC:
				quit(nil)
			}
		}

		Focus.HandleEvent(ev)

		renderScreen()
	}
}

func initScreen() t.Screen {
	screen, err := t.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	if err = screen.Init(); err != nil {
		log.Fatal(err)
	}
	return screen
}

func renderScreen() {
	Screen.Clear()
	Layout.Render(Screen, true)
	Screen.Show()
}

func quit(reason error) {
	log.SetOutput(os.Stdout)
	Screen.Fini()

	if reason != nil {
		log.Println(reason)
	}

	err := Journal.Unmount()
	if err != nil {
		log.Println(err)
	}

	os.Exit(0)
}
