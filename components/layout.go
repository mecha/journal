package components

import t "github.com/gdamore/tcell/v2"

type Layout struct {
	provider LayoutProvider
	focus    FocusProvider
}

type LayoutProvider func(screen t.Screen, region Rect, hasFocus bool) map[Rect]Component

var _ Component = (*Layout)(nil)

func NewLayout(provider LayoutProvider) *Layout {
	return &Layout{provider, func() Component { return nil }}
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
	for region, child := range c.provider(screen, region, hasFocus) {
		child.Render(screen, region, child == focus)
	}
}
