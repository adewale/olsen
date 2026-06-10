package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
)

// Doc-sync tests: the binary is the source of truth, and the documentation
// must agree with it. The audit found README examples whose flags the CLI
// silently ignored; these tests turn that class of drift into test failures.
//
// Direction 1: every command the binary advertises is documented in README.md.
// Direction 2: every `olsen ...` invocation in README.md / CLAUDE.md names a
// real command and uses only flags that command actually defines.

var (
	helpCache   = map[string]string{}
	helpCacheMu sync.Mutex
)

// commandHelp runs `olsen <args...>` once and caches its combined output.
func commandHelp(t *testing.T, args ...string) string {
	t.Helper()
	key := strings.Join(args, " ")
	helpCacheMu.Lock()
	defer helpCacheMu.Unlock()
	if out, ok := helpCache[key]; ok {
		return out
	}
	out, _, _ := runCommand(t, args...)
	helpCache[key] = out
	return out
}

// binaryCommands extracts the command names from `olsen help`.
func binaryCommands(t *testing.T) []string {
	t.Helper()
	help := commandHelp(t, "help")

	var commands []string
	inCommands := false
	re := regexp.MustCompile(`^\s{2}([a-z][a-z0-9-]*)\s{2,}\S`)
	for _, line := range strings.Split(help, "\n") {
		switch {
		case strings.HasPrefix(line, "Commands:"):
			inCommands = true
		case inCommands && strings.TrimSpace(line) == "":
			inCommands = false
		case inCommands:
			if m := re.FindStringSubmatch(line); m != nil {
				commands = append(commands, m[1])
			}
		}
	}
	if len(commands) == 0 {
		t.Fatalf("could not extract any commands from help output:\n%s", help)
	}
	return commands
}

func readRepoDoc(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", name))
	if err != nil {
		t.Fatalf("failed to read %s: %v", name, err)
	}
	return string(data)
}

// TestEveryCommandIsDocumented asserts each command the binary offers is
// mentioned in README.md, so new commands cannot ship undocumented.
func TestEveryCommandIsDocumented(t *testing.T) {
	readme := readRepoDoc(t, "README.md")

	for _, cmd := range binaryCommands(t) {
		if cmd == "help" || cmd == "version" {
			continue // meta-commands need no user documentation
		}
		if !strings.Contains(readme, "olsen "+cmd) {
			t.Errorf("command %q is advertised by `olsen help` but never shown in README.md", cmd)
		}
	}
}

// docInvocation is one `olsen <command> ...` example extracted from docs.
type docInvocation struct {
	source  string // file and line for error messages
	command string
	flags   []string // flag names without leading dashes
}

var (
	invocationRe = regexp.MustCompile(`(?:^|[\s"` + "`" + `])(?:\./bin/|\./)?olsen\s+([a-z][a-z0-9-]*)((?:\s+(?:--?[a-zA-Z][\w-]*|<[^>]+>|[^\s|;&)<>"]+))*)`)
	// Flags must start a whitespace-delimited token; otherwise the "-photos"
	// inside a filename like "my-photos.db" would be read as a flag.
	flagRe = regexp.MustCompile(`(?:^|\s)--?([a-zA-Z][\w-]*)`)
)

// extractInvocations pulls every documented olsen invocation out of a file.
func extractInvocations(t *testing.T, name string) []docInvocation {
	t.Helper()
	content := readRepoDoc(t, name)

	var invocations []docInvocation
	for i, line := range strings.Split(content, "\n") {
		for _, m := range invocationRe.FindAllStringSubmatch(line, -1) {
			inv := docInvocation{
				source:  name + ":" + itoa(i+1),
				command: m[1],
			}
			for _, fm := range flagRe.FindAllStringSubmatch(m[2], -1) {
				inv.flags = append(inv.flags, fm[1])
			}
			invocations = append(invocations, inv)
		}
	}
	return invocations
}

func itoa(i int) string {
	return string(rune('0'+i/1000%10)) + string(rune('0'+i/100%10)) + string(rune('0'+i/10%10)) + string(rune('0'+i%10))
}

// TestDocumentedInvocationsAreReal asserts every `olsen <cmd> --flag` example
// in README.md and CLAUDE.md names a real command and only flags that the
// command's own --help output defines.
//
// Regression context: README documented `olsen index <dir> --db photos.db`
// for months while the parser silently ignored everything after the
// positional argument.
func TestDocumentedInvocationsAreReal(t *testing.T) {
	known := map[string]bool{}
	for _, cmd := range binaryCommands(t) {
		known[cmd] = true
	}
	// Commands that exist behind build tags and are intentionally absent
	// from the default binary's help.
	taggedOnly := map[string]bool{"benchmark-libraw": true, "benchmark-thumbnails": true}

	for _, doc := range []string{"README.md", "CLAUDE.md"} {
		for _, inv := range extractInvocations(t, doc) {
			if taggedOnly[inv.command] {
				continue
			}
			if !known[inv.command] {
				t.Errorf("%s documents `olsen %s`, but the binary has no such command", inv.source, inv.command)
				continue
			}
			if len(inv.flags) == 0 {
				continue
			}
			help := commandHelp(t, inv.command, "--help")
			for _, flag := range inv.flags {
				if flag == "help" || flag == "h" || flag == "version" || flag == "v" {
					continue
				}
				// flag.PrintDefaults lists flags as "  -name".
				if !strings.Contains(help, "\n  -"+flag+" ") && !strings.Contains(help, "\n  -"+flag+"\n") {
					t.Errorf("%s documents `olsen %s --%s`, but `olsen %s --help` defines no such flag.\nHelp output:\n%s",
						inv.source, inv.command, flag, inv.command, help)
				}
			}
		}
	}
}
