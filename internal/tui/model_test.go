package tui

import (
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		s      string
		maxLen int
		want   string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello world", 8, "hello..."},
		{"hello world", 3, "hel"},
		{"hello world", 2, "he"},
		{"hello world", 1, "h"},
		{"hello world", 0, ""},
		{"hi", 4, "hi"},
		{"abcdefgh", 7, "abcd..."},
	}
	for _, tt := range tests {
		got := truncate(tt.s, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.s, tt.maxLen, got, tt.want)
		}
	}
}

func TestHighlightMatch(t *testing.T) {
	// Use unstyled styles so we can verify the string content
	noStyle := lipgloss.NewStyle()

	tests := []struct {
		s     string
		query string
		want  string
	}{
		{"my-session", "ses", "my-session"},  // match present
		{"my-session", "xyz", "my-session"},  // no match
		{"My-Session", "my-s", "My-Session"}, // case insensitive
		{"frontend", "front", "frontend"},    // match at start
		{"backend", "end", "backend"},        // match at end
	}
	for _, tt := range tests {
		got := highlightMatch(tt.s, tt.query, noStyle, noStyle)
		// Strip any ANSI sequences for comparison since unstyled lipgloss
		// may still produce reset sequences
		plain := stripStyleCodes(got)
		if plain != tt.want {
			t.Errorf("highlightMatch(%q, %q) plain = %q, want %q", tt.s, tt.query, plain, tt.want)
		}
	}
}

func TestHighlightMatch_ContainsQuery(t *testing.T) {
	base := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	hl := lipgloss.NewStyle().Bold(true)

	result := highlightMatch("my-session", "ses", base, hl)
	// The highlighted portion should be present unsplit
	if !strings.Contains(result, "ses") {
		t.Errorf("expected highlighted result to contain 'ses', got %q", result)
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		s     string
		width int
		want  string
	}{
		{"hi", 5, "hi   "},
		{"hello", 5, "hello"},
		{"toolong", 3, "toolong"}, // doesn't truncate
		{"", 3, "   "},
	}
	for _, tt := range tests {
		got := padRight(tt.s, tt.width)
		if got != tt.want {
			t.Errorf("padRight(%q, %d) = %q, want %q", tt.s, tt.width, got, tt.want)
		}
	}
}

func TestPadWidthUnicode(t *testing.T) {
	left := padLeft("界", 4)
	if got := lipgloss.Width(left); got != 4 {
		t.Fatalf("padLeft unicode width=%d want 4 (%q)", got, left)
	}
	right := padRight("界", 4)
	if got := lipgloss.Width(right); got != 4 {
		t.Fatalf("padRight unicode width=%d want 4 (%q)", got, right)
	}
}

func TestTruncateUnicodeWidth(t *testing.T) {
	got := truncate("你好世界", 5)
	if w := lipgloss.Width(got); w > 5 {
		t.Fatalf("truncate width=%d exceeds max: %q", w, got)
	}
}

func TestPreviewMsgIgnoresStaleSession(t *testing.T) {
	m := initialModel()
	m.sessions = []Session{{Name: "alpha"}, {Name: "beta"}}
	m.cursor = 1

	updated, _ := m.Update(previewMsg{name: "alpha", content: "stale"})
	got := updated.(Model)
	if got.preview != "" {
		t.Fatalf("stale preview should be ignored, got %q", got.preview)
	}

	updated, _ = got.Update(previewMsg{name: "beta", content: "fresh"})
	got = updated.(Model)
	if got.preview != "fresh" {
		t.Fatalf("current preview should be applied, got %q", got.preview)
	}
}

func TestPreviewMsgAppliesForCurrentSessionDuringResize(t *testing.T) {
	m := initialModel()
	m.sessions = []Session{{Name: "alpha"}}
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	got := updated.(Model)

	updated, _ = got.Update(previewMsg{name: "alpha", content: "ok"})
	got = updated.(Model)
	if got.preview != "ok" {
		t.Fatalf("preview should be updated for current session, got %q", got.preview)
	}
}

func TestPreviewMsgDefaultsToBottom(t *testing.T) {
	m := initialModel()
	m.width = 120
	m.height = 12
	m.sessions = []Session{{Name: "alpha"}}
	m.markSessionsChanged()

	updated, _ := m.Update(previewMsg{name: "alpha", content: "one\ntwo\nthree\nfour\nfive"})
	got := updated.(Model)
	if got.previewScrollY != got.previewMaxScrollY() {
		t.Fatalf("previewScrollY = %d, want bottom %d", got.previewScrollY, got.previewMaxScrollY())
	}
}

func TestPreviewBracketScrollChangesPreviewOffset(t *testing.T) {
	m := initialModel()
	m.width = 120
	m.height = 12
	m.preview = "one\ntwo\nthree\nfour\nfive"

	updated, _ := m.handleKey(tea.KeyPressMsg(tea.Key{Text: "]", Code: ']'}))
	got := updated.(Model)
	if got.previewScrollY != 1 {
		t.Fatalf("previewScrollY after ] = %d, want 1", got.previewScrollY)
	}

	updated, _ = got.handleKey(tea.KeyPressMsg(tea.Key{Text: "[", Code: '['}))
	got = updated.(Model)
	if got.previewScrollY != 0 {
		t.Fatalf("previewScrollY after [ = %d, want 0", got.previewScrollY)
	}
}

func TestPreviewShiftBracketScrollsToTopAndBottom(t *testing.T) {
	m := initialModel()
	m.width = 120
	m.height = 12
	m.preview = "one\ntwo\nthree\nfour\nfive"

	updated, _ := m.handleKey(tea.KeyPressMsg(tea.Key{Text: "}", Code: ']'}))
	got := updated.(Model)
	if got.previewScrollY != got.previewMaxScrollY() {
		t.Fatalf("previewScrollY after shift+] = %d, want bottom %d", got.previewScrollY, got.previewMaxScrollY())
	}

	updated, _ = got.handleKey(tea.KeyPressMsg(tea.Key{Text: "{", Code: '['}))
	got = updated.(Model)
	if got.previewScrollY != 0 {
		t.Fatalf("previewScrollY after shift+[ = %d, want 0", got.previewScrollY)
	}
}

func TestRenderPreviewContentSlicesVertically(t *testing.T) {
	got := renderPreviewContent("one\ntwo\nthree", 1, 2, 0, 10)
	if plain := stripStyleCodes(got); plain != "two       \nthree     " {
		t.Fatalf("renderPreviewContent = %q, want two/three slice", plain)
	}
}

func TestPreviewPaneHeightIsCapped(t *testing.T) {
	content := renderPreviewContent(strings.Repeat("very-long-preview-line ", 20), 0, 3, 0, 20)
	pane := previewBorderStyle.
		Width(22).
		Height(5).
		MaxWidth(22).
		MaxHeight(5).
		Render(content)
	if h := lipgloss.Height(pane); h > 5 {
		t.Fatalf("preview pane height = %d, want <= 5", h)
	}
}

func TestPreviewMouseWheelScrollsOnlyInsidePreviewPane(t *testing.T) {
	m := initialModel()
	m.width = 120
	m.height = 12
	m.sessions = []Session{{Name: "alpha"}}
	m.markSessionsChanged()
	m.preview = "one\ntwo\nthree\nfour\nfive\nsix"

	inside := tea.MouseWheelMsg(tea.Mouse{X: m.listOuterWidth() + 1, Y: 1, Button: tea.MouseWheelDown})
	updated, _ := m.Update(inside)
	got := updated.(Model)
	if got.previewScrollY != previewWheelScrollLines {
		t.Fatalf("previewScrollY after wheel = %d, want %d", got.previewScrollY, previewWheelScrollLines)
	}
	scrolled := got.previewScrollY

	outside := tea.MouseWheelMsg(tea.Mouse{X: 0, Y: 1, Button: tea.MouseWheelDown})
	updated, _ = got.Update(outside)
	got = updated.(Model)
	if got.previewScrollY != scrolled {
		t.Fatalf("previewScrollY changed for wheel outside preview pane: %d", got.previewScrollY)
	}
}

func TestViewEnablesMouseMode(t *testing.T) {
	m := initialModel()
	m.width = 80
	m.height = 20
	view := m.View()
	if view.MouseMode != tea.MouseModeCellMotion {
		t.Fatalf("MouseMode = %v, want MouseModeCellMotion", view.MouseMode)
	}
}

func TestVisibleSessionsInvalidatesAfterFilterAndSortChange(t *testing.T) {
	m := initialModel()
	m.sessions = []Session{
		{Name: "beta", PID: "2"},
		{Name: "alpha", PID: "1"},
	}
	m.markSessionsChanged()

	visible := m.visibleSessions()
	if len(visible) != 2 || visible[0].Name != "alpha" {
		t.Fatalf("unexpected initial ordering: %+v", visible)
	}

	m.filterText = "bet"
	m.markVisibleChanged()
	visible = m.visibleSessions()
	if len(visible) != 1 || visible[0].Name != "beta" {
		t.Fatalf("filter invalidation failed: %+v", visible)
	}

	m.filterText = ""
	m.sortAsc = false
	m.markVisibleChanged()
	visible = m.visibleSessions()
	if len(visible) != 2 || visible[0].Name != "beta" {
		t.Fatalf("sort invalidation failed: %+v", visible)
	}
}

// stripStyleCodes removes ANSI escape sequences for test comparison.
func stripStyleCodes(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
}
