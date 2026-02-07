package components

import (
	"slices"

	t "github.com/gdamore/tcell/v2"
)

type FocusManager struct {
	content        Component
	quicklist      []Component
	stack          []Component
	onFocusChangedFn func(current Component)
}

func NewFocusManager(content Component, quicklist []Component) *FocusManager {
	return &FocusManager{
		content:   content,
		quicklist: quicklist,
		stack:     []Component{},
	}
}

func (fm *FocusManager) Current() Component {
	if len(fm.stack) == 0 {
		return nil
	}
	return fm.stack[len(fm.stack)-1]
}

// Switches current focus to a new component, preserving previous stack entries
func (fm *FocusManager) SwitchTo(c Component) {
	if len(fm.stack) == 0 {
		fm.stack = append(fm.stack, c)
	} else {
		fm.stack[len(fm.stack)-1] = c
	}
	fm.onFocusChangedFn(c)
}

// Pushes focus onto the stack, to return to the previous component later
func (fm *FocusManager) Push(c Component) {
	fm.stack = append(fm.stack, c)
	fm.onFocusChangedFn(c)
}

// Pops focus from the stack, returning to the previous component
func (fm *FocusManager) Pop() Component {
	if len(fm.stack) == 0 {
		return nil
	}
	last := fm.stack[len(fm.stack)-1]
	fm.stack = fm.stack[:len(fm.stack)-1]
	fm.onFocusChangedFn(fm.Current())
	return last
}

func (fm *FocusManager) OnFocusChanged(fn func(current Component)) *FocusManager {
	fm.onFocusChangedFn = fn
	return fm
}

func (fm *FocusManager) HandleEvent(ev t.Event) bool {
	current := fm.Current()

	if ev, ok := ev.(*t.EventKey); ok {
		switch key := ev.Key(); key {
		case t.KeyRune:
			rune := ev.Rune()
			index := int(rune - '1')
			if index >= 0 && index < len(fm.quicklist) {
				fm.SwitchTo(fm.quicklist[index])
				return true
			}

		case t.KeyTab, t.KeyBacktab:
			currIdx := slices.Index(fm.quicklist, current)
			if currIdx >= 0 {
				var nextNum int
				if key == t.KeyTab {
					nextNum = (currIdx + 1) % len(fm.quicklist)
				} else {
					nextNum = (currIdx - 1 + len(fm.quicklist)) % len(fm.quicklist)
				}
				fm.SwitchTo(fm.quicklist[nextNum])
				return true
			}
		}
	}

	if current == nil {
		return false
	}

	consumed := current.HandleEvent(ev)
	return consumed
}

func (fm *FocusManager) Render(screen t.Screen, region Rect, hasFocus bool) {
	fm.content.Render(screen, region, hasFocus)
}
