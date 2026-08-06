//go:build noserve

// Built with `go build -tags noserve`. This stub replaces server.go so
// the resulting binary never imports net/http (and therefore never
// links crypto/tls and the rest of that dependency graph) — this is
// what makes devia-cli meaningfully smaller than devia. `devia serve`
// still exists as a command in this build; it just reports plainly
// that this particular binary doesn't include it.
package server

import "errors"

// Run reports that the API isn't built into this binary. quiet is
// accepted (not used) purely to keep this signature identical to
// server.go's — cli/serve.go calls one function name regardless of
// which build tag is active.
func Run(host string, port int, quiet bool) error {
	return errors.New("this binary was built with -tags noserve and does not include the API server; rebuild without that tag (`make build`) to get `devia serve`")
}
