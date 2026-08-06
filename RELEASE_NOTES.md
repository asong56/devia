# devia v1.0.0

**devia** is a zero-dependency, single-binary developer toolbox for the command line. Every common task that used to mean reaching for a browser, a Python one-liner, or three different npm packages — hashing, encoding, UUID generation, JSON formatting, cert inspection, cron scheduling, diff, and more — is a subcommand away, scriptable from any language, and pipes cleanly into the rest of your shell.

---

## Download

Each archive contains two binaries built from the same source:

| File | Contents |
|---|---|
| `devia-darwin-arm64.zip` | `devia` + `devia-cli` — macOS, Apple Silicon |
| `devia-linux-amd64.zip` | `devia` + `devia-cli` — Linux, x86\_64 |
| `devia-windows-x64.zip` | `devia.exe` + `devia-cli.exe` — Windows, x86\_64 |

**`devia`** — the full build. Includes everything, including `devia serve` (a local JSON API that exposes every command over HTTP).

**`devia-cli`** — the minimal build (`-tags noserve`). `net/http` and `crypto/tls` are never linked in, so the binary is measurably smaller. Identical in every other way; `devia serve` reports that the API is not included rather than silently failing.

Both binaries are statically linked (`CGO_ENABLED=0`), stripped of debug symbols and file paths (`-trimpath -ldflags="-s -w"`), and have no runtime dependencies — drop either one in your `PATH` and it works.

---

## Commands

```
devia [--json] <command> [subcommand] [flags] [args]
```

| Command | What it does |
|---|---|
| `hash` | MD5 / SHA-1 / SHA-256 / SHA-512, plain or HMAC, hex or base64, text or file (streamed) |
| `checksum` | Hash a file and optionally compare against an expected value; exits non-zero on mismatch |
| `base64` | encode / decode, URL-safe alphabet, data URI wrapping, binary file output |
| `json` | format / minify / validate |
| `escape` / `unescape` | JSON, URL, URL-path, HTML, Unicode |
| `uuid` | Cryptographically random UUIDv4, batch generation, uppercase option |
| `text` | 13 case transforms: lower, upper, sentence, title, camel, pascal, snake, constant, kebab, cobol, train, alternating, inverse |
| `lorem` | Lorem ipsum by word / sentence / paragraph, variable count |
| `timestamp` | Unix timestamp ↔ human date, timezone-aware, custom Go time layout |
| `radix` | Integer base conversion; auto-detects `0x` / `0o` / `0b` prefixes; shows bin/oct/dec/hex simultaneously |
| `cron` | Parse a 5- or 6-field cron expression, describe it in plain English, list the next N run times |
| `regex` | Test (find all matches with positions) and replace (with `$1`-style group references); RE2 semantics |
| `diff` | Line-level diff of two files or two text arguments, unified output |
| `cert` | Decode a PEM or DER certificate — subject, issuer, SANs, validity window, key type |
| `serve` | Start a local JSON API on `127.0.0.1:7654`; every command is a POST endpoint |

---

## Scripting contract

**Input** — the last positional argument is the input. If omitted, stdin is read when it is piped (not a terminal):

```sh
echo -n hello | devia hash --algo=sha256
cat cert.pem  | devia cert decode
```

**Output** — plain mode writes the result to stdout, nothing else. Errors go to stderr. This keeps stdout clean for piping:

```sh
SUM=$(devia hash --algo=sha256 --file archive.tar.gz)
```

**`--json` mode** — pass `--json` anywhere before the subcommand to get a single JSON object on stdout regardless of success or failure. Useful when calling devia from another language:

```sh
devia --json hash --algo=sha256 "hello" | jq -r .result
```

```python
import subprocess, json

def devia(*args):
    p = subprocess.run(["devia", "--json", *args], capture_output=True, text=True)
    out = json.loads(p.stdout)
    if not out["ok"]:
        raise RuntimeError(f"devia {out['code']}: {out['error']}")
    return out["result"]
```

**Exit codes**

| Code | Meaning |
|---|---|
| 0 | Success |
| 1 | Internal / unexpected error |
| 2 | Usage error (missing argument, unknown subcommand) |
| 3 | Invalid input data (bad JSON, bad regex, bad base64, …) |
| 4 | File or resource not found |

---

## JSON API (`devia serve`)

```sh
devia serve                     # binds to http://127.0.0.1:7654
devia serve --port 8080         # custom port
```

Every command is available as a `POST` endpoint that accepts and returns JSON. The response envelope matches `--json` mode exactly — `{"ok":true,"result":...}` or `{"ok":false,"error":"...","code":N}` — so you can share validation and error-handling code between the CLI and the API.

```sh
curl -s localhost:7654/api/v1/hash \
  -d '{"text":"hello world","algo":"sha256"}' | jq
```

Open `http://localhost:7654/` in a browser to see the full endpoint table.

Only available in `devia` (the full build). `devia-cli` reports clearly that the API is not included.

---

## Building from source

Requires Go 1.26+. No third-party dependencies — `go.mod` has no `require` entries.

```sh
git clone <repo>
cd devia

make build        # full binary:    ./devia
make build-min    # cli-only build: ./devia-cli
make build-all    # cross-compile all 6 platforms × 2 variants into ./build/
```

---

## Design notes

- **Zero dependencies.** Standard library only: `flag` for argument parsing, `net/http` (Go 1.22+ method routing) for the API, `crypto/rand` for UUID generation, `regexp` (RE2) for regex. The cron parser, diff engine, and MIME sniffer are all hand-written — no external packages.
- **RE2 regex.** Go's `regexp` package uses RE2 semantics: linear-time matching, no backtracking, no ReDoS. Backreferences and lookaheads are not supported. This is intentional for a tool that may be used against untrusted input.
- **Diff is line-level, O(n·m).** Capped at a few thousand lines; returns an error beyond that rather than consuming unbounded memory.
- **Cron uses stepping search.** The next-run finder steps forward by minute with a 4-year safety ceiling. Expressions that never match (e.g. February 30th) fail with exit code 3 rather than looping indefinitely.
- **`net/http` is isolated behind a build tag.** `internal/core` (all business logic) never imports `net/http`. The MIME type detection for base64 data URIs uses a hand-written magic-byte sniffer rather than `http.DetectContentType`, which would otherwise pull `crypto/tls` into a CLI-only binary. This is why `devia-cli` is meaningfully smaller, not just marginally.
