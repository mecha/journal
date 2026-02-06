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
	screen        t.Screen
	journal       *j.Journal
	focusedComp   c.Component
	focusableList []c.Component
	helpStrings   map[c.Component]string
	logWriter     io.Writer = &LogWriter{}
)

// components
var (
	layout *c.Layout

	titlePanel *c.Panel

	calendarPanel *c.Panel
	calendar      *c.Calendar

	tagsPanel    *c.Panel
	tagsMux      *c.Mux
	tagsList     *c.List
	tagsFileList *c.List

	previewPanel *c.Panel
	preview      *c.Markdown

	logsPanel *c.Panel
	logsList  *c.List

	helpbar *c.Text

	confirmDelToggle *c.FocusToggle
	confirmDel       *c.Confirm

	pswdInputToggle *c.FocusToggle
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

	journal = j.NewJournal(Flags.path, Flags.mntPath, Flags.idleTimeout)
	journal.OnUnmount(onJournalUnmount)

	titlePanel = c.NewPanel("", c.NewText("Journal v0.1.0"))

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

	tagsFileList = c.NewList([]string{}).OnEnter(func(i int, item string) {
		day, month, year := journal.GetEntryAtPath(item)
		if year == 0 {
			return
		}
		tagsMux.SwitchTo(0)
		err := journal.EditEntry(day, month, year)
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
		tagsFileList.SetItems(files)
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

	preview = c.NewMarkdownComponent("")
	previewPanel = c.NewPanel("[3]─Preview", preview)

	logsList = c.NewList([]string{})
	logsPanel = c.NewPanel("[4]─Log", logsList)

	helpbar = c.NewText("").Style(theme.Help)

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
		c.NewInputComponent("Password", mountJournal).SetMask('*').ClearOnEnter(true),
	)
	setFocus(pswdInputToggle)

	layout = c.NewLayout(
		func(screen t.Screen, region c.Rect, hasFocus bool) map[c.Rect]c.Component {
			x, y, w, h := region.XYWH()
			const titleH = 3
			const calW = 45
			const calH = 15
			const helpH = 1
			logsH := 3
			if focusedComp == logsPanel {
				logsH = 10
			}
			previewH := h - logsH - helpH
			tagsH := h - titleH - calH - logsH - helpH
			return map[c.Rect]c.Component{
				c.NewRect(x, y, calW, titleH):          titlePanel,
				c.NewRect(x, 3, calW, calH):            calendarPanel,
				c.NewRect(x, 18, calW, tagsH):          tagsMux,
				c.NewRect(x+calW, y, w-calW, previewH): previewPanel,
				c.NewRect(x, h-logsH-1, w, logsH):      logsPanel,
				c.NewRect(x, h-helpH, w, helpH):        helpbar,

				region: confirmDelToggle,
				c.CenterFit(region, c.NewSize(min(w, 40), 3)): pswdInputToggle,
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
	helpbar.SetText(helpStrings[comp])
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
		preview.SetContent(entry)
	default:
		preview.SetContent("[No entry]")
	}
}

const MountWaitTime = time.Millisecond * 150

func mountJournal(password string) {
	if journal.IsMounted() {
		return
	}

	err := journal.Mount(password)
	if err != nil {
		log.Println(err)
		return
	}

	setFocus(calendarPanel)
	screen.HideCursor()

	time.Sleep(MountWaitTime)
	updateTags()
	updatePreview(calendar.DayUnderCursor())
	renderScreen()
}

func onJournalUnmount() {
	preview.SetContent("")
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
	msg := strings.TrimSpace(string(data))
	line := strings.ReplaceAll(msg, "\n", " | ")
	logsList.AddItem(line)
	renderScreen()
	return len(data), nil
}
