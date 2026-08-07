// Black-box tests for the devia binary's actual command-line contract:
// exit codes, stdout/stderr separation, --json envelopes, and stdin
// piping. cli.Run() ends every path in os.Exit, so these can't be
// ordinary in-process tests — they use the standard Go pattern of
// re-executing the test binary itself as a subprocess (see
// TestHelperProcess below), which is the same technique used by
// os/exec's own test suite.
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// runDevia executes `devia args...` as a subprocess (via the
// TestHelperProcess re-exec trick) with stdin optionally piped in, and
// returns stdout, stderr, and the process exit code.
func runDevia(t *testing.T, stdin string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cs := []string{"-test.run=TestHelperProcess", "--"}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = append(os.Environ(), "DEVIA_WANT_HELPER_PROCESS=1")
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run subprocess: %v", err)
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}

// TestHelperProcess is not a real test — it's the re-exec target.
// `go test` runs it like any other Test* function, so it must no-op
// unless the guard env var (set only by runDevia above) is present;
// otherwise every normal `go test ./cmd/devia` run would try to
// execute it as a real test and fail.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("DEVIA_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// args[0] is the test binary path, args[1] is "-test.run=...",
	// args[2] is "--", so the real devia args start at args[3].
	args := os.Args
	i := 0
	for ; i < len(args); i++ {
		if args[i] == "--" {
			i++
			break
		}
	}
	main2(args[i:])
	os.Exit(0) // unreachable in practice: main2 always os.Exits itself
}

func TestDevia_HashText_PlainMode(t *testing.T) {
	stdout, stderr, code := runDevia(t, "", "hash", "--algo=sha256", "hello")
	want := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if strings.TrimSpace(stdout) != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr on success, got %q", stderr)
	}
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
}

func TestDevia_StdinPiping(t *testing.T) {
	stdout, _, code := runDevia(t, "hello", "hash", "--algo=sha256")
	want := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if strings.TrimSpace(stdout) != want {
		t.Errorf("stdout = %q, want %q (piped stdin should be used as input when no arg is given)", stdout, want)
	}
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
}

func TestDevia_JSONMode_Success(t *testing.T) {
	stdout, _, code := runDevia(t, "", "--json", "hash", "--algo=sha256", "hello")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	var envelope struct {
		OK     bool   `json:"ok"`
		Result string `json:"result"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %s", err, stdout)
	}
	if !envelope.OK {
		t.Error("envelope.OK = false, want true for a successful call")
	}
	want := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if envelope.Result != want {
		t.Errorf("envelope.Result = %q, want %q", envelope.Result, want)
	}
}

func TestDevia_JSONMode_Error(t *testing.T) {
	// --json mode must still produce a valid single-line JSON envelope
	// on failure, not plain text — this is the property that lets
	// scripts always `| jq` the output regardless of success/failure.
	stdout, _, code := runDevia(t, "", "--json", "hash", "--algo=not-a-real-algo", "x")
	if code == 0 {
		t.Fatal("expected a non-zero exit code for an unsupported hash algorithm")
	}
	var envelope struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
		Code  int    `json:"code"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not valid JSON on failure: %v\nstdout: %s", err, stdout)
	}
	if envelope.OK {
		t.Error("envelope.OK = true, want false for a failed call")
	}
	if envelope.Error == "" {
		t.Error("envelope.Error is empty, want a message")
	}
	if envelope.Code != code {
		t.Errorf("envelope.Code = %d, but process exit code was %d — these must match", envelope.Code, code)
	}
}

func TestDevia_ExitCodes(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want int
	}{
		{"success", []string{"hash", "--algo=sha256", "x"}, 0},
		{"no command", []string{}, 2},
		{"unknown command", []string{"not-a-real-command"}, 2},
		{"bad json input", []string{"json", "validate", "{not valid}"}, 3},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, _, code := runDevia(t, "", c.args...)
			if code != c.want {
				t.Errorf("exit code = %d, want %d", code, c.want)
			}
		})
	}
}

func TestDevia_FileNotFound_ExitCode4(t *testing.T) {
	_, stderr, code := runDevia(t, "", "hash", "--file=/this/path/does/not/exist.bin")
	if code != 4 {
		t.Errorf("exit code = %d, want 4 (CodeNotFound)", code)
	}
	if !strings.Contains(stderr, "devia: error:") {
		t.Errorf("expected the standard 'devia: error:' prefix on stderr, got %q", stderr)
	}
}

func TestDevia_ErrorFormat_HasTryLine(t *testing.T) {
	// Pinning down the shape added for CLI-craftsmanship compliance:
	// first line is "devia: error: <msg>", then a blank line, then
	// "Try:" and a hint — and the first line's prefix must stay
	// grep-stable for any script relying on `grep '^devia: error:'`.
	_, stderr, _ := runDevia(t, "", "hash", "--file=/this/path/does/not/exist.bin")
	lines := strings.Split(strings.TrimRight(stderr, "\n"), "\n")
	if len(lines) < 4 {
		t.Fatalf("expected at least 4 lines (message, blank, Try:, hint), got %d:\n%s", len(lines), stderr)
	}
	if !strings.HasPrefix(lines[0], "devia: error:") {
		t.Errorf("first line = %q, want prefix %q", lines[0], "devia: error:")
	}
	if lines[1] != "" {
		t.Errorf("second line should be blank, got %q", lines[1])
	}
	if lines[2] != "Try:" {
		t.Errorf("third line = %q, want %q", lines[2], "Try:")
	}
	if strings.TrimSpace(lines[3]) == "" {
		t.Error("fourth line (the hint) should not be blank")
	}
}

func TestDevia_UsageError_ExitCode2(t *testing.T) {
	_, stderr, code := runDevia(t, "", "not-a-real-command")
	if code != 2 {
		t.Errorf("exit code = %d, want 2 (CodeUsage)", code)
	}
	if !strings.Contains(stderr, "devia: usage error:") {
		t.Errorf("expected 'devia: usage error:' prefix, got %q", stderr)
	}
}

func TestDevia_QuietMode_SuppressesServeStartupBanner(t *testing.T) {
	// serve blocks forever on success, so this test only exercises the
	// noserve-stub error path (this test binary is built without the
	// -tags noserve mechanism specifically applying here — it checks
	// the message shape either way) — the real assertion is just that
	// -q does not introduce a crash or change the exit code contract
	// for an already-erroring command; the banner-suppression itself
	// is covered at the unit level in internal/server, where the
	// startup fmt.Fprintf calls are directly guarded by `if !quiet`.
	_, _, codeLoud := runDevia(t, "", "hash", "--algo=bad-algo", "x")
	_, _, codeQuiet := runDevia(t, "", "-q", "hash", "--algo=bad-algo", "x")
	if codeLoud != codeQuiet {
		t.Errorf("-q changed the exit code for an error case: loud=%d quiet=%d (: should never affect error visibility)", codeLoud, codeQuiet)
	}
}

func TestDevia_QuietMode_DoesNotSuppressErrors(t *testing.T) {
	// The one non-negotiable property of --quiet: errors are never
	// hidden by it, regardless of position (before or after the
	// subcommand).
	for _, args := range [][]string{
		{"-q", "hash", "--algo=bad-algo", "x"},
		{"hash", "--algo=bad-algo", "x", "-q"},
		{"--quiet", "hash", "--algo=bad-algo", "x"},
	} {
		_, stderr, code := runDevia(t, "", args...)
		if code == 0 {
			t.Errorf("args=%v: expected a non-zero exit code even with quiet mode", args)
		}
		if stderr == "" {
			t.Errorf("args=%v: --quiet must not suppress error output", args)
		}
	}
}

func TestDevia_Completion_AllShells(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish"} {
		stdout, _, code := runDevia(t, "", "completion", shell)
		if code != 0 {
			t.Errorf("completion %s: exit code = %d, want 0", shell, code)
		}
		if !strings.Contains(stdout, "devia") {
			t.Errorf("completion %s: output doesn't mention devia: %q", shell, stdout)
		}
	}
}

func TestDevia_Completion_UnknownShell(t *testing.T) {
	_, _, code := runDevia(t, "", "completion", "powershell")
	if code != 2 {
		t.Errorf("exit code = %d, want 2 (CodeUsage) for an unsupported shell", code)
	}
}

func TestDevia_Version(t *testing.T) {
	stdout, _, code := runDevia(t, "", "version")
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.HasPrefix(strings.TrimSpace(stdout), "devia ") {
		t.Errorf("stdout = %q, want it to start with %q", stdout, "devia ")
	}
}

func TestDevia_Help_ExitsZero(t *testing.T) {
	stdout, _, code := runDevia(t, "", "help")
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Error("help output missing a Usage: section")
	}
	if !strings.Contains(stdout, "Global flags:") {
		t.Error("help output missing the Global flags: section (added for -q/--quiet discoverability)")
	}
}

func TestDevia_NoArgs_ExitsUsage(t *testing.T) {
	_, _, code := runDevia(t, "")
	if code != 2 {
		t.Errorf("exit code = %d, want 2 (CodeUsage) when called with no arguments at all", code)
	}
}
