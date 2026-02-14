package main

import (
	"log"
	"slices"
	"time"

	c "github.com/mecha/journal/components"
	"github.com/mecha/journal/theme"

	t "github.com/gdamore/tcell/v2"
)

type TagsProps struct {
	state    *TagsState
	journal  *Journal
	hasFocus bool
}

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

func TagsBrowser(r c.Renderer, props TagsProps) c.EventHandler {
	state := props.state

	title := "[2]─Tags"
	if state.isShowRefs {
		title += " > References"
	}

	handler := c.Box(r, c.BoxProps{
		Title:   title,
		Borders: c.BordersRound,
		Style:   theme.Borders(props.hasFocus),
		Children: func(r c.Renderer) c.EventHandler {
			if !state.isShowRefs {
				return c.List(r, c.ListProps[string]{
					State:        state.tagList,
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
				return c.List(r, c.ListProps[time.Time]{
					State:        state.refList,
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
		},
	})

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
