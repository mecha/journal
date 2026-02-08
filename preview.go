package main

import (
	"log"
	"strings"
	"time"

	c "journal-tui/components"
	j "journal-tui/journal"

	t "github.com/gdamore/tcell/v2"
)

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
