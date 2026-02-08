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
	"journal-tui/utils"

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
	logsPanel := createLogs()

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

type DayCallback func(day, month, year int)

type Preview struct {
	panel *c.Panel
	text  *c.Text
}

type EventUpdatePreview struct {
	day, month, year int
}

func (ev *EventUpdatePreview) When() time.Time { return time.Now() } // we care when, just return now() lmao

type PreviewComp struct {
	journal *j.Journal
	panel   *c.Panel
	text    *c.Text
}

func CreatePreview(journal *j.Journal) *PreviewComp {
	text := c.NewText([]string{})

	return &PreviewComp{
		journal,
		c.NewPanel("[3]─Preview", text),
		text,
	}
}

func (p *PreviewComp) HandleEvent(ev t.Event) bool {
	return p.panel.HandleEvent(ev)
}

func (p *PreviewComp) Render(r c.Renderer, hasFocus bool) {
	p.panel.Render(r, hasFocus)
}

func (p *PreviewComp) Update(day, month, year int) {
	if p.journal.IsMounted() {
		entry, has, err := p.journal.GetEntry(day, month, year)
		switch {
		case err != nil:
			log.Println(err)
		case has:
			p.text.SetLines(strings.Split(entry, "\n"))
		default:
			p.text.SetLines([]string{"[No entry]"})
		}
	} else {
		p.text.SetLines([]string{"[Journal is locked]"})
	}
}

type DayPicker struct {
	journal       *j.Journal
	preview       *PreviewComp
	calendar      *c.Calendar
	confirmDelete *c.Confirm
	gotoPrompt    *c.InputPrompt
}

func CreateDayPicker(journal *j.Journal, preview *PreviewComp) *DayPicker {
	calendar := c.NewCalendar().
		UnderlineDay(func(day, month, year int) bool {
			hasEntry, _ := journal.HasEntry(day, month, year)
			return hasEntry
		}).
		OnDayChanged(preview.Update)

	return &DayPicker{
		journal:  journal,
		preview:  preview,
		calendar: calendar,
		gotoPrompt: c.NewInputPrompt(
			"Go to (dd/mm/yyyy)",
			c.NewInput(),
			func(input *c.Input, cancelled bool) {
				if !cancelled {
					value := input.Value()
					day, month, year, err := utils.ParseDayMonthYear(value)
					if err == nil {
						calendar.SetDay(day, month, year)
					} else {
						log.Println(err)
					}
				}
				input.SetValue("")
				Screen.HideCursor()
				Focus.Pop()
				renderScreen()
			}),
		confirmDelete: c.NewConfirm("Are you sure you want to delete this journal entry?", func(accepted bool) {
			if accepted {
				day, month, year := calendar.Current()
				journal.DeleteEntry(day, month, year)
				log.Printf("deleted entry: %s", Journal.EntryPath(day, month, year))
				preview.Update(day, month, year)
			}
			Focus.Pop()
		}),
	}
}

func (d *DayPicker) HandleEvent(ev t.Event) bool {
	if !d.journal.IsMounted() {
		return false
	}

	if Focus.Is(d.confirmDelete) && d.confirmDelete.HandleEvent(ev) {
		return true
	}

	if Focus.Is(d.gotoPrompt) && d.gotoPrompt.HandleEvent(ev) {
		return true
	}

	switch ev := ev.(type) {
	case *t.EventKey:
		switch ev.Key() {
		case t.KeyEnter:
			d.journal.EditEntry(d.calendar.Current())
			d.preview.Update(d.calendar.Current())
			return true
		case t.KeyRune:
			switch ev.Rune() {
			case 'd':
				if has, _ := d.journal.HasEntry(d.calendar.Current()); has {
					Focus.Push(d.confirmDelete)
				}
				return true
			case 'g':
				Focus.Push(d.gotoPrompt)
				return true
			}
		}
	}

	return d.calendar.HandleEvent(ev)
}

func (dp *DayPicker) Render(r c.Renderer, hasFocus bool) {
	_, month, year := dp.calendar.Current()
	title := fmt.Sprintf("[1]─%s %d", time.Month(month).String(), year)
	panelRegion := c.DrawPanel(r, title, t.StyleDefault, hasFocus)

	dp.calendar.Render(panelRegion, hasFocus)

	popupRegion := c.CenteredRegion(Screen, 40, 3)

	if Focus.Is(dp.gotoPrompt) {
		dp.gotoPrompt.Render(popupRegion, true)
	}

	if Focus.Is(dp.confirmDelete) {
		dp.confirmDelete.Render(popupRegion, true)
	}
}

type TagBrowser struct {
	journal     *j.Journal
	tagList     *c.List[string]
	fileList    *c.List[CalendarDay]
	selectedTag string
}

type CalendarDay struct{ day, month, year int }

func CreateTagBrowser(journal *j.Journal, updatePreview DayCallback, resetPreview func()) *TagBrowser {
	b := &TagBrowser{
		journal:     journal,
		tagList:     c.NewList([]string{}),
		fileList:    c.NewList([]CalendarDay{}),
		selectedTag: "",
	}

	b.tagList.
		OnEnter(func(i int, tag string) {
			files, err := b.journal.SearchTag(tag)
			if err != nil {
				log.Print(err)
			}

			items := []CalendarDay{}
			for _, file := range files {
				day, month, year := b.journal.GetEntryAtPath(file)
				if year > 0 {
					items = append(items, CalendarDay{day, month, year})
				}
			}

			b.fileList.SetItems(items)
			b.selectedTag = tag
		})

	b.fileList.
		RenderWith(func(item CalendarDay) string {
			date := time.Date(item.year, time.Month(item.month), item.day, 0, 0, 0, 0, time.Local)
			return date.Format("02 Jan 2006")
		}).
		OnEnter(func(i int, item CalendarDay) {
			err := b.journal.EditEntry(item.day, item.month, item.year)
			if err != nil {
				log.Print(err)
			}
			b.selectedTag = ""
			resetPreview()
		})

	return b
}

func (b *TagBrowser) UpdateTags() {
	if b.journal.IsMounted() {
		tags, _ := b.journal.Tags()
		slices.Sort(tags)
		b.tagList.SetItems(tags)
	} else {
		b.tagList.SetItems([]string{})
	}
}

func (b *TagBrowser) HandleEvent(ev t.Event) bool {
	switch ev := ev.(type) {
	case *t.EventKey:
		switch ev.Key() {
		case t.KeyEscape:
			b.selectedTag = ""
			return true
		}
	}

	if b.selectedTag == "" {
		return b.tagList.HandleEvent(ev)
	} else {
		return b.fileList.HandleEvent(ev)
	}
}

func (b *TagBrowser) Render(r c.Renderer, hasFocus bool) {
	isShowingFiles := len(b.selectedTag) > 0

	title := "[2]─Tags"
	if isShowingFiles {
		title += " > References"
	}

	region := c.DrawPanel(r, title, t.StyleDefault, hasFocus)

	if isShowingFiles {
		b.fileList.Render(region, hasFocus)
	} else {
		b.tagList.Render(region, hasFocus)
	}
}

func createLogs() *c.Panel {
	logText := c.NewText([]string{})

	writer := logText.Writer()
	writer.OnWrite(renderScreen)
	log.SetOutput(writer)

	return c.NewPanel("[4]─Log", logText)
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
