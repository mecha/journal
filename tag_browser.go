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
	state         *TagsState
	journal       *Journal
	hasFocus      bool
	onSelectRef   func(time.Time)
	onDeselectRef func()
}

type TagsState struct {
	isShowRefs bool
	tags       []string
	refs       []time.Time
	tagList    *c.ListState[string]
	refList    *c.ListState[time.Time]
}

func (state *TagsState) update(journal *Journal) {
	if journal.isMounted {
		tags, err := journal.Tags()
		if err != nil {
			log.Println(err)
		}
		slices.Sort(tags)
		state.tags = tags
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
						state.refList.Cursor = 0
						state.isShowRefs = true
						if len(entries) > 0 {
							props.onSelectRef(entries[0])
						}
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
						props.onSelectRef(item)
					},
					OnEnter: func(i int, item time.Time) {
						err := props.journal.EditEntry(item, false)
						if err != nil {
							log.Print(err)
						}
						state.isShowRefs = false
					},
				})
			}
		},
	})

	return c.Chain(
		handler,
		c.HandleKey(func(ev *t.EventKey) bool {
			switch ev.Key() {
			case t.KeyEsc:
				if state.isShowRefs {
					state.isShowRefs = false
					if props.onDeselectRef != nil {
						props.onDeselectRef()
					}
					return true
				}
			}
			switch ev.Rune() {
			case 'r':
				state.update(props.journal)
				return true
			}
			return false
		}),
	)
}
