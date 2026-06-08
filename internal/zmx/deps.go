package zmx

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/atotto/clipboard"
)

type runtimeDeps struct {
	command        func(name string, arg ...string) *exec.Cmd
	commandContext func(ctx context.Context, name string, arg ...string) *exec.Cmd
	clipboardWrite func(text string) error
}

var deps = runtimeDeps{
	command:        exec.Command,
	commandContext: exec.CommandContext,
	clipboardWrite: clipboard.WriteAll,
}

func runCombinedOutput(name string, arg ...string) ([]byte, error) {
	return deps.command(name, arg...).CombinedOutput()
}

func withoutSessionPrefixEnv() []string {
	env := os.Environ()
	filtered := make([]string, 0, len(env))
	for _, entry := range env {
		if strings.HasPrefix(entry, "ZMX_SESSION_PREFIX=") {
			continue
		}
		filtered = append(filtered, entry)
	}
	return filtered
}

func commandWithoutSessionPrefix(ctx context.Context, name string, arg ...string) *exec.Cmd {
	cmd := deps.commandContext(ctx, name, arg...)
	cmd.Env = withoutSessionPrefixEnv()
	return cmd
}

func combinedOutputWithoutSessionPrefix(name string, arg ...string) ([]byte, error) {
	cmd := deps.command(name, arg...)
	cmd.Env = withoutSessionPrefixEnv()
	return cmd.CombinedOutput()
}
