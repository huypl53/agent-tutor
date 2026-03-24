package tui

import "github.com/charmbracelet/lipgloss"

// Styles holds all TUI visual styles.
type Styles struct {
	FocusedBorder   lipgloss.Style
	UnfocusedBorder lipgloss.Style
	StatusBar       lipgloss.Style
}

// DefaultStyles returns the default TUI styles.
func DefaultStyles() Styles {
	return Styles{
		FocusedBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39")),
		UnfocusedBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")),
		StatusBar: lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 1),
	}
}
