package main

import (
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

// TODO: need something to encapsulate focus logic (num key to focus specific comp, tab to cycle, help text updates to match focus)

var (
	screen      t.Screen
	journal     *j.Journal
	focusedComp c.Component
	logWriter   io.Writer = &LogWriter{}
)

// components
var (
	layout *c.Layout

	titlePanel *c.Panel

	calendar *c.Calendar

	tagsPanel *c.Panel
	tagsMux   *c.Mux
	tagsList  *c.List

	previewPanel *c.Panel
	preview      *c.Markdown

	logsPanel *c.Panel
	logsList   *c.List

	help *c.Text

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

	// TODO: put calendar in a panel
	calendar = c.NewCalendar(journal).OnDayChanged(updatePreview)

	tagsList = c.NewList([]string{})
	tagsPanel = c.NewPanel("[2]─Tags", tagsList)
	tagsMux = c.NewMux([]c.Component{tagsPanel})

	preview = c.NewMarkdownComponent("")
	previewPanel = c.NewPanel("[4]─Preview", preview)

	logsList = c.NewList([]string{})
	logsPanel = c.NewPanel("[3]─Log", logsList)

	help = c.NewText("").Style(theme.Help)

	confirmDelToggle = c.NewFocusToggle(
		c.NewConfirm("Are you sure you want to delete this journal entry?", func(accepted bool) {
			if accepted {
				journal.DeleteEntry(calendar.DayUnderCursor())
			}
			setFocus(calendar)
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
			const logsH = 3
			const helpH = 1
			previewH := h - logsH - helpH
			tagsH := h - titleH - calH - logsH - helpH
			return map[c.Rect]c.Component{
				c.NewRect(x, y, calW, titleH):          titlePanel,
				c.NewRect(x, 3, calW, calH):            calendar,
				c.NewRect(x, 18, calW, tagsH):          tagsMux,
				c.NewRect(x+calW, y, w-calW, previewH): previewPanel,
				c.NewRect(x, h-logsH-1, w, logsH):      logsPanel,
				c.NewRect(x, h-helpH, w, helpH):        help,

				region: confirmDelToggle,
				c.CenterFit(region, c.NewSize(min(w, 40), 3)): pswdInputToggle,
			}
		},
	).WithFocus(func() c.Component { return focusedComp })

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
	updateHelpText()
}

func handleEvent(ev t.Event) {
	switch ev := ev.(type) {
	case *t.EventResize:
		screen.Sync()

	case *t.EventKey:
		switch ev.Key() {
		case t.KeyRune:
			switch ev.Rune() {
			case '1':
				setFocus(calendar)
			case '2':
				setFocus(tagsMux)
			case '3':
				setFocus(logsPanel)
			case '4':
				setFocus(previewPanel)
			case 'd':
				if focusedComp == calendar {
					d, m, y := calendar.DayUnderCursor()
					if hasEntry, _ := journal.HasEntry(d, m, y); hasEntry {
						setFocus(confirmDelToggle)
					}
				}
			case 'q':
				quit(nil)
			}
		case t.KeyCtrlC:
			quit(nil)
		case t.KeyTab:
			switch focusedComp {
			case calendar:
				setFocus(tagsMux)
			case tagsMux:
				setFocus(logsPanel)
			case logsPanel:
				setFocus(previewPanel)
			case previewPanel:
				setFocus(calendar)
			}
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
	if !journal.IsMounted() {
		return
	}

	hasEntry, err := journal.HasEntry(day, month, year)
	if err != nil {
		log.Println(err)
		return
	}

	if hasEntry {
		filepath := journal.EntryPath(day, month, year)
		content, err := os.ReadFile(filepath)
		if err != nil {
			log.Println(err)
		} else {
			preview.SetContent(string(content))
		}
	} else {
		preview.SetContent("[No entry]")
	}
}

func updateHelpText() {
	switch focusedComp {
	case calendar:
		help.SetText("Select day: ⬍/⬌ | Edit: <ENTER> | Delete: d | Today: t | Next/Previous month: n/p | Exit: q")
	case tagsMux:
		help.SetText("Select: ⬍ | View entries: <ENTER>")
	case logsPanel:
		help.SetText("Select: ⬍ | Clear: c")
	case previewPanel:
		help.SetText("Scroll: ⬍")
	case confirmDelToggle:
		help.SetText("Delete: y | Keep: n/<ESC>")
	case pswdInputToggle:
		help.SetText("Submit: <enter>")
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

	setFocus(calendar)
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
