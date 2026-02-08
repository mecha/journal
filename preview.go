package main

import (
	"log"
	"strings"
	"time"

	c "journal-tui/components"
	"journal-tui/theme"

	t "github.com/gdamore/tcell/v2"
)

type Preview struct {
	journal *Journal
	text    *c.Text
}

func CreatePreview(journal *Journal) *Preview {
	text := c.NewText([]string{})
	return &Preview{journal, text}
}

func (p *Preview) HandleEvent(ev t.Event) bool {
	return p.text.HandleEvent(ev)
}

func (p *Preview) Render(r c.Renderer, hasFocus bool) {
	region := c.DrawPanel(r, "[3]─Preview", theme.Borders(hasFocus))
	p.text.Render(region, hasFocus)
}

func (p *Preview) Update(date time.Time) {
	if p.journal.IsMounted() {
		entry, has, err := p.journal.GetEntry(date)
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
