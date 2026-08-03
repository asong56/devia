# devia

A small command-line toolbox for everyday developer tasks: hashing,
base64, JSON formatting, escaping, UUIDs, text case conversion,
timestamps, cron schedules, regex testing, diffing, and X.509
certificate inspection. Single static binary, zero external
dependencies — everything is built on the Go standard library.

There's also an optional local JSON API (`devia serve`), for when you'd
rather send a request than shell out to a binary.

```
$ devia hash --algo=sha256 "hello world"
b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9

$ devia uuid
c3f1a2b4-9e6d-4a1f-8b2c-7d5e6f3a1b9c

$ echo '{"name":"devia","fast":true}' | devia json format
{
  "name": "devia",
  "fast": true
}
```

## Installing / building

You'll need [Go 1.22 or later](https://go.dev/dl/) — the API server
relies on the method-based routing that `net/http` gained in that
release.

```bash
git clone <this repo>
cd devia

make build       # devia:     full build (CLI + `devia serve`)
make build-min   # devia-cli: CLI only, no net/http linked, smallest binary
make build-all   # cross-compiles both variants for linux/macOS/windows, amd64+arm64
```

On Windows, use `build.bat` instead of `build.sh`.

> **Note:** `go.mod` currently declares the module as just `devia`,
> which is all `make build` and friends need. If you'd like
> `go install github.com/you/devia@latest` to work once you've pushed
> this to your own account, update the module line in `go.mod` and the
> `devia/core` import in every file to match your actual repo path.

Every build passes `-trimpath -ldflags="-s -w"` (strips the symbol
table, debug info, and embedded file paths). That's a free 25–35% size
reduction with no behavior change. If you want to go further, running
the result through [UPX](https://upx.github.io/) (`upx --best --lzma
devia`) can shrink it more — it's just not wired into the default
build, since UPX-packed binaries occasionally trip antivirus
heuristics, and that trade-off is yours to make, not the build's to
assume.

Go's runtime itself has a fixed baseline size (roughly 1–1.5MB
regardless of how little code you write — this is a property of the
language, not this project), so don't expect a Rust-sized binary. To
see actual numbers for your platform:

```bash
make size
```

## Command reference

```
devia [--json] <command> [subcommand] [flags] [args]

hash        --algo=md5|sha1|sha256|sha512 [--hmac=key] [--base64] [--file=path] [text]
checksum    [--algo=..] [--compare=hash] <file>
base64      encode|decode [--file=path] [--out=path] [--data-uri] [--url] [text]
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
```

### Examples

```bash
devia hash --algo=sha256 "hello"
devia hash --file big.iso --algo=sha256              # streamed, not loaded into memory
devia checksum installer.exe --compare abc123...     # exit 0 = match, 1 = mismatch

devia base64 encode --file logo.png --data-uri       # data:image/png;base64,...
devia base64 decode "aGVsbG8="

devia json format < response.json
echo '{"a":1}' | devia json minify

devia escape json 'hello "world"'
devia unescape unicode '\u4f60\u597d'

devia uuid --count 5
devia text camel "hello-world"        # helloWorld
devia text snake "helloWorld"         # hello_world

devia timestamp now
devia timestamp to-date 1735689600 --tz Asia/Shanghai

devia radix 0xFF                      # binary/octal/decimal/hex, all four at once
devia cron "*/15 * * * *" --next 3

devia regex test --pattern '\d+' "room 42, floor 3"
devia diff --a old.txt --b new.txt

devia cert decode server.crt
```

## Built for scripting

This is the part the tool actually cares most about — being pleasant
to call from a shell script or another program.

**Output is split cleanly between stdout and stderr.** Plain mode
(the default) writes only the result to stdout — no extra labels or
prompts — so `X=$(devia hash --algo=md5 "x")` just works. Errors and
diagnostics always go to stderr, in both plain and `--json` mode.

**`--json` mode** turns stdout into a single line of JSON, always in
one of these two shapes:

```json
{"ok":true,"result":...}
{"ok":false,"error":"...","code":N}
```

```bash
devia --json hash --algo=sha256 "test" | jq -r .result
```

**Exit codes are standardized** across every command, so a script can
branch on the *kind* of failure instead of pattern-matching stderr
text:

| Code | Meaning |
|---|---|
| 0 | success |
| 1 | internal / runtime error |
| 2 | usage error (missing argument, unknown subcommand) |
| 3 | invalid input data (bad JSON, bad base64, bad regex, ...) |
| 4 | file or resource not found |

```bash
devia json validate < maybe.json
case $? in
  0) echo "valid" ;;
  3) echo "bad json, skipping" ;;
  *) echo "unexpected failure" >&2; exit 1 ;;
esac
```

**Stdin is read automatically** whenever the last positional argument
is omitted and stdin is piped (not a terminal, so it never hangs
waiting for interactive input):

```bash
echo -n hello | devia hash --algo=sha256
cat cert.pem | devia cert decode --json
```

### Calling devia from other languages

```python
import subprocess, json

def devia(*args):
    p = subprocess.run(["devia", "--json", *args], capture_output=True, text=True)
    out = json.loads(p.stdout)
    if not out["ok"]:
        raise RuntimeError(f"devia error {out['code']}: {out['error']}")
    return out["result"]

digest = devia("hash", "--algo=sha256", "hello world")
```

```javascript
const { execFileSync } = require("child_process");
function devia(...args) {
  const out = JSON.parse(execFileSync("devia", ["--json", ...args]));
  if (!out.ok) throw new Error(`devia error ${out.code}: ${out.error}`);
  return out.result;
}
```

Both examples use `execFile`/`subprocess.run` with an argument array
rather than building a shell command string, so you don't have to
think about escaping.

## Running the JSON API

```bash
devia serve                 # http://127.0.0.1:7654
```

Open the root URL in a browser for a quick reference of every
endpoint. Every endpoint accepts a `POST` with a JSON body, and
responds with the same envelope shape as the CLI's `--json` mode:

```bash
curl -s localhost:7654/api/v1/hash \
  -d '{"text":"hello","algo":"sha256"}' | jq
```

The endpoint list mirrors the command list above — `/api/v1/hash`,
`/api/v1/base64/encode`, `/api/v1/json/format`, `/api/v1/uuid`, and so
on (see the embedded `indexHTML` in `server.go`, or just visit `/`
after starting the server).

If you build the CLI-only variant (`-tags noserve`), the `serve`
command still exists, but it fails immediately with a clear message
explaining that this particular binary doesn't include the API,
rather than silently doing nothing.

## Project layout

```
devia/
├── main.go            command routing, --json flag extraction, help text
├── output.go           stdout/stderr/exit-code handling, stdin reading
├── flags.go             a small wrapper around flag.FlagSet
├── cmd_*.go               the CLI adapter for each command
├── server.go            (build tag: !noserve) the HTTP API
├── server_stub.go       (build tag: noserve)  a stand-in that reports the API isn't built in
└── core/                all business logic — zero external dependencies, shared by both the CLI and the API
    ├── core.go             the shared error/exit-code system
    ├── hash.go / checksum.go
    ├── base64.go / mime.go
    ├── jsonfmt.go
    ├── escape.go
    ├── uuid.go
    ├── text.go
    ├── timestamp.go
    ├── radix.go
    ├── cron.go
    ├── regex.go
    ├── diff.go
    ├── cert.go
    └── lorem.go
```

Every function in `core` is agnostic about whether it was called from
the CLI or an HTTP handler — that's intentional. The logic is written
once, and the CLI and API are both thin adapters on top of it.

## Testing

```bash
go vet ./...
go test ./...
make build && make build-min
```

(`make vet` and `make test` are equivalent shortcuts.)

The `core` package (where essentially all of the logic lives) has a
test suite covering each command's happy path, its edge cases, and its
error handling. The CLI and API layers are thin wiring on top of
`core` and aren't separately tested beyond the CI smoke test that
exercises the built binaries directly.