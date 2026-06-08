package zmx

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
)

// FetchPreview returns the last `lines` lines of `zmx history <name> --vt`.
// SGR color sequences are preserved so the preview matches the source session.
func FetchPreview(name string, lines int) string {
	return FetchPreviewForScope(name, lines, false)
}

func FetchPreviewForScope(name string, lines int, global bool) string {
	if lines < 1 {
		lines = 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := commandForScope(ctx, global, "zmx", "history", name, "--vt")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Sprintf("(preview unavailable: %v)", err)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return fmt.Sprintf("(preview unavailable: %v)", err)
	}

	preview, readErr := tailLinesFromReader(stdout, lines)
	waitErr := cmd.Wait()
	if ctx.Err() == context.DeadlineExceeded {
		return "(preview unavailable: timed out)"
	}
	if readErr != nil {
		return fmt.Sprintf("(preview unavailable: %v)", readErr)
	}
	if waitErr != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return fmt.Sprintf("(preview unavailable: %v: %s)", waitErr, msg)
		}
		return fmt.Sprintf("(preview unavailable: %v)", waitErr)
	}
	if preview == "" {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return fmt.Sprintf("(preview unavailable: %s)", msg)
		}
	}
	return preview
}

func tailLinesFromReader(r io.Reader, lines int) (string, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)

	tail := make([]string, 0, lines)
	for scanner.Scan() {
		tail = append(tail, preserveSGR(scanner.Text()))
		if len(tail) > lines {
			tail = tail[1:]
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return strings.Join(tail, "\n"), nil
}

// ScrollPreview applies a horizontal offset and width to raw preview text,
// truncating and padding each line for display in the preview pane.
func ScrollPreview(raw string, offsetX, maxWidth int) string {
	lines := strings.Split(raw, "\n")
	for i, line := range lines {
		line = strings.ReplaceAll(line, "\t", "    ")
		if offsetX > 0 {
			line = ansi.Cut(line, offsetX, ansi.StringWidth(line))
		}
		line = ansi.Truncate(line, maxWidth, "")
		if w := ansi.StringWidth(line); w < maxWidth {
			line += strings.Repeat(" ", maxWidth-w)
		}
		lines[i] = line + ansi.ResetStyle
	}
	return strings.Join(lines, "\n")
}

func preserveSGR(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' {
			start := i
			i++
			if i >= len(s) {
				break
			}
			switch s[i] {
			case '[': // CSI sequence: ESC [ ... <final byte 0x40-0x7E>
				i++
				for i < len(s) && (s[i] < 0x40 || s[i] > 0x7E) {
					i++
				}
				if i < len(s) {
					final := s[i]
					i++
					if final == 'm' {
						b.WriteString(s[start:i])
					}
				}
			case ']': // OSC sequence: ESC ] ... (BEL or ST)
				i++
				for i < len(s) {
					if s[i] == '\x07' {
						i++
						break
					}
					if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '\\' {
						i += 2
						break
					}
					i++
				}
			case '(', ')': // Charset designation: ESC ( X or ESC ) X
				i++
				if i < len(s) {
					i++
				}
			default: // ESC + single character
				i++
			}
		} else if s[i] == '\r' {
			// Skip carriage return — we only want newlines
			i++
		} else if s[i] < 0x20 && s[i] != '\n' && s[i] != '\t' {
			// Skip other control characters
			i++
		} else {
			b.WriteByte(s[i])
			i++
		}
	}
	return b.String()
}
