package main

import (
	"log"
	"slices"
	"time"

	c "journal-tui/components"
	j "journal-tui/journal"

	t "github.com/gdamore/tcell/v2"
)

type DayCallback func(day, month, year int)

type TagBrowser struct {
	journal      *j.Journal
	tagList      *c.List[string]
	fileList     *c.List[CalendarDay]
	selectedTag  string
	resetPreview func()
}

type CalendarDay struct{ day, month, year int }

func CreateTagBrowser(journal *j.Journal, updatePreview DayCallback, resetPreview func()) *TagBrowser {
	b := &TagBrowser{
		journal:      journal,
		tagList:      c.NewList([]string{}),
		fileList:     c.NewList([]CalendarDay{}),
		selectedTag:  "",
		resetPreview: resetPreview,
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
		OnSelect(func(i int, item CalendarDay) {
			updatePreview(item.day, item.month, item.year)
		}).
		OnEnter(func(i int, item CalendarDay) {
			err := b.journal.EditEntry(item.day, item.month, item.year)
			if err != nil {
				log.Print(err)
			}
			b.selectedTag = ""
			b.resetPreview()
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
			b.resetPreview()
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
