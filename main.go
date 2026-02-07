package main

import (
	"fmt"
	"io"
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
	screen         t.Screen
	journal        *j.Journal
	focusedComp    c.Component
	focusableList  []c.Component
	helpStrings    map[c.Component]string
	previewContent string
	logsContent    string
	logWriter      io.Writer = &LogWriter{}
)

// components
var (
	layout *c.Layout

	titlePanel *c.Panel

	calendarPanel *c.Panel
	calendar      *c.Calendar

	tagsPanel    *c.Panel
	tagsMux      *c.Mux
	tagsList     *c.List[string]
	tagsFileList *c.List[CalendarDay]

	previewPanel *c.Panel
	preview      *c.Text

	logsPanel *c.Panel
	logs      *c.Text

	helpbar *c.Text

	confirmDelToggle *c.FocusToggle
	confirmDel       *c.Confirm

	pswdInputToggle *c.FocusToggle
)

type CalendarDay struct{ day, month, year int }

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

	journal = j.NewJournal(Flags.path, Flags.mntPath, Flags.idleTimeout)
	journal.OnUnmount(onJournalUnmount)

	titlePanel = c.NewPanel("", c.NewTextScroller([]string{"Journal v0.1.0"}))

	calendar = c.NewCalendar().
		UnderlineDay(func(day, month, year int) bool {
			hasEntry, _ := journal.HasEntry(day, month, year)
			return hasEntry
		}).
		OnDayChanged(func(day, month, year int) {
			updateCalendarPanelTitle(month, year)
			updatePreview(day, month, year)
		})
	calendarPanel = c.NewPanel("", c.NewKeyHandler(calendar, func(ev *t.EventKey) bool {
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
					setFocus(confirmDelToggle)
				}
				return true
			}
		}
		return false
	}))
	_, month, year := calendar.DayUnderCursor()
	updateCalendarPanelTitle(month, year)

	tagsFileList = c.NewList([]CalendarDay{}).
		RenderWith(func(item CalendarDay) string {
			date := time.Date(item.year, time.Month(item.month), item.day, 0, 0, 0, 0, time.Local)
			return date.Format("02 Jan 2006")
		}).
		OnEnter(func(i int, item CalendarDay) {
			tagsMux.SwitchTo(0)
			err := journal.EditEntry(item.day, item.month, item.year)
			if err != nil {
				log.Print(err)
			}
			updatePreview(calendar.DayUnderCursor())
		})
	tagsList = c.NewList([]string{}).OnEnter(func(i int, item string) {
		files, err := journal.SearchTag(item)
		if err != nil {
			log.Print(err)
		}
		items := []CalendarDay{}
		for _, file := range files {
			day, month, year := journal.GetEntryAtPath(file)
			if year == 0 {
				continue
			}
			items = append(items, CalendarDay{day, month, year})
		}
		tagsFileList.SetItems(items)
		tagsMux.SwitchTo(1)
	})
	tagsMux = c.NewMux([]c.Component{
		c.NewPanel("[2]─Tags", tagsList),
		c.NewPanel("[2]─Tags > References",
			c.NewKeyHandler(tagsFileList, func(ev *t.EventKey) bool {
				if ev.Key() == t.KeyEscape {
					tagsMux.SwitchTo(0)
					return true
				}
				return false
			}),
		),
	})

	preview = c.NewTextScroller([]string{})
	previewPanel = c.NewPanel("[3]─Preview", preview)

	logs = c.NewTextScroller([]string{})
	logsPanel = c.NewPanel("[4]─Log", logs)

	helpbar = c.NewTextScroller([]string{}).SetStyle(theme.Help)

	confirmDelToggle = c.NewFocusToggle(
		c.NewConfirm("Are you sure you want to delete this journal entry?", func(accepted bool) {
			if accepted {
				journal.DeleteEntry(calendar.DayUnderCursor())
				updatePreview(calendar.DayUnderCursor())
			}
			setFocus(calendarPanel)
		}),
	)

	pswdInputToggle := c.NewFocusToggle(
		c.NewPanel("Password", c.NewInputComponent(mountJournal).SetMask('*').ClearOnEnter(true)),
	)
	setFocus(pswdInputToggle)

	layout = c.NewLayout(
		func(screen t.Screen, region c.Rect, hasFocus bool) []c.LayoutTile {
			x, y, w, h := region.XYWH()
			const titleH = 3
			const calW = 45
			const calH = 15
			const helpH = 1
			logsH := 6
			if focusedComp == logsPanel {
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
				c.NewLayoutTile(region, confirmDelToggle),
				c.NewLayoutTile(c.CenterFit(region, c.NewSize(min(w, 40), 3)), pswdInputToggle),
			}
		},
	).WithFocus(func() c.Component { return focusedComp })

	focusableList = []c.Component{calendarPanel, tagsMux, previewPanel, logsPanel}

	helpStrings = map[c.Component]string{
		calendarPanel:    "Select day: ⬍/⬌ | Edit: <ENTER> | Delete: d | Today: t | Next/Previous month: n/p | Exit: q",
		tagsMux:          "Select: ⬍ | View entries: <ENTER>",
		logsPanel:        "Select: ⬍ | Clear: c",
		previewPanel:     "Scroll: ⬍",
		confirmDelToggle: "Delete: y | Keep: n/<ESC>",
		pswdInputToggle:  "Submit: <enter>",
	}

	go func() {
		for range time.NewTicker(3 * time.Second).C {
			updateTags()
		}
	}()

	defer func() {
		recover()
		log.SetOutput(os.Stdout)
		journal.Unmount()
	}()

	log.SetOutput(logWriter)
	for {
		ev := screen.PollEvent()
		handleEvent(ev)
		renderScreen()
	}
}

func setFocus(comp c.Component) {
	focusedComp = comp
	helpbar.SetLines([]string{helpStrings[comp]})
}

func handleEvent(ev t.Event) {
	switch ev := ev.(type) {
	case *t.EventResize:
		screen.Sync()

	case *t.EventKey:
		switch key := ev.Key(); key {
		case t.KeyRune:
			rune := ev.Rune()
			switch rune {
			case '1', '2', '3', '4', '5', '6', '7', '8', '9':
				num := int(rune - '1')
				if num >= 0 && num < len(focusableList) {
					setFocus(focusableList[num])
				}
				return
			case 'q':
				quit(nil)
			}
		case t.KeyCtrlC:
			quit(nil)
		case t.KeyTab, t.KeyBacktab:
			currNum := slices.Index(focusableList, focusedComp)
			if currNum >= 0 {
				var nextNum int
				if key == t.KeyTab {
					nextNum = (currNum + 1) % len(focusableList)
				} else {
					nextNum = (currNum - 1 + len(focusableList)) % len(focusableList)
				}
				nextComp := focusableList[nextNum]
				setFocus(nextComp)
			}
			return
		}
	}

	if focusedComp != nil {
		focusedComp.HandleEvent(ev)
	}
}

func renderScreen() {
	w, h := screen.Size()
	region := c.NewRect(0, 0, w, h)

	screen.Clear()
	layout.Render(screen, region, true)
	screen.Show()
}

func updateCalendarPanelTitle(month, year int) {
	title := fmt.Sprintf("[1]─%s %d", time.Month(month).String(), year)
	calendarPanel.SetTitle(title)
}

func updateTags() {
	if journal.IsMounted() {
		tags, _ := journal.Tags()
		slices.Sort(tags)
		tagsList.SetItems(tags)
	} else {
		tagsList.SetItems([]string{})
	}
	renderScreen()
}

func updatePreview(day, month, year int) {
	entry, has, err := journal.GetEntry(day, month, year)
	switch {
	case err != nil:
		log.Println(err)
	case has:
		preview.SetLines(strings.Split(entry, "\n"))
	default:
		preview.SetLines([]string{"[No entry]"})
	}
}

func mountJournal(password string) {
	if journal.IsMounted() {
		return
	}

	err := journal.Mount(password)
	if err != nil {
		log.Println("failed to unlock journal; ", err)
		return
	}

	setFocus(calendarPanel)
	screen.HideCursor()

	updateTags()
	updatePreview(calendar.DayUnderCursor())
	renderScreen()
}

func onJournalUnmount() {
	preview.SetLines([]string{"[Journal is locked]"})
	setFocus(pswdInputToggle)
	updateTags()
	renderScreen()
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

type LogWriter struct{}

func (w *LogWriter) Write(data []byte) (int, error) {
	newLines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	logs.AddLines(newLines)
	logs.ScrollToBottom()
	renderScreen()
	return len(data), nil
}
