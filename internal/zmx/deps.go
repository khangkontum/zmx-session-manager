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
	return scopedEnv(false)
}

func scopedEnv(global bool) []string {
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

func commandWithoutSessionPrefix(ctx context.Context, name string, arg ...string) *exec.Cmd {
	return commandForScope(ctx, false, name, arg...)
}

func commandForScope(ctx context.Context, global bool, name string, arg ...string) *exec.Cmd {
	cmd := deps.commandContext(ctx, name, arg...)
	cmd.Env = scopedEnv(global)
	return cmd
}

func combinedOutputWithoutSessionPrefix(name string, arg ...string) ([]byte, error) {
	return combinedOutputForScope(false, name, arg...)
}

func combinedOutputForScope(global bool, name string, arg ...string) ([]byte, error) {
	cmd := deps.command(name, arg...)
	cmd.Env = scopedEnv(global)
	return cmd.CombinedOutput()
}
