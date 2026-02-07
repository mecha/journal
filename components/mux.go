package components

import t "github.com/gdamore/tcell/v2"

type Mux struct {
	children []Component
	current  int
}

func NewMux(children []Component) *Mux {
	return &Mux{
		children: children,
		current:  0,
	}
}

func (c *Mux) SwitchTo(n int) *Mux {
	c.current = (n + len(c.children)) % len(c.children)
	return c
}

func (c *Mux) HandleEvent(ev t.Event) bool {
	if c.current < 0 || c.current >= len(c.children) {
		panic("invalid current child state in Mux")
	}
	child := c.children[c.current]
	consumed := child.HandleEvent(ev)
	return consumed
}

func (c *Mux) Render(r Renderer, hasFocus bool) {
	if c.current < 0 || c.current >= len(c.children) {
		panic("invalid current child state in Mux")
	}
	child := c.children[c.current]
	child.Render(r, hasFocus)
}
