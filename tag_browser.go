package main

import (
	"log"
	"slices"
	"time"

	c "github.com/mecha/journal/components"
	"github.com/mecha/journal/theme"

	t "github.com/gdamore/tcell/v2"
)

type TagBrowser struct {
	journal      *Journal
	tagList      *c.List[string]
	fileList     *c.List[time.Time]
	selectedTag  string
	resetPreview func()
}

func CreateTagBrowser(journal *Journal, updatePreview func(time.Time), resetPreview func()) *TagBrowser {
	b := &TagBrowser{
		journal:      journal,
		tagList:      c.NewList([]string{}),
		fileList:     c.NewList([]time.Time{}),
		selectedTag:  "",
		resetPreview: resetPreview,
	}

	b.tagList.
		OnEnter(func(i int, tag string) {
			entries, err := b.journal.SearchTag(tag)
			if err != nil {
				log.Print(err)
			}

			b.fileList.SetItems(entries)
			b.selectedTag = tag
		})

	b.fileList.
		RenderWith(func(item time.Time) string {
			return item.Format("02 Jan 2006")
		}).
		OnSelect(func(i int, item time.Time) {
			updatePreview(item)
		}).
		OnEnter(func(i int, item time.Time) {
			err := b.journal.EditEntry(item)
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

	region := c.DrawPanel(r, title, theme.Borders(hasFocus))

	if isShowingFiles {
		b.fileList.Render(region, hasFocus)
	} else {
		b.tagList.Render(region, hasFocus)
	}
}
