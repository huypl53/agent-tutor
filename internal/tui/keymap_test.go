package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestParseKeyBinding(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"ctrl+space", "ctrl+ "},
		{"ctrl+q", "ctrl+q"},
		{"ctrl+a", "ctrl+a"},
		{"f1", "f1"},
	}
	for _, tt := range tests {
		kb := ParseKeyBinding(tt.input)
		if len(kb.Keys()) == 0 {
			t.Errorf("ParseKeyBinding(%q) returned no keys", tt.input)
			continue
		}
		if kb.Keys()[0] != tt.want {
			t.Errorf("ParseKeyBinding(%q).Keys()[0] = %q, want %q", tt.input, kb.Keys()[0], tt.want)
		}
	}
}

func TestKeyToTmux(t *testing.T) {
	tests := []struct {
		name string
		msg  tea.KeyMsg
		want []string
	}{
		{"enter", tea.KeyMsg{Type: tea.KeyEnter}, []string{"Enter"}},
		{"tab", tea.KeyMsg{Type: tea.KeyTab}, []string{"Tab"}},
		{"up", tea.KeyMsg{Type: tea.KeyUp}, []string{"Up"}},
		{"down", tea.KeyMsg{Type: tea.KeyDown}, []string{"Down"}},
		{"left", tea.KeyMsg{Type: tea.KeyLeft}, []string{"Left"}},
		{"right", tea.KeyMsg{Type: tea.KeyRight}, []string{"Right"}},
		{"escape", tea.KeyMsg{Type: tea.KeyEscape}, []string{"Escape"}},
		{"backspace", tea.KeyMsg{Type: tea.KeyBackspace}, []string{"BSpace"}},
		{"delete", tea.KeyMsg{Type: tea.KeyDelete}, []string{"DC"}},
		{"ctrl+c", tea.KeyMsg{Type: tea.KeyCtrlC}, []string{"C-c"}},
		{"ctrl+d", tea.KeyMsg{Type: tea.KeyCtrlD}, []string{"C-d"}},
		{"ctrl+z", tea.KeyMsg{Type: tea.KeyCtrlZ}, []string{"C-z"}},
		{"ctrl+l", tea.KeyMsg{Type: tea.KeyCtrlL}, []string{"C-l"}},
		{"runes", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h', 'i'}}, []string{"-l", "hi"}},
		{"space", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}, []string{"Space"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := KeyToTmux(tt.msg)
			if len(got) != len(tt.want) {
				t.Fatalf("KeyToTmux(%v) = %v, want %v", tt.msg, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("KeyToTmux(%v)[%d] = %q, want %q", tt.msg, i, got[i], tt.want[i])
				}
			}
		})
	}
}
