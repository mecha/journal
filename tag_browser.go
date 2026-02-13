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
	journal        *Journal
	tags           []string
	files          []time.Time
	tagListState   *c.ListState[string]
	fileListState  *c.ListState[time.Time]
	isShowingFiles bool
	handler        c.EventHandler

	fileList      *c.List[time.Time]
	updatePreview func(time.Time)
	resetPreview  func()
}

func CreateTagBrowser(journal *Journal, updatePreview func(time.Time), resetPreview func()) *TagBrowser {
	b := &TagBrowser{
		journal:       journal,
		tags:          []string{},
		tagListState:  &c.ListState[string]{},
		fileListState: &c.ListState[time.Time]{},
		resetPreview:  resetPreview,
		updatePreview: updatePreview,
	}

	return b
}

func (b *TagBrowser) UpdateTags() {
	if b.journal.IsMounted() {
		b.tags, _ = b.journal.Tags()
		slices.Sort(b.tags)
		// b.tagList.SetItems(b.tags)
	} else {
		b.tags = []string{}
		// b.tagList.SetItems([]string{})
	}
}

func (b *TagBrowser) HandleEvent(ev t.Event) bool {
	if b.handler != nil && b.handler(ev) {
		return true
	}

	switch ev := ev.(type) {
	case *t.EventKey:
		switch ev.Key() {
		case t.KeyEscape:
			b.isShowingFiles = false
			b.resetPreview()
			return true
		}
	}

	return false
}

func (b *TagBrowser) Render(r c.Renderer, hasFocus bool) {
	title := "[2]─Tags"
	if b.isShowingFiles {
		title += " > References"
	}

	region := c.DrawPanel(r, title, theme.Borders(hasFocus))

	if b.isShowingFiles {
		b.handler = c.DrawList(region, b.fileListState, c.ListProps[time.Time]{
			Items:        b.files,
			ShowSelected: hasFocus,
			RenderFunc: func(item time.Time) string {
				return item.Format("02 Jan 2006")
			},
			OnSelect: func(i int, item time.Time) {
				b.updatePreview(item)
			},
			OnEnter: func(i int, item time.Time) {
				err := b.journal.EditEntry(item)
				if err != nil {
					log.Print(err)
				}
				b.isShowingFiles = false
				b.resetPreview()
			},
		})
	} else {
		b.handler = c.DrawList(region, b.tagListState, c.ListProps[string]{
			Items:        b.tags,
			ShowSelected: hasFocus,
			RenderFunc:   func(tag string) string { return tag },
			OnEnter: func(i int, tag string) {
				entries, err := b.journal.SearchTag(tag)
				if err != nil {
					log.Print(err)
				}
				b.files = entries
				b.isShowingFiles = true
			},
		})
	}
}
