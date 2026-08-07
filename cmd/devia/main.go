// Command devia is the entrypoint binary. It contains no real logic of
// its own: everything lives in devia/internal/cli (command dispatch
// and every command's CLI adapter) and devia/internal/server (the
// optional JSON API). main() and main2() exist only to wire os.Args
// into cli.Run() and to give the black-box tests in main_test.go a
// callable entry point that takes an explicit args slice.
package main

import (
	"os"

	"devia/internal/cli"

	// Blank-imported so time.LoadLocation (used by `devia timestamp`'s
	// --tz flag for names like Asia/Shanghai) works from the embedded
	// tzdata copy instead of depending on the deployment environment
	// having /usr/share/zoneinfo — devia is built with CGO_ENABLED=0
	// and may well end up in a scratch/distroless container that has
	// no system timezone database at all. Costs ~450KB in the binary;
	// a silently-broken --tz flag on a minimal image costs more.
	_ "time/tzdata"
)

func main() {
	main2(os.Args[1:])
}

// main2 exists only so cmd/devia/main_test.go's subprocess re-exec
// harness (TestHelperProcess) can call the exact same dispatch main()
// calls, but with an explicit args slice instead of relying on
// os.Args — the test binary's own os.Args[1:] includes go test's own
// flags, not devia's.
func main2(args []string) {
	cli.Run(args)
}
