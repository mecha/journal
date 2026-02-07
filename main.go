package main

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"time"

	c "journal-tui/components"
	j "journal-tui/journal"
	"journal-tui/theme"

	t "github.com/gdamore/tcell/v2"
)

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

	titlePanel := c.NewPanel("", c.NewText([]string{"Journal v0.1.0"}))

	previewPanel, updatePreview := createPreview(Journal)

	calendarPanel, calendar, confirmDelete := createCalendar(Journal, updatePreview)

	tagsMux, updateTags := createTags(Journal, calendar, updatePreview)

	logsPanel := createLogs()

	passwordInput := c.NewFocusToggle(
		c.NewPanel("Password",
			c.NewInputComponent().
				SetMask('*').
				ClearOnEnter(true).
				OnEnter(func(password string) {
					if !Journal.IsMounted() {
						err := Journal.Mount(password)
						if err != nil {
							log.Println("failed to unlock journal; ", err)
							return
						}

						Focus.Pop()
						Screen.HideCursor()

						updateTags()
						updatePreview(calendar.DayUnderCursor())
						renderScreen()
					}
				}),
		),
	)

	helpbar := c.NewText([]string{}).SetStyle(theme.Help)
	helpMap := map[c.Component]string{
		calendarPanel: "Select day: ⬍/⬌ | Edit: <ENTER> | Delete: d | Today: t | Next/Previous month: n/p | Exit: q",
		tagsMux:       "Select: ⬍ | View entries: <ENTER>",
		logsPanel:     "Select: ⬍ | Clear: c",
		previewPanel:  "Scroll: ⬍",
		confirmDelete: "Delete: y | Keep: n/<ESC>",
		passwordInput: "Submit: <enter>",
	}

	Layout = c.NewLayout(
		func(screen t.Screen, region c.Rect, hasFocus bool) []c.LayoutTile {
			x, y, w, h := region.XYWH()
			const titleH = 3
			const calW = 45
			const calH = 15
			const helpH = 1
			logsH := 6
			if Focus.Current() == logsPanel {
				logsH = min(14, h)
			}
			previewH := h - logsH - helpH
			tagsH := h - titleH - calH - logsH - helpH
			return []c.LayoutTile{
				c.NewLayoutTile(c.NewRect(x, y, calW, titleH), titlePanel),
				c.NewLayoutTile(c.NewRect(x, 3, calW, calH), calendarPanel),
				c.NewLayoutTile(c.NewRect(x, 18, calW, tagsH), tagsMux),
				c.NewLayoutTile(c.NewRect(x+calW, y, w-calW, previewH), previewPanel),
				c.NewLayoutTile(c.NewRect(x, h-logsH-1, w, logsH), logsPanel),
				c.NewLayoutTile(c.NewRect(x, h-helpH, w, helpH), helpbar),
				c.NewLayoutTile(region, confirmDelete),
				c.NewLayoutTile(c.CenterFit(region, c.NewSize(min(w, 40), 3)), passwordInput),
			}
		},
	).WithFocus(func() c.Component { return Focus.Current() })

	Focus = c.NewFocusManager(
		Layout,
		[]c.Component{
			calendarPanel,
			tagsMux,
			previewPanel,
			logsPanel,
		}).
		OnFocusChanged(func(current c.Component) {
			helpbar.SetLines([]string{helpMap[current]})
		})

	Focus.SwitchTo(calendarPanel)
	Focus.Push(passwordInput)

	Journal.OnUnmount(func() {
		Focus.Push(passwordInput)
		updateTags()
		updatePreview(0, 0, 0)
		renderScreen()
	})

	defer func() {
		recover()
		log.SetOutput(os.Stdout)
		Journal.Unmount()
	}()

	go func() {
		for range time.NewTicker(3 * time.Second).C {
			updateTags()
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
	w, h := Screen.Size()
	region := c.NewRect(0, 0, w, h)

	Screen.Clear()
	Layout.Render(Screen, region, true)
	Screen.Show()
}

type DayCallback func(day, month, year int)

func createPreview(journal *j.Journal) (*c.Panel, DayCallback) {
	previewText := c.NewText([]string{})
	previewPanel := c.NewPanel("[3]─Preview", previewText)

	updatePreview := func(day, month, year int) {
		if journal.IsMounted() {
			entry, has, err := journal.GetEntry(day, month, year)
			switch {
			case err != nil:
				log.Println(err)
			case has:
				previewText.SetLines(strings.Split(entry, "\n"))
			default:
				previewText.SetLines([]string{"[No entry]"})
			}
		} else {
			previewText.SetLines([]string{"[Journal is locked]"})
		}
	}

	return previewPanel, updatePreview
}

func createCalendar(journal *j.Journal, updatePreview DayCallback) (panel *c.Panel, calendar *c.Calendar, confirmDelete *c.FocusToggle) {
	formatTitle := func(month, year int) string {
		return fmt.Sprintf("[1]─%s %d", time.Month(month).String(), year)
	}

	calendar = c.NewCalendar().
		UnderlineDay(func(day, month, year int) bool {
			hasEntry, _ := journal.HasEntry(day, month, year)
			return hasEntry
		}).
		OnDayChanged(func(day, month, year int) {
			panel.SetTitle(formatTitle(month, year))
			updatePreview(day, month, year)
		})

	now := time.Now()
	month, year := int(now.Month()), now.Year()

	confirmDelete = c.NewFocusToggle(
		c.NewConfirm("Are you sure you want to delete this journal entry?", func(accepted bool) {
			if accepted {
				journal.DeleteEntry(calendar.DayUnderCursor())
				updatePreview(calendar.DayUnderCursor())
			}
			Focus.Pop()
		}),
	)

	panel = c.NewPanel(
		formatTitle(month, year),
		c.NewKeyHandler(calendar,
			func(ev *t.EventKey) bool {
				if !journal.IsMounted() {
					return false
				}
				switch ev.Key() {
				case t.KeyEnter:
					journal.EditEntry(calendar.DayUnderCursor())
					updatePreview(calendar.DayUnderCursor())
				case t.KeyRune:
					switch ev.Rune() {
					case 'd':
						if has, _ := journal.HasEntry(calendar.DayUnderCursor()); has {
							Focus.Push(confirmDelete)
						}
						return true
					}
				}
				return false
			}))

	return panel, calendar, confirmDelete
}

type CalendarDay struct{ day, month, year int }

func createTags(journal *j.Journal, calendar *c.Calendar, updatePreview DayCallback) (mux *c.Mux, updateTags func()) {
	fileList := c.NewList([]CalendarDay{}).
		RenderWith(func(item CalendarDay) string {
			date := time.Date(item.year, time.Month(item.month), item.day, 0, 0, 0, 0, time.Local)
			return date.Format("02 Jan 2006")
		}).
		OnEnter(func(i int, item CalendarDay) {
			err := journal.EditEntry(item.day, item.month, item.year)
			if err != nil {
				log.Print(err)
			}
			mux.SwitchTo(0)
			updatePreview(calendar.DayUnderCursor())
		})

	tagList := c.NewList([]string{}).
		OnEnter(func(i int, item string) {
			files, err := journal.SearchTag(item)
			if err != nil {
				log.Print(err)
			}

			items := []CalendarDay{}
			for _, file := range files {
				day, month, year := journal.GetEntryAtPath(file)
				if year > 0 {
					items = append(items, CalendarDay{day, month, year})
				}
			}

			fileList.SetItems(items)
			mux.SwitchTo(1)
		})

	mux = c.NewMux([]c.Component{
		c.NewPanel("[2]─Tags", tagList),
		c.NewPanel("[2]─Tags > References",
			c.NewKeyHandler(fileList,
				func(ev *t.EventKey) bool {
					if ev.Key() == t.KeyEscape {
						mux.SwitchTo(0)
						return true
					}
					return false
				}),
		),
	})

	updateTags = func() {
		if journal.IsMounted() {
			tags, _ := journal.Tags()
			slices.Sort(tags)
			tagList.SetItems(tags)
		} else {
			tagList.SetItems([]string{})
		}
	}

	return mux, updateTags
}

func createLogs() *c.Panel {
	logText := c.NewText([]string{})
	logsPanel := c.NewPanel("[4]─Log", logText)

	writer := logText.Writer()
	writer.OnWrite(renderScreen)
	log.SetOutput(writer)

	return logsPanel
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
