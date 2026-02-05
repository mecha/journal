package components

import (
	"fmt"
	"journal-tui/journal"
	"slices"

	t "github.com/gdamore/tcell/v2"
)

type TagList struct {
	journal *journal.Journal
	list    *ListComponent
}

var _ Component = (*TagList)(nil)

func NewTagList(journal *journal.Journal, x, y int) *TagList {
	return &TagList{
		journal: journal,
		list: NewListComponent(
			fmt.Sprintf("[2]─%s", "Tags"),
			x, y,
			45, 8,
		),
	}
}

func (t *TagList) Move(x, y int)   { t.list.Move(x, y) }
func (t *TagList) Resize(w, h int) { t.list.Resize(w, h) }

func (t *TagList) RefreshTags() bool {
	if t.journal.IsMounted() {
		tags, _ := t.journal.Tags()
		slices.Sort(tags)
		changed := t.list.SetItems(tags)
		return changed
	} else {
		t.list.SetItems([]string{})
		return false
	}
}

func (t *TagList) HandleEvent(ev t.Event) (consumed bool) {
	return t.list.HandleEvent(ev)
}

func (t *TagList) Render(screen t.Screen, hasFocus bool) {
	t.list.Render(screen, hasFocus)
}
