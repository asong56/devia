package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"devia/internal/core"
)

// Exit codes. Mirrors core's Code* constants exactly, kept as separate
// names here so this file reads standalone as the CLI contract.
const (
	ExitOK       = 0 // success
	ExitError    = 1 // internal/unexpected error
	ExitUsage    = 2 // bad command-line usage
	ExitInput    = 3 // invalid input data
	ExitNotFound = 4 // file/resource not found
)

// jsonMode is set once at startup from the global --json flag. When
// true, every command emits a single JSON object to stdout instead of
// plain text, so scripts can pipe into jq / json.loads / JSON.parse
// without guessing at a text format.
var jsonMode bool

// quietMode is set once at startup from the global -q/--quiet flag.
// It suppresses purely informational stderr chatter (e.g. the "serve"
// startup banner) — it never suppresses errors. Errors are the one
// thing --quiet must not hide: a quiet tool that swallows failures
// silently is worse than a noisy one.
var quietMode bool

type jsonEnvelope struct {
	OK     bool        `json:"ok"`
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
	Code   int         `json:"code,omitempty"`
}

// printResult writes the successful result and exits 0. In --json mode
// it's wrapped as {"ok":true,"result":...}; otherwise strings print
// raw (no quotes, no trailing structure — the point is that `devia
// hash ...` output is directly usable, not something you have to strip
// quotes from), and non-strings are pretty-printed as JSON since plain
// text has no other reasonable representation for e.g. a match list.
func printResult(result interface{}) {
	if jsonMode {
		json.NewEncoder(os.Stdout).Encode(jsonEnvelope{OK: true, Result: result})
		os.Exit(ExitOK)
	}
	switch v := result.(type) {
	case string:
		fmt.Println(v)
	default:
		b, _ := json.MarshalIndent(v, "", "  ")
		fmt.Println(string(b))
	}
	os.Exit(ExitOK)
}

// fail prints err and exits with its standard code. stdout carries the
// JSON envelope in --json mode (so stdout is always valid JSON when
// that flag is set, success or failure); stderr always carries a
// human-readable explanation regardless of mode or --quiet, since
// something crawling stderr for logs shouldn't need --json to see what
// broke, and a failure is never something --quiet should hide.
//
// The stderr block follows a fixed shape — one line naming what
// failed, then a blank line and a "Try:" line naming the next step —
// so the answer to "what do I do about this" is never more than two
// lines away from the error itself.
func fail(err error) {
	code := core.CodeOf(err)
	if jsonMode {
		json.NewEncoder(os.Stdout).Encode(jsonEnvelope{OK: false, Error: err.Error(), Code: code})
	}
	printErrorBlock("error", err.Error(), recoveryHint(code))
	os.Exit(code)
}

// usageError reports a CLI usage mistake (missing arg, unknown
// subcommand) distinctly from a data/runtime error, exiting 2.
func usageError(msg string) {
	if jsonMode {
		json.NewEncoder(os.Stdout).Encode(jsonEnvelope{OK: false, Error: msg, Code: ExitUsage})
	}
	printErrorBlock("usage error", msg, "devia help")
	os.Exit(ExitUsage)
}

// printErrorBlock writes the fixed-shape stderr error report. The
// first line's format ("devia: <kind>: <msg>") is kept stable on its
// own so any existing `grep '^devia: error:'` still matches — only the
// "Try:" line is new.
func printErrorBlock(kind, msg, hint string) {
	fmt.Fprintf(os.Stderr, "devia: %s: %s\n", kind, msg)
	if hint != "" {
		fmt.Fprintf(os.Stderr, "\nTry:\n  %s\n", hint)
	}
}

// recoveryHint gives a generic next step keyed on the error's code —
// not on its exact message, so this stays one small switch instead of
// threading a bespoke hint through every one of core's ~30 error
// sites. CodeNotFound and CodeInput get a more specific nudge; other
// codes fall back to pointing at help.
func recoveryHint(code int) string {
	switch code {
	case core.CodeNotFound:
		return "check the path and try again"
	case core.CodeInput:
		return "devia help   (to check the expected input format)"
	default:
		return "devia help"
	}
}

// readInput returns arg if non-empty, otherwise reads stdin if it's
// piped (not a terminal). This is what makes `cat f.json | devia json
// format` and `devia json format '{"a":1}'` both work with the same
// command. Trailing newline from shell `echo` is trimmed.
func readInput(arg string) (string, error) {
	if arg != "" {
		return arg, nil
	}
	stat, err := os.Stdin.Stat()
	if err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
		b, rerr := io.ReadAll(os.Stdin)
		if rerr != nil {
			return "", rerr
		}
		return strings.TrimRight(string(b), "\n"), nil
	}
	return "", errors.New("no input provided: pass it as an argument or pipe it via stdin")
}

func readFileErr(err error, path string) error {
	if os.IsNotExist(err) {
		return core.NewNotFoundError("file not found: " + path)
	}
	return err
}
