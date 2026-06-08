package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	tea "charm.land/bubbletea/v2"
	"github.com/khangkontum/zmx-session-manager/internal/tui"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Printf("zsm %s (%s, %s)\n", version, commit, date)
		return
	}

	zmxPath, err := exec.LookPath("zmx")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: zmx not found in PATH")
		os.Exit(1)
	}

	p := tea.NewProgram(tui.NewModel())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// If the user pressed Enter to attach, exec into zmx attach
	if m, ok := finalModel.(tui.Model); ok && m.AttachTarget() != "" {
		env := withoutSessionPrefixEnv(m.AttachGlobal())
		syscall.Exec(zmxPath, []string{"zmx", "attach", m.AttachTarget()}, env)
	}
}

func withoutSessionPrefixEnv(global bool) []string {
	env := os.Environ()
	filtered := make([]string, 0, len(env))
	for _, entry := range env {
		if strings.HasPrefix(entry, "ZMX_SESSION_PREFIX=") {
			continue
		}
		if global && strings.HasPrefix(entry, "ZMX_DIR=") {
			continue
		}
		filtered = append(filtered, entry)
	}
	return filtered
}
