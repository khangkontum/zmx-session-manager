package tui

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/khangkontum/zmx-session-manager/internal/zmx"
)

func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.state == stateConfirmKill {
		return m.handleConfirmKey(msg)
	}

	if isQuit(msg) {
		return m, tea.Quit
	}

	if isRune(msg, "?") {
		m.showHelp = true
		return m, nil
	}

	if m.handlePreviewJump(msg) {
		return m, nil
	}

	if m.handlePreviewScroll(msg) {
		return m, nil
	}

	// Escape or Backspace clears active filter in normal mode
	if (msg.Code == tea.KeyEscape || msg.Code == tea.KeyBackspace) && m.filterText != "" {
		m.filterText = ""
		m.markVisibleChanged()
		m.cursor = 0
		m.listOffset = 0
		m.previewScrollY = 0
		return m, m.previewCmd()
	}

	visible := m.visibleSessions()

	// Ctrl+A toggles select all
	if msg.Code == 'a' && msg.Mod.Contains(tea.ModCtrl) {
		m.toggleSelectAll(visible)
		return m, nil
	}

	switch msg.Code {
	case tea.KeyUp:
		return m.moveCursor(-1)

	case tea.KeyDown:
		return m.moveCursor(1)

	case tea.KeyLeft:
		if m.previewScrollX > 0 {
			m.previewScrollX -= 4
			if m.previewScrollX < 0 {
				m.previewScrollX = 0
			}
		}

	case tea.KeyRight:
		m.scrollPreviewRight(4)

	case tea.KeySpace:
		if m.cursor < len(visible) {
			name := visible[m.cursor].Name
			if m.selected[name] {
				delete(m.selected, name)
			} else {
				m.selected[name] = true
			}
		}

	case tea.KeyEnter:
		if m.cursor < len(visible) {
			m.attachTarget = visible[m.cursor].Name
			m.attachGlobal = m.globalSessions
			return m, tea.Quit
		}

	default:
		if msg.Text != "" {
			switch msg.Text {
			case "j":
				return m.moveCursor(1)
			case "k":
				return m.moveCursor(-1)
			case "x":
				targets := m.killTargets()
				if len(targets) > 0 {
					m.state = stateConfirmKill
				}
			case "c":
				if m.cursor < len(visible) {
					name := visible[m.cursor].Name
					text := fmt.Sprintf("zmx attach %s", name)
					if m.globalSessions {
						text = fmt.Sprintf("env -u ZMX_DIR -u ZMX_SESSION_PREFIX %s", text)
					}
					if err := zmx.CopyToClipboard(text); err != nil {
						m.status = fmt.Sprintf("Copy failed: %v", err)
						m.addLog(confirmStyle.Render(fmt.Sprintf("  ✗ Copy failed: %v", err)))
					} else {
						m.status = "Copied!"
						m.addLog(statusStyle.Render(fmt.Sprintf("  Copied: %s", text)))
					}
					return m, clearStatusAfter(2 * time.Second)
				}
			case "r":
				return m, fetchSessionsCmd(m.globalSessions)
			case "g":
				return m.toggleGlobalSessions()
			case "/":
				m.state = stateFilter
			case "s":
				if m.sortAsc {
					m.sortAsc = false
				} else {
					m.sortAsc = true
					m.sortMode = (m.sortMode + 1) % sortModeCount
				}
				m.markVisibleChanged()
				m.cursor = 0
				m.listOffset = 0
				m.previewScrollY = 0
				return m, m.previewCmd()
			}
		}
	}

	return m, nil
}

func (m Model) handleHelpKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if msg.Code == tea.KeyEscape || msg.Code == tea.KeyEnter || isRune(msg, "?") || isRune(msg, "q") {
		m.showHelp = false
	}
	return m, nil
}

func (m Model) toggleGlobalSessions() (tea.Model, tea.Cmd) {
	m.globalSessions = !m.globalSessions
	m.sessions = nil
	m.preview = ""
	m.previewScrollX = 0
	m.previewScrollY = 0
	m.selected = make(map[string]bool)
	m.filterText = ""
	m.cursor = 0
	m.listOffset = 0
	m.err = nil
	m.markSessionsChanged()
	if m.globalSessions {
		m.status = "Global sessions"
	} else {
		m.status = "Sessions: " + m.scopeLabel()
	}
	return m, tea.Batch(fetchSessionsCmd(m.globalSessions), clearStatusAfter(2*time.Second))
}

func (m Model) moveCursor(delta int) (tea.Model, tea.Cmd) {
	visible := m.visibleSessions()
	next := m.cursor + delta
	if next < 0 || next >= len(visible) {
		return m, nil
	}
	m.cursor = next
	m.previewScrollX = 0
	m.previewScrollY = 0
	m.ensureVisible()
	return m, m.previewCmd()
}

func (m *Model) toggleSelectAll(visible []Session) {
	if len(visible) == 0 {
		return
	}
	allSelected := true
	for _, s := range visible {
		if !m.selected[s.Name] {
			allSelected = false
			break
		}
	}
	if allSelected {
		for _, s := range visible {
			delete(m.selected, s.Name)
		}
	} else {
		for _, s := range visible {
			m.selected[s.Name] = true
		}
	}
}

// pruneSelections removes selections for sessions not in the current visible set.
func (m *Model) pruneSelections() {
	visible := m.visibleSessions()
	allowed := make(map[string]bool, len(visible))
	for _, s := range visible {
		allowed[s.Name] = true
	}
	for name := range m.selected {
		if !allowed[name] {
			delete(m.selected, name)
		}
	}
}

func (m Model) handleFilterKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if isQuit(msg) {
		return m, tea.Quit
	}

	switch msg.Code {
	case tea.KeyEscape:
		m.filterText = ""
		m.markVisibleChanged()
		m.state = stateNormal
		m.cursor = 0
		m.listOffset = 0
		m.previewScrollY = 0
		return m, m.previewCmd()

	case tea.KeyEnter:
		m.state = stateNormal
		m.clampCursor()
		return m, m.previewCmd()

	case tea.KeyBackspace:
		if len(m.filterText) > 0 {
			m.filterText = m.filterText[:len(m.filterText)-1]
			m.markVisibleChanged()
			m.pruneSelections()
			m.cursor = 0
			m.listOffset = 0
			m.previewScrollY = 0
		} else {
			// Backspace on empty filter exits filter mode
			m.state = stateNormal
			return m, m.previewCmd()
		}

	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
			m.previewScrollY = 0
			m.ensureVisible()
			return m, m.previewCmd()
		}

	case tea.KeyDown:
		visible := m.visibleSessions()
		if m.cursor < len(visible)-1 {
			m.cursor++
			m.previewScrollY = 0
			m.ensureVisible()
			return m, m.previewCmd()
		}

	default:
		if msg.Text != "" {
			m.filterText += msg.Text
			m.markVisibleChanged()
			m.pruneSelections()
			m.cursor = 0
			m.listOffset = 0
			m.previewScrollY = 0
		}
	}

	return m, nil
}

func (m Model) handleConfirmKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if isQuit(msg) {
		return m, tea.Quit
	}
	if msg.Code == tea.KeyEscape || msg.Code == tea.KeyBackspace {
		m.state = stateNormal
		return m, nil
	}
	if isRune(msg, "y") {
		targets := m.killTargets()
		total := len(targets)
		m.state = stateKilling
		m.killDoneNames = nil

		m.addLog(titleStyle.Render(fmt.Sprintf("Killing %d session(s)...", total)))

		first := targets[0]
		m.killQueue = targets[1:]
		m.killNow = first
		m.addLog(helpStyle.Render("  ⋯ " + first))
		return m, killOneCmd(first, m.globalSessions)
	}
	if isRune(msg, "n") {
		m.state = stateNormal
	}
	return m, nil
}

func (m *Model) handleLogScroll(msg tea.KeyPressMsg) {
	if !isRune(msg, "[") && !isRune(msg, "]") {
		return
	}
	maxOffset := len(m.logLines) - logContentHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if isRune(msg, "[") && m.logOffset > 0 {
		m.logOffset--
	}
	if isRune(msg, "]") && m.logOffset < maxOffset {
		m.logOffset++
	}
}

func (m *Model) handlePreviewScroll(msg tea.KeyPressMsg) bool {
	if !isPreviewScrollKey(msg, "[") && !isPreviewScrollKey(msg, "]") {
		return false
	}

	if isPreviewScrollKey(msg, "[") && m.previewScrollY > 0 {
		m.previewScrollY--
	}
	if isPreviewScrollKey(msg, "]") {
		m.previewScrollY++
		m.clampPreviewScroll()
	}
	return true
}

func (m *Model) handlePreviewJump(msg tea.KeyPressMsg) bool {
	switch msg.Text {
	case "{":
		m.previewScrollY = 0
		return true
	case "}":
		m.scrollPreviewBottom()
		return true
	}
	return false
}

func (m *Model) handlePreviewWheel(msg tea.MouseWheelMsg) bool {
	mouse := msg.Mouse()
	if !m.isInPreviewPane(mouse.X, mouse.Y) {
		return false
	}

	switch mouse.Button {
	case tea.MouseWheelUp:
		m.previewScrollY -= previewWheelScrollLines
		m.clampPreviewScroll()
	case tea.MouseWheelDown:
		m.previewScrollY += previewWheelScrollLines
		m.clampPreviewScroll()
	case tea.MouseWheelLeft:
		if m.previewScrollX > 0 {
			m.previewScrollX -= 4
			if m.previewScrollX < 0 {
				m.previewScrollX = 0
			}
		}
	case tea.MouseWheelRight:
		m.scrollPreviewRight(4)
	default:
		return false
	}
	return true
}

func (m *Model) handleSessionClick(msg tea.MouseClickMsg) (bool, tea.Cmd) {
	mouse := msg.Mouse()
	if mouse.Button != tea.MouseLeft || !m.isInListPane(mouse.X, mouse.Y) {
		return false, nil
	}

	row := mouse.Y - 1
	if row < 0 {
		return false, nil
	}
	visible := m.visibleSessions()
	idx := m.listOffset + row
	if idx < 0 || idx >= len(visible) {
		return false, nil
	}
	if idx == m.cursor {
		return true, nil
	}
	m.cursor = idx
	m.previewScrollX = 0
	m.previewScrollY = 0
	m.ensureVisible()
	return true, m.previewCmd()
}

func isPreviewScrollKey(msg tea.KeyPressMsg, key string) bool {
	if isRune(msg, key) {
		return true
	}
	if len(key) == 1 {
		return msg.Code == rune(key[0])
	}
	return false
}

func (m *Model) isInListPane(x, y int) bool {
	return x >= 0 && x < m.listOuterWidth() && y > 0 && y < m.previewContentHeight()+1
}

func (m *Model) isInPreviewPane(x, y int) bool {
	return x >= m.listOuterWidth() && x < m.width && y >= 0 && y < m.previewContentHeight()+2
}

func (m *Model) scrollPreviewRight(delta int) {
	maxW := previewMaxWidth(m.preview)
	limit := maxW - m.previewInnerWidth()
	if limit < 0 {
		limit = 0
	}
	if m.previewScrollX+delta <= limit {
		m.previewScrollX += delta
	} else {
		m.previewScrollX = limit
	}
}

func (m *Model) ensureVisible() {
	h := m.mainContentHeight(1)
	if h <= 0 {
		return
	}
	if m.cursor < m.listOffset {
		m.listOffset = m.cursor
	}
	if m.cursor >= m.listOffset+h {
		m.listOffset = m.cursor - h + 1
	}
}
