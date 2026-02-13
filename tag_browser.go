package main

import (
	"log"
	"slices"
	"time"

	c "github.com/mecha/journal/components"
	"github.com/mecha/journal/theme"

	t "github.com/gdamore/tcell/v2"
)

type TagsState struct {
	isShowRefs bool
	tags       []string
	refs       []time.Time
	tagList    *c.ListState[string]
	refList    *c.ListState[time.Time]
}

func (state *TagsState) update(journal *Journal) {
	if journal.IsMounted() {
		state.tags, _ = journal.Tags()
		slices.Sort(state.tags)
	} else {
		state.tags = []string{}
	}
}

type TagsProps struct {
	journal  *Journal
	hasFocus bool
}

func DrawTags(r c.Renderer, state *TagsState, props TagsProps) c.EventHandler {
	title := "[2]─Tags"
	if state.isShowRefs {
		title += " > References"
	}

	region := c.DrawPanel(r, title, theme.Borders(props.hasFocus))

	var handler c.EventHandler

	if !state.isShowRefs {
		handler = c.DrawList(region, state.tagList, c.ListProps[string]{
			Items:        state.tags,
			ShowSelected: props.hasFocus,
			RenderFunc:   func(tag string) string { return tag },
			OnEnter: func(i int, tag string) {
				entries, err := props.journal.SearchTag(tag)
				if err != nil {
					log.Print(err)
				}
				state.refs = entries
				state.isShowRefs = true
			},
		})
	} else {
		handler = c.DrawList(region, state.refList, c.ListProps[time.Time]{
			Items:        state.refs,
			ShowSelected: props.hasFocus,
			RenderFunc: func(item time.Time) string {
				return item.Format("02 Jan 2006")
			},
			OnSelect: func(i int, item time.Time) {
				// b.updatePreview(item)
			},
			OnEnter: func(i int, item time.Time) {
				err := props.journal.EditEntry(item)
				if err != nil {
					log.Print(err)
				}
				state.isShowRefs = false
				// b.resetPreview()
			},
		})
	}

	return func(ev t.Event) bool {
		if state.isShowRefs {
			if ev, isKey := ev.(*t.EventKey); isKey && ev.Key() == t.KeyEsc {
				state.isShowRefs = false
				return true
			}
		}
		return handler(ev)
	}
}
