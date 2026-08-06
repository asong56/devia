// Package cli implements devia's command-line interface: argument
// dispatch, the shared flag/output helpers, and every command's CLI
// adapter. Every command's actual logic lives in devia/internal/core;
// this package (and internal/server for the HTTP API) is a thin,
// script-friendly shell around it.
package cli

import (
	"fmt"
	"os"

	"devia/internal/version"
)

// commandNames is the single source of truth for devia's top-level
// subcommands — used by the "unknown command" hint and by `devia
// completion`, so the two can never drift apart from the switch below.
var commandNames = []string{
	"hash", "checksum", "base64", "json", "escape", "unescape", "uuid",
	"text", "lorem", "timestamp", "radix", "cron", "regex", "diff",
	"cert", "serve", "completion", "help", "version",
}

// Run parses args (typically os.Args[1:]) and dispatches to the
// matching command. It never returns normally — every path ends in
// os.Exit, which is what keeps the exit-code contract (see output.go)
// centralized in one place instead of scattered across commands.
func Run(args []string) {
	var found bool
	args, found = extractFlag(args, "--json")
	jsonMode = found
	args, found = extractFlag(args, "--quiet")
	quietMode = found
	if !quietMode {
		args, found = extractFlag(args, "-q")
		quietMode = found
	}

	if len(args) == 0 {
		printHelp()
		os.Exit(ExitUsage)
	}

	cmd := args[0]
	rest := args[1:]

	switch cmd {
	case "help", "-h", "--help":
		printHelp()
		os.Exit(ExitOK)
	case "version", "-v", "--version":
		fmt.Println("devia " + version.Version)
		os.Exit(ExitOK)
	case "hash":
		cmdHash(rest)
	case "checksum":
		cmdChecksum(rest)
	case "base64":
		cmdBase64(rest)
	case "json":
		cmdJSON(rest)
	case "escape":
		cmdEscape(rest, false)
	case "unescape":
		cmdEscape(rest, true)
	case "uuid":
		cmdUUID(rest)
	case "text":
		cmdText(rest)
	case "lorem":
		cmdLorem(rest)
	case "timestamp":
		cmdTimestamp(rest)
	case "radix":
		cmdRadix(rest)
	case "cron":
		cmdCron(rest)
	case "regex":
		cmdRegex(rest)
	case "diff":
		cmdDiff(rest)
	case "cert":
		cmdCert(rest)
	case "serve":
		cmdServe(rest)
	case "completion":
		cmdCompletion(rest)
	default:
		usageError("unknown command: " + cmd)
	}
}

// extractFlag removes the first occurrence of flag (a bare token, not
// a flag.FlagSet flag) from args, wherever it appears, and reports
// whether it was found. Used only for the single global --json flag so
// it can appear before or after the subcommand: both `devia --json
// hash x` and `devia hash --json x` work.
func extractFlag(args []string, flagName string) ([]string, bool) {
	out := make([]string, 0, len(args))
	found := false
	for _, a := range args {
		if a == flagName {
			found = true
			continue
		}
		out = append(out, a)
	}
	return out, found
}

func printHelp() {
	fmt.Print(`devia - a small, scriptable developer toolbox (zero dependencies)

Usage:
  devia [flags] <command> [subcommand] [flags] [args]

Global flags:
  -h, --help      show this help
  -v, --version   print the version
  -q, --quiet     suppress non-essential stderr chatter (errors still show)
      --json      emit a single JSON line instead of plain text

Commands:
  hash        --algo=md5|sha1|sha256|sha512 [--hmac=key] [--base64] [--file=path] [text]
  checksum    [--algo=..] [--compare=hash] <file>
  base64      encode|decode [--file=path] [--out=path] [--dry-run] [--data-uri] [--url] [text]
  json        format|minify|validate [--indent=".."] [--file=path] [text]
  escape      json|url|url-path|html|unicode [text]
  unescape    json|url|url-path|html|unicode [text]
  uuid        [--count=N] [--upper]
  text        <mode> [text]
              modes: lower upper sentence title camel pascal snake
                     constant kebab cobol train alternating inverse
  lorem       [--type=word|sentence|paragraph] [--count=N] [--classic]
  timestamp   now | to-date <unix> | from-date <date>   [--tz=..] [--format=..]
  radix       [--from=N] <number>                        (auto-detects 0x/0o/0b)
  cron        [--next=N] <expr>                           (5 or 6 field expression)
  regex       test --pattern=.. [--flags=ims] [text]
              replace --pattern=.. --with=.. [text]
  diff        --a=file --b=file | <textA> <textB>
  cert        decode <file>                               (or pipe PEM via stdin)
  serve       [--host=127.0.0.1] [--port=7654]            start the JSON API
  completion  bash|zsh|fish                               print a completion script

Input:
  The last positional argument is the input. If omitted, devia reads
  stdin when it's piped:
    echo -n hello | devia hash --algo=sha256

Output:
  Plain mode (default): result on stdout, nothing else. Errors go to
  stderr and set the exit code — stdout stays clean for piping.

  --json mode: a single JSON line on stdout, always:
    {"ok":true,"result":...}
    {"ok":false,"error":"...","code":N}

Exit codes:
  0 ok   1 internal error   2 usage error   3 invalid input   4 not found

devia version ` + version.Version + "\n")
}
