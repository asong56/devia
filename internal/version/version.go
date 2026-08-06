// Package version holds devia's release version as a single constant.
// It exists as its own tiny package (rather than living in
// internal/cli) so that both internal/cli (the `devia version`
// command) and internal/server (the /api/v1/health endpoint) can
// import it without creating a cli<->server import cycle.
package version

const Version = "1.0.0"
