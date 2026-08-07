package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devia/internal/core"
)

// readOwnSource reads a source file from this package's own directory
// (the test binary's working directory is the package directory under
// `go test`), for the one test below that cross-checks commandNames
// against Run()'s actual dispatch switch by scanning the source text.
func readOwnSource(t *testing.T, filename string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(".", filename))
	if err != nil {
		t.Fatalf("reading %s: %v", filename, err)
	}
	return string(b)
}

// readFileErrHelperStat returns a genuine os.ErrNotExist-satisfying
// error (by actually stat-ing a path that can't exist), rather than a
// hand-built stand-in — os.IsNotExist has historically been picky
// about exactly what shape of error it recognizes, so the test should
// exercise the real thing.
func readFileErrHelperStat(t *testing.T) (os.FileInfo, error) {
	t.Helper()
	return os.Stat(filepath.Join(t.TempDir(), "definitely-does-not-exist"))
}

// customStatErr is a plain error that does NOT satisfy os.IsNotExist,
// standing in for something like a permission error.
type customStatErr struct{}

func (*customStatErr) Error() string { return "permission denied (test stand-in)" }

func TestExtractFlag_Present(t *testing.T) {
	out, found := extractFlag([]string{"hash", "--json", "x"}, "--json")
	if !found {
		t.Error("expected found=true")
	}
	want := []string{"hash", "x"}
	if len(out) != len(want) || out[0] != want[0] || out[1] != want[1] {
		t.Errorf("out = %v, want %v", out, want)
	}
}

func TestExtractFlag_Absent(t *testing.T) {
	in := []string{"hash", "x"}
	out, found := extractFlag(in, "--json")
	if found {
		t.Error("expected found=false")
	}
	if len(out) != len(in) || out[0] != in[0] || out[1] != in[1] {
		t.Errorf("out = %v, want unchanged %v", out, in)
	}
}

func TestExtractFlag_AnyPosition(t *testing.T) {
	// The whole point of extractFlag over a simple index check is that
	// --json (and --quiet) must work whether they appear before or
	// after the subcommand.
	before, foundBefore := extractFlag([]string{"--json", "hash", "x"}, "--json")
	after, foundAfter := extractFlag([]string{"hash", "x", "--json"}, "--json")
	if !foundBefore || !foundAfter {
		t.Fatalf("expected found=true in both positions, got before=%v after=%v", foundBefore, foundAfter)
	}
	if len(before) != 2 || len(after) != 2 {
		t.Errorf("expected the flag removed regardless of position: before=%v after=%v", before, after)
	}
}

func TestExtractFlag_OnlyFirstOccurrenceReported_AllRemoved(t *testing.T) {
	// A duplicated flag (unusual, but not something the parser should
	// choke on) should still be found and every occurrence stripped —
	// not just the first — so it never leaks through to a subcommand's
	// own FlagSet as an unexpected positional argument.
	out, found := extractFlag([]string{"--json", "hash", "--json", "x"}, "--json")
	if !found {
		t.Fatal("expected found=true")
	}
	for _, a := range out {
		if a == "--json" {
			t.Errorf("--json should be fully stripped, still present in %v", out)
		}
	}
}

func TestExtractFlag_EmptyArgs(t *testing.T) {
	out, found := extractFlag(nil, "--json")
	if found {
		t.Error("expected found=false for empty args")
	}
	if len(out) != 0 {
		t.Errorf("expected empty output, got %v", out)
	}
}

func TestCommandNames_MatchesRunDispatch(t *testing.T) {
	// commandNames feeds `devia completion`'s generated scripts. If it
	// drifts from the actual switch in Run, shell completion would
	// silently start suggesting commands that don't exist (or omit
	// ones that do) — this test fails loudly instead the moment that
	// happens, by cross-checking against Run's source text.
	runSrc := readOwnSource(t, "run.go")
	for _, name := range commandNames {
		if name == "help" || name == "version" {
			// These are handled by dedicated case clauses with extra
			// aliases (-h/--help, -v/--version) rather than a plain
			// `case "help":` on its own line in the same shape as the
			// rest — check for their presence a different way.
			if !strings.Contains(runSrc, `"`+name+`"`) {
				t.Errorf("commandNames contains %q but run.go's dispatch has no matching case", name)
			}
			continue
		}
		if !strings.Contains(runSrc, `case "`+name+`":`) {
			t.Errorf("commandNames contains %q but Run() has no matching case clause in its switch — completion would suggest a command that doesn't exist", name)
		}
	}
}

func TestRecoveryHint_KnownCodes(t *testing.T) {
	cases := []struct {
		code int
		want string
	}{
		{core.CodeNotFound, "check the path and try again"},
		{core.CodeInput, "devia help   (to check the expected input format)"},
		{core.CodeError, "devia help"},
		{core.CodeUsage, "devia help"},
		{9999, "devia help"}, // unrecognized code must still fall back sanely, not panic
	}
	for _, c := range cases {
		got := recoveryHint(c.code)
		if got != c.want {
			t.Errorf("recoveryHint(%d) = %q, want %q", c.code, got, c.want)
		}
	}
}

func TestReadFileErr_NotFound(t *testing.T) {
	_, statErr := readFileErrHelperStat(t)
	got := readFileErr(statErr, "missing.txt")
	if core.CodeOf(got) != core.CodeNotFound {
		t.Errorf("CodeOf(readFileErr(ENOENT, ...)) = %d, want CodeNotFound (%d)", core.CodeOf(got), core.CodeNotFound)
	}
}

func TestReadFileErr_OtherErrorPassesThrough(t *testing.T) {
	// A non-"not exist" error (e.g. permission denied) must be
	// returned as-is, not miscategorized as CodeNotFound — those are
	// different problems with different recovery hints.
	other := &customStatErr{}
	got := readFileErr(other, "some/path")
	if got != other {
		t.Errorf("expected the original error to pass through unchanged, got %v", got)
	}
}

func TestReadInput_ArgTakesPrecedence(t *testing.T) {
	// When an explicit argument is given, readInput must return it
	// immediately without even looking at stdin — this is the only
	// branch of readInput safe to unit-test in-process, since the
	// stdin-piping branch depends on the test binary's own stdin and
	// is covered instead by the black-box CLI tests in cmd/devia.
	got, err := readInput("explicit text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "explicit text" {
		t.Errorf("got %q, want %q", got, "explicit text")
	}
}
