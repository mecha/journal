package components

import t "github.com/gdamore/tcell/v2"

// A function that can handle an event.
// Returns true if the event was consumed, and false if the event was ignored.
type EventHandler func(ev t.Event) bool

// An event handler function that only handles key events.
type KeyEventHandler func(ev *t.EventKey) bool

// Utility for easily creating key event handlers
func HandleKey(handler KeyEventHandler) EventHandler {
	if handler == nil {
		return nil
	}
	return func(ev t.Event) bool {
		if ev, isKey := ev.(*t.EventKey); isKey && handler(ev) {
			return true
		}
		return false
	}
}

func Chain(handlers ...EventHandler) EventHandler {
	return func(ev t.Event) bool {
		for _, handler := range handlers {
			if handler(ev) {
				return true
			}
		}
		return false
	}
}
