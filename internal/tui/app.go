package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/huypl53/agent-tutor/internal/config"
	"github.com/huypl53/agent-tutor/internal/tmux"
)

type tickMsg struct{}

// Model is the top-level bubbletea model for the dual-pane TUI.
type Model struct {
	panes      [2]*PaneModel
	activePane int
	statusBar  *StatusBar
	keyMap     KeyMap
	styles     Styles
	tm         *tmux.Manager
	width      int
	height     int
	layout     string
	splitRatio int
	ready      bool
}

// New creates a new TUI Model wired to the given tmux manager and config.
func New(tm *tmux.Manager, cfg *config.Config, sessName string) Model {
	polling := cfg.TUI.Polling
	pane0 := NewPaneModel(tm, "0", "User Terminal", polling.ActiveMs, polling.IdleMs, polling.IdleThresholdS)
	pane1 := NewPaneModel(tm, "1", "Claude Code", polling.ActiveMs, polling.IdleMs, polling.IdleThresholdS)

	km := KeyMapFromConfig(cfg.TUI.FocusKey, cfg.TUI.QuitKey)
	sb := NewStatusBar(sessName, cfg.GetIntensity())
	sb.SetFocusHint(cfg.TUI.FocusKey)
	sb.SetActivePane(0, "User Terminal")

	layout := cfg.TUI.Layout
	if layout == "" {
		layout = "horizontal"
	}
	splitRatio := cfg.TUI.SplitRatio
	if splitRatio <= 0 || splitRatio >= 100 {
		splitRatio = 50
	}

	return Model{
		panes:      [2]*PaneModel{pane0, pane1},
		activePane: 0,
		statusBar:  sb,
		keyMap:     km,
		styles:     DefaultStyles(),
		tm:         tm,
		layout:     layout,
		splitRatio: splitRatio,
	}
}

// Init returns the initial command that starts the tick loop and clears the screen.
func (m Model) Init() tea.Cmd {
	return tea.Batch(tick(m.panes[0].TickInterval()), tea.ClearScreen)
}

func tick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

// Update handles messages and returns the updated model and any commands.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.syncTmuxPaneSizes()
		return m, nil

	case tickMsg:
		for i := range m.panes {
			_ = m.panes[i].Capture()
		}
		i0 := m.panes[0].TickInterval()
		i1 := m.panes[1].TickInterval()
		interval := i0
		if i1 < i0 {
			interval = i1
		}
		return m, tick(interval)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keyMap.FocusNext):
			m.activePane = 1 - m.activePane
			label := m.panes[m.activePane].Label()
			m.statusBar.SetActivePane(m.activePane, label)
			return m, nil
		default:
			tmuxKeys := KeyToTmux(msg)
			if tmuxKeys != nil {
				_ = m.panes[m.activePane].SendKeys(tmuxKeys...)
				m.panes[m.activePane].MarkActive()
			}
			return m, nil
		}
	}

	return m, nil
}

// View renders the full TUI: two panes plus a status bar.
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	statusHeight := 1
	// 4 accounts for top+bottom borders of 2 pane boxes
	availHeight := m.height - statusHeight - 4

	pane0Content := m.renderPane(0, availHeight)
	pane1Content := m.renderPane(1, availHeight)

	var panes string
	if m.layout == "vertical" {
		panes = lipgloss.JoinVertical(lipgloss.Left, pane0Content, pane1Content)
	} else {
		panes = lipgloss.JoinHorizontal(lipgloss.Top, pane0Content, pane1Content)
	}

	status := m.renderStatusBar()
	return lipgloss.JoinVertical(lipgloss.Left, panes, status)
}

func (m Model) paneDimensions(idx, availHeight int) (width, height int) {
	if m.layout == "vertical" {
		width = m.width - 2
		first := (availHeight * m.splitRatio) / 100
		if idx == 0 {
			height = first
		} else {
			height = availHeight - first
		}
	} else {
		first := (m.width * m.splitRatio) / 100
		if idx == 0 {
			width = first - 2
		} else {
			width = m.width - first - 2
		}
		height = availHeight
	}
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	return
}

func (m Model) renderPane(idx int, availHeight int) string {
	pane := m.panes[idx]
	style := m.styles.UnfocusedBorder
	if idx == m.activePane {
		style = m.styles.FocusedBorder
	}

	paneWidth, paneHeight := m.paneDimensions(idx, availHeight)

	content := m.fitContent(pane.Content(), paneWidth, paneHeight)
	title := fmt.Sprintf(" %s ", pane.Label())
	return style.Width(paneWidth).Height(paneHeight).Render(title + "\n" + content)
}

func (m Model) fitContent(content string, width, height int) string {
	lines := strings.Split(content, "\n")
	visible := height - 1
	if visible < 0 {
		visible = 0
	}
	if len(lines) > visible {
		lines = lines[len(lines)-visible:]
	}
	for i, line := range lines {
		if ansi.StringWidth(line) > width {
			lines[i] = ansi.Truncate(line, width, "")
		}
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderStatusBar() string {
	return m.styles.StatusBar.Width(m.width).Render(m.statusBar.RenderPlain())
}

func (m Model) syncTmuxPaneSizes() {
	statusHeight := 1
	// 4 accounts for top+bottom borders of 2 pane boxes
	availHeight := m.height - statusHeight - 4

	for i := range m.panes {
		pw, ph := m.paneDimensions(i, availHeight)
		if pw > 0 && ph > 0 {
			_ = m.tm.ResizePane(m.panes[i].targetID, pw, ph)
		}
	}
}
