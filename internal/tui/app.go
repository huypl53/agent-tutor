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

// tickMsg triggers a new async capture cycle.
type tickMsg struct{}

// capturedMsg carries the results of an async pane capture.
type capturedMsg struct {
	contents [2]string
	errors   [2]error
	cursorX  int
	cursorY  int
}

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

// Init returns the initial command that starts the capture loop and clears the screen.
func (m Model) Init() tea.Cmd {
	return tea.Batch(captureAllCmd(m.panes), tea.ClearScreen)
}

// captureAllCmd runs both pane captures in a goroutine (non-blocking).
func captureAllCmd(panes [2]*PaneModel) tea.Cmd {
	return func() tea.Msg {
		var msg capturedMsg
		for i, p := range panes {
			content, err := p.tm.CapturePaneANSI(p.targetID)
			msg.contents[i] = content
			msg.errors[i] = err
		}
		// Only fetch cursor for pane 0 (user terminal)
		if x, y, err := panes[0].tm.CursorPosition(panes[0].targetID); err == nil {
			msg.cursorX = x
			msg.cursorY = y
		}
		return msg
	}
}

// sendKeysCmd forwards keys to tmux in a goroutine (non-blocking).
func sendKeysCmd(p *PaneModel, keys []string) tea.Cmd {
	return func() tea.Msg {
		_ = p.tm.SendKeysRaw(p.targetID, keys...)
		return nil
	}
}

// resizePanesCmd resizes tmux panes in a goroutine (non-blocking).
func resizePanesCmd(tm *tmux.Manager, panes [2]*PaneModel, dimensions [2][2]int) tea.Cmd {
	return func() tea.Msg {
		for i, p := range panes {
			w, h := dimensions[i][0], dimensions[i][1]
			if w > 0 && h > 0 {
				_ = tm.ResizePane(p.targetID, w, h)
			}
		}
		return nil
	}
}

func scheduleTick(d time.Duration) tea.Cmd {
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
		return m, m.syncTmuxPaneSizesCmd()

	case tickMsg:
		// Tick fired — launch async capture
		return m, captureAllCmd(m.panes)

	case capturedMsg:
		// Async capture completed — update pane content
		for i := range m.panes {
			p := m.panes[i]
			if msg.errors[i] != nil {
				p.errCount++
				if p.errCount >= 5 {
					p.content = fmt.Sprintf("[tmux capture failed: %v]", msg.errors[i])
				}
			} else {
				p.errCount = 0
				if msg.contents[i] != p.content {
					p.lastActive = time.Now()
					p.content = msg.contents[i]
				}
				if i == 0 {
					p.cursorX = msg.cursorX
					p.cursorY = msg.cursorY
				}
			}
		}
		// Schedule next tick using adaptive interval
		i0 := m.panes[0].TickInterval()
		i1 := m.panes[1].TickInterval()
		interval := i0
		if i1 < i0 {
			interval = i1
		}
		return m, scheduleTick(interval)

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
				m.panes[m.activePane].MarkActive()
				return m, sendKeysCmd(m.panes[m.activePane], tmuxKeys)
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
	isActive := idx == m.activePane
	if isActive {
		style = m.styles.FocusedBorder
	}

	paneWidth, paneHeight := m.paneDimensions(idx, availHeight)

	content := m.fitContent(pane.Content(), paneWidth, paneHeight, idx == 0, pane.cursorX, pane.cursorY)
	indicator := " "
	if isActive {
		indicator = "▸"
	}
	title := fmt.Sprintf(" %s%s ", indicator, pane.Label())
	return style.Width(paneWidth).Height(paneHeight).Render(title + "\n" + content)
}

func (m Model) fitContent(content string, width, height int, showCursor bool, cursorX, cursorY int) string {
	lines := strings.Split(content, "\n")
	visible := height - 1
	if visible < 0 {
		visible = 0
	}
	if len(lines) > visible {
		if showCursor {
			// Scroll to keep cursor visible (user terminal pane)
			start := cursorY - visible + 1
			if start < 0 {
				start = 0
			}
			if start+visible > len(lines) {
				start = len(lines) - visible
			}
			lines = lines[start : start+visible]
			// Adjust cursorY relative to the visible window
			cursorY = cursorY - start
		} else {
			// No cursor info — show last N lines (Claude pane)
			lines = lines[len(lines)-visible:]
		}
	}
	for i, line := range lines {
		if ansi.StringWidth(line) > width {
			lines[i] = ansi.Truncate(line, width, "")
		}
	}
	if showCursor && cursorY >= 0 && cursorY < len(lines) {
		lines[cursorY] = injectCursor(lines[cursorY], cursorX)
	}
	return strings.Join(lines, "\n")
}

// injectCursor highlights the character at column x with reverse video,
// preserving the original character (like a real terminal block cursor).
func injectCursor(line string, x int) string {
	lineWidth := ansi.StringWidth(line)
	if x >= lineWidth {
		return line + strings.Repeat(" ", x-lineWidth) + "\033[7m \033[0m"
	}
	prefix := ansi.Truncate(line, x, "")
	ch := ansi.Strip(ansi.Cut(line, x, x+1))
	if ch == "" {
		ch = " "
	}
	suffix := ansi.Cut(line, x+1, lineWidth)
	return prefix + "\033[7m" + ch + "\033[0m" + suffix
}

func (m Model) renderStatusBar() string {
	return m.styles.StatusBar.Width(m.width).Render(m.statusBar.RenderPlain())
}

func (m Model) syncTmuxPaneSizesCmd() tea.Cmd {
	statusHeight := 1
	// 4 accounts for top+bottom borders of 2 pane boxes
	availHeight := m.height - statusHeight - 4

	var dims [2][2]int
	for i := range m.panes {
		pw, ph := m.paneDimensions(i, availHeight)
		dims[i] = [2]int{pw, ph}
	}
	return resizePanesCmd(m.tm, m.panes, dims)
}
