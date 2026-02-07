package components

import t "github.com/gdamore/tcell/v2"

type Layout struct {
	provider LayoutProvider
	focus    FocusProvider
}

type LayoutProvider func(screen t.Screen, region Rect, hasFocus bool) []LayoutTile

type LayoutTile struct {
	region  Rect
	content Component
}

type FocusProvider func() Component

var _ Component = (*Layout)(nil)

func NewLayout(provider LayoutProvider) *Layout {
	return &Layout{provider, func() Component { return nil }}
}

func NewLayoutTile(region Rect, content Component) LayoutTile {
	return LayoutTile{region, content}
}

func (c *Layout) WithFocus(provider FocusProvider) *Layout {
	c.focus = provider
	return c
}

func (c *Layout) HandleEvent(ev t.Event) bool {
	return false
}

func (c *Layout) Render(screen t.Screen, region Rect, hasFocus bool) {
	focus := c.focus()
	for _, tile := range c.provider(screen, region, hasFocus) {
		tile.content.Render(screen, tile.region, tile.content == focus)
	}
}
