package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// KeyMap holds the key bindings for the TUI.
type KeyMap struct {
	FocusNext key.Binding
	Quit      key.Binding
}

// ParseKeyBinding converts a config string like "ctrl+space" or "f1" into a
// bubbletea key.Binding.
func ParseKeyBinding(s string) key.Binding {
	// Normalise: "ctrl+space" → "ctrl+ " (bubbletea uses literal space char)
	norm := strings.ToLower(s)
	if strings.HasSuffix(norm, "+space") {
		norm = strings.TrimSuffix(norm, "space") + " "
	}
	return key.NewBinding(key.WithKeys(norm))
}

// KeyMapFromConfig builds a KeyMap from configuration strings.
func KeyMapFromConfig(focusKey, quitKey string) KeyMap {
	return KeyMap{
		FocusNext: ParseKeyBinding(focusKey),
		Quit:      ParseKeyBinding(quitKey),
	}
}

// KeyToTmux converts a bubbletea KeyMsg into the argument(s) for tmux
// send-keys. Returns nil for unhandled keys.
func KeyToTmux(msg tea.KeyMsg) []string {
	switch msg.Type {
	case tea.KeyEnter:
		return []string{"Enter"}
	case tea.KeyTab:
		return []string{"Tab"}
	case tea.KeyUp:
		return []string{"Up"}
	case tea.KeyDown:
		return []string{"Down"}
	case tea.KeyLeft:
		return []string{"Left"}
	case tea.KeyRight:
		return []string{"Right"}
	case tea.KeyEscape:
		return []string{"Escape"}
	case tea.KeyBackspace:
		return []string{"BSpace"}
	case tea.KeyDelete:
		return []string{"DC"}
	case tea.KeyHome:
		return []string{"Home"}
	case tea.KeyEnd:
		return []string{"End"}
	case tea.KeyPgUp:
		return []string{"PgUp"}
	case tea.KeyPgDown:
		return []string{"PgDn"}

	// Ctrl keys
	case tea.KeyCtrlA:
		return []string{"C-a"}
	case tea.KeyCtrlB:
		return []string{"C-b"}
	case tea.KeyCtrlC:
		return []string{"C-c"}
	case tea.KeyCtrlD:
		return []string{"C-d"}
	case tea.KeyCtrlE:
		return []string{"C-e"}
	case tea.KeyCtrlF:
		return []string{"C-f"}
	case tea.KeyCtrlG:
		return []string{"C-g"}
	// KeyCtrlH overlaps with Backspace on many terminals; handled above.
	// KeyCtrlI overlaps with Tab; handled above.
	// KeyCtrlJ overlaps with Enter on some terminals.
	case tea.KeyCtrlK:
		return []string{"C-k"}
	case tea.KeyCtrlL:
		return []string{"C-l"}
	// KeyCtrlM overlaps with Enter; handled above.
	case tea.KeyCtrlN:
		return []string{"C-n"}
	case tea.KeyCtrlO:
		return []string{"C-o"}
	case tea.KeyCtrlP:
		return []string{"C-p"}
	case tea.KeyCtrlQ:
		return []string{"C-q"}
	case tea.KeyCtrlR:
		return []string{"C-r"}
	case tea.KeyCtrlS:
		return []string{"C-s"}
	case tea.KeyCtrlT:
		return []string{"C-t"}
	case tea.KeyCtrlU:
		return []string{"C-u"}
	case tea.KeyCtrlV:
		return []string{"C-v"}
	case tea.KeyCtrlW:
		return []string{"C-w"}
	case tea.KeyCtrlX:
		return []string{"C-x"}
	case tea.KeyCtrlY:
		return []string{"C-y"}
	case tea.KeyCtrlZ:
		return []string{"C-z"}

	// Function keys
	case tea.KeyF1:
		return []string{"F1"}
	case tea.KeyF2:
		return []string{"F2"}
	case tea.KeyF3:
		return []string{"F3"}
	case tea.KeyF4:
		return []string{"F4"}
	case tea.KeyF5:
		return []string{"F5"}
	case tea.KeyF6:
		return []string{"F6"}
	case tea.KeyF7:
		return []string{"F7"}
	case tea.KeyF8:
		return []string{"F8"}
	case tea.KeyF9:
		return []string{"F9"}
	case tea.KeyF10:
		return []string{"F10"}
	case tea.KeyF11:
		return []string{"F11"}
	case tea.KeyF12:
		return []string{"F12"}

	case tea.KeyRunes:
		s := string(msg.Runes)
		if s == " " {
			return []string{"Space"}
		}
		// -l tells tmux to interpret the argument as literal text.
		return []string{"-l", s}
	}

	// F13-F20 and other uncommon keys
	if msg.Type >= tea.KeyF13 && msg.Type <= tea.KeyF20 {
		n := int(msg.Type-tea.KeyF13) + 13
		return []string{fmt.Sprintf("F%d", n)}
	}

	return nil
}
