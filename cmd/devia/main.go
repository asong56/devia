// Command devia is the entrypoint binary. It contains no logic of its
// own: everything lives in devia/internal/cli (command dispatch and
// every command's CLI adapter) and devia/internal/server (the
// optional JSON API). Keeping package main to these three lines means
// `go build ./cmd/devia` produces a binary named devia and nothing
// else needs to know it's the entrypoint.
package main

import (
	"os"

	"devia/internal/cli"
)

func main() {
	cli.Run(os.Args[1:])
}
