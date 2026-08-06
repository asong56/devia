package cli

import (
	"flag"
	"io"
)

// newFlagSet creates a FlagSet in ContinueOnError mode with its
// default usage output silenced — we report parse errors ourselves via
// usageError() so they follow the same --json / exit-code contract as
// every other error in this program, instead of flag's own
// os.Exit(2)-and-print-to-stderr behavior.
func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet("devia "+name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

func parseArgs(fs *flag.FlagSet, args []string) {
	if err := fs.Parse(args); err != nil {
		usageError(err.Error())
	}
}
