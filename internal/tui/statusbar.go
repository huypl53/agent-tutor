package tui

import "fmt"

type StatusBar struct {
	sessionName string
	coaching    string
	activePane  int
	activeLabel string
	focusHint   string
}

func NewStatusBar(sessionName, coaching string) *StatusBar {
	return &StatusBar{
		sessionName: sessionName,
		coaching:    coaching,
		focusHint:   "ctrl+space",
	}
}

func (s *StatusBar) SetFocusHint(hint string) { s.focusHint = hint }

func (s *StatusBar) SetActivePane(pane int, label string) {
	s.activePane = pane
	s.activeLabel = label
}

func (s *StatusBar) RenderPlain() string {
	return fmt.Sprintf(" [%s]  %s  coaching: %s  %s: switch pane",
		s.activeLabel, s.sessionName, s.coaching, s.focusHint)
}
