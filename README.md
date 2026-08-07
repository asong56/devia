# devia

一个零依赖、单二进制的开发者工具集。CLI 优先设计，兼带一个可选的本地 JSON API。

## ⚠️ 关于这份代码的诚实声明

这次重构（拆分目录结构、加 CI）同样是在一个**没有网络、没有装 Go** 的沙箱里做的——我没法在这里 `go build` 或 `go vet` 来实际验证。每个文件我都逐行手工核对过（import 路径、包名、括号配对、跨包引用是否一致），但**你拿到手的第一件事应该是本地跑一遍**：

```bash
go vet ./...
go build ./cmd/devia
./devia hash --algo=sha256 "test"
```

如果报错，把完整错误信息发我，我立刻改——不会因为"看起来应该没问题"就假装这是编译验证过的成品。项目里有 `.github/workflows/release.yml`，只要推到 GitHub 就会自动跑 `go vet` + `go build`，以后编译是否通过不用再靠人眼审查或本地环境去赌。

## 这次改了什么（相对上一版）

上一版所有代码平铺在仓库根目录（`main.go`、`flags.go`、`cmd_*.go`、`server.go` 全部挤在一起，`package main` 一个包装了路由分发、参数解析、输出格式化、七八个命令的实现和 HTTP API）。这次按职责拆开：

- **`cmd/devia/`**：只有入口，三行代码，`main()` 调 `cli.Run()`，别的什么都不做。
- **`internal/cli/`**：命令行层——参数解析、`--json`/退出码约定、每个命令的 CLI 适配器（按领域分了 `hash.go`／`encode.go`／`text.go`／`time.go`／`data.go`／`serve.go`，不再是不分类的 `cmd_*.go`）。
- **`internal/server/`**：可选的 HTTP API，`server.go`（真实实现，`!noserve` 标签）和 `stub.go`（`noserve` 标签下的桩实现）——和上一版的隔离机制完全一样，只是挪进了自己的目录。
- **`internal/core/`**：业务逻辑，内容未变，只是从 `core/` 挪到 `internal/core/`，加 `internal/` 前缀是为了让 Go 工具链在语言层面强制这些包不会被仓库之外的代码导入（这是个纯 CLI 工具，本来就不该有对外的 Go API）。
- **`internal/version/`**：单独一个只有一行常量的包，避免 `cli` 和 `server` 互相导入造成循环依赖，同时保证版本号只有一处定义。
- **`.github/workflows/release.yml`**：见下面「CI / 发布」。

上一轮（拆目录、加 CI）用户可见的命令行为一个字节没变。这一轮不一样：对照你发的 `CLI-Craftsmanship.md` 做了一次检查，补了几个真实缺口——新增了 `-q/--quiet`、`devia completion`、`base64 decode --dry-run`，错误信息在 stderr 上也多了一行"下一步该干嘛"。都是**纯新增**：没有任何旧 flag、旧退出码、旧输出格式被改名或删除，老脚本不会因为这轮改动而失效。具体对照见下面「对照 CLI Craftsmanship 检查表」。

## 对照 CLI Craftsmanship 检查表

你发的文档每一条我都过了一遍，附实际情况——不是每条都改了代码，有几条是"已经满足"或"权衡后不做"，也如实写出来：

| 文档要求 | 现状 | 说明 |
|---|---|---|
| `-h/--help` | ✅ 已有 | `help`/`-h`/`--help` 三种写法都行 |
| `-v/--version` | ✅ 已有 | `version`/`-v`/`--version` |
| `-q/--quiet` | ✅ 这轮新增 | 压制非必要的 stderr 提示（目前只有 `serve` 的启动横幅），**从不压制错误**——quiet 工具静默吞掉失败比吵闹的工具更危险，所以这条线没得商量 |
| `--dry-run` | ⚠️ 只加在 `base64 decode --out` 上 | devia 绝大多数命令本来就是无副作用的纯函数（hash、json format、text 等——运行本身就是"预览"，没有"真正执行"和"预演"的区别）。唯一有副作用的操作是"把解码后的字节写到文件"，这条加了 `--dry-run`；`serve` 启动服务没加，因为绑端口这件事本身检查一下 host/port 参数是否合法就行，不构成一个值得"预演"的破坏性操作。到处都加 `--dry-run` 反而违反文档自己说的"不是加更多功能" |
| stdin/stdout/stderr 分离 | ✅ 已满足 | `stdout` 只出结果，`stderr` 只出错误/诊断，管道场景本来就这么设计的（见「脚本调用契约」） |
| exit code | ✅ 已满足，且更细 | 不只是 0/1，细分成 0/1/2/3/4（成功/内部错误/用法错误/输入非法/未找到），脚本可以按错误类型分支，不用 grep 文本 |
| Help 一屏、常用优先 | ⚠️ 部分满足 | `Commands` 表紧跟在 `Usage` 后面（先给命令表，后给细节说明），符合"常用优先"；但完整 help 有 40 多行，中小终端要滚一下——内容密度已经很高（没有营销话术），要压到严格一屏需要拆成 `devia help` 简版 + `devia help <command>` 详版两层，目前没做，这轮先加了 `Global flags` 这块小节让 flag 可发现性更好 |
| Error Design（what/why/how to recover） | ✅ 这轮改了 | 之前是单行 `devia: error: <msg>`；现在错误第一行格式不变（老脚本 `grep '^devia: error:'` 还能匹配），但下面多了空行 + `Try:` + 下一步建议，跟文档给的例子结构一致 |
| Shell 自动补全 | ✅ 这轮新增 | `devia completion bash\|zsh\|fish`，零依赖（标准库 `fmt`/`strings` 拼字符串），只补全一级子命令名字，不做 flag 级补全——理由同 `--dry-run`：flag 级补全意味着要把每个 `FlagSet` 的所有 flag 同步进补全脚本并永远保持同步，收益远低于"至少能把命令名字面全"这个基本需求 |
| glob 输入 | ✅ 已满足（免费的） | 所有吃文件路径的命令（`checksum`、`json --file`、`diff --a/--b`、`cert decode`）走的是普通位置参数/flag，shell 自己会展开 `*.json` 这种 glob，devia 不需要自己实现 |
| 无配置/无账号/无云依赖 | ✅ 已满足 | 没有配置文件、没有登录状态、`serve` 是完全本地可选的，`download → run → done` |
| 恶意输入 / 边界情况 | ✅ 这轮补了测试 | 184 个测试函数，19 个 `_test.go` 文件，覆盖 `internal/core` 全部 16 个业务文件 + `internal/cli` 的纯函数部分 + 一套通过子进程重新执行自身二进制的黑盒测试（`cmd/devia/main_test.go`，专门测 `os.Exit`/退出码/`--json`/stdin 管道这些没法在进程内直接测的行为）。写测试的过程中真的挖出并修了两个此前没发现的 bug（见下面「这轮写测试时顺手修的」）。**诚实的边界**：这份测试代码本身没有在这个沙箱里真正跑过 `go test`——没有 Go 工具链——我做的是逐行人工核对 + 用 Python 独立算出每个断言里的期望值（`hashlib`/`base64`/`datetime` 算哈希和时间戳，手写一份等价的 diff 算法在 Python 里跑出预期输出，用真实 `openssl` 生成测试证书再核对字段），但"这些测试真的能通过"这句话我没法拍胸脯保证，第一次跑 `go test ./...` 请把结果发我 |

## 这轮写测试时顺手修的

写测试不是照抄代码逻辑再断言一遍——真正有用的测试会逼你去想"这里还有没有没考虑到的情况"，这轮真的碰到了两个：

1. **`json format`/`json minify` 和 `json validate` 对同一份畸形输入给出矛盾的结论。** 三个命令底层调用的是 `encoding/json` 的两套不同 API：`validate` 用 `json.Unmarshal`，天然拒绝"一个合法 JSON 值后面跟着多余内容"（比如 `{"a":1} garbage` 或者两个粘在一起的 JSON 对象）；但 `format`/`minify` 用的是 `json.Decoder.Decode`，这个 API 是为读取*一串*JSON 值设计的，`Decode` 一次只消费第一个值，后面的内容不报错、直接忽略。结果就是 `devia json validate` 正确拒绝这种输入，而 `devia json format` 会不声不响地只格式化第一个片段——同一个工具里两个命令对同一份输入给出不一致的判断。修法：加了个 `dec.More()` 检查，`format`/`minify` 现在和 `validate` 一样会拒绝这种输入。这个 bug 是在给 `jsonfmt_test.go` 写"两个对象粘在一起该报错"这个用例时发现的，纯人工审查代码没看出来，因为两段代码单独看都"没问题"。
2. **`--tz=Asia/Shanghai` 这种具名时区在没装系统时区数据库的机器上会直接报错。** `time.LoadLocation` 依赖 `/usr/share/zoneinfo`，而 devia 是 `CGO_ENABLED=0` 静态编译的，一个常见部署场景就是丢进 `scratch`/`distroless` 这类没有任何系统文件的最小化容器镜像——这种镜像里 `/usr/share/zoneinfo` 根本不存在，`timestamp`/`cron` 一用命名时区就炸。修法：在 `cmd/devia/main.go` 里空白导入了标准库的 `time/tzdata`，把整个时区数据库（~450KB）直接编进二进制，不再依赖部署环境。这个是写 `timestamp_test.go` 之前，为了确认测试环境本身有没有时区数据库时顺带查出来的，跟测试断言本身没关系，但性质上是同一轮"认真核实一遍再动笔"带出来的副产品。

## 设计上和更早版本的区别

更早一版用了 `urfave/cli` 和 `gin`——这是错的，等于给一个体积敏感的工具集背了 gin 的整条依赖链（json-iterator、validator、protobuf、yaml.v3……）。现在：

- **`go.mod` 里没有一行 `require`。** 全部标准库：`flag` 做命令行解析，`net/http`（Go 1.22 内置的方法路由 `mux.HandleFunc("POST /path", ...)`）做 API，不需要任何路由框架。
- **`net/http` 只存在于 `internal/server/server.go`，且被 build tag 隔离。** `internal/core` 包（所有业务逻辑）完全不碰 `net/http`——连 base64 图片的 MIME 检测都是自己写的十几行魔数嗅探（`internal/core/mime.go`），不用 `http.DetectContentType`，因为那个函数会把整个 `net/http`（以及它牵连的 `crypto/tls`）链接进本该是纯 CLI 的二进制里。这就是为什么下面的 `devia-cli`（`-tags noserve`）能明显小于完整版——不是靠事后裁剪，是从依赖图设计阶段就切开的，这次重构原样保留了这个切法。
- **UUID 用 `crypto/rand`**，不是 `math/rand` 拼出来的伪随机——它标识东西，就该是真随机。
- **Cron、Diff 都是手写的**，没有引 `robfig/cron` 或 `sergi/go-diff`。Cron 是标准的字段解析 + 步进查找；Diff 是经典 LCS 回溯，O(n·m)，对几千行以内的文本足够快，超过阈值会直接报错而不是吃满内存。

## 脚本调用契约

### 输出分离
- **stdout**：纯结果，没有多余的提示语。`devia hash --algo=md5 "x"` 输出就是那一行哈希，可以直接 `X=$(devia hash ...)`。
- **stderr**：所有错误和诊断信息，不管是不是 `--json` 模式都会写。`-q/--quiet` 能压掉非必要的提示（目前只有 `serve` 的启动横幅），但**错误永远不会被压掉**——`--quiet` 和 `--json` 一样是全局 flag，前后都能放：`devia -q serve` 或 `devia serve -q` 效果一样。
- **`--json` 模式**：stdout 变成单行 JSON，成功 `{"ok":true,"result":...}`，失败 `{"ok":false,"error":"...","code":N}`。可以直接接 `jq`。

```bash
devia --json hash --algo=sha256 "test" | jq -r .result
```

### 错误格式
非 `--json` 模式下，出错时 stderr 是固定形状——第一行说清楚出了什么事，空一行后跟一行 `Try:` 给下一步：

```
$ devia hash --file nope.txt
devia: error: file not found: nope.txt

Try:
  check the path and try again
```

第一行格式（`devia: error: <msg>` / `devia: usage error: <msg>`）保持稳定，老脚本里的 `grep '^devia: error:'` 不会因为这次改动而失效——`Try:` 是纯新增的两行。

### 标准退出码
| 码 | 含义 |
|---|---|
| 0 | 成功 |
| 1 | 内部/运行时错误 |
| 2 | 用法错误（缺参数、未知子命令） |
| 3 | 输入数据非法（比如坏的 JSON、坏的 base64、坏的正则） |
| 4 | 文件/资源不存在 |

这样脚本可以按错误类型分支，而不是 grep stderr 文本：

```bash
devia json validate < maybe.json
case $? in
  0) echo "valid" ;;
  3) echo "bad json, skipping" ;;
  *) echo "unexpected failure" >&2; exit 1 ;;
esac
```

### stdin 管道
最后一个位置参数留空时自动读 stdin（非终端才读，不会卡住等待输入）：

```bash
echo -n hello | devia hash --algo=sha256
cat cert.pem | devia cert decode --json
```

### 从其他语言调用

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

用 `execFile`/`subprocess.run` 而不是 shell 字符串拼接，参数按数组传，不用担心特殊字符转义问题。

## 安装 / 编译

需要 [Go 1.26+](https://go.dev/dl/)（`go.mod` 里的 `go` 指令定了这个下限；`net/http` 的内置方法路由从 Go 1.22 开始就有，这次升级单纯是紧跟官方最新稳定版）。

```bash
git clone <this repo>
cd devia

go test ./...    # 跑测试（184 个测试函数）；加 -v 看逐个用例，加 -race 跑竞态检测
make build       # 完整版：devia   (CLI + serve)
make build-min   # 精简版：devia-cli (无 API，体积最小)
make build-all   # 全平台交叉编译，两个变体各 6 个目标
```

Windows 用 `build.bat`。

编译参数固定加了 `-trimpath -ldflags="-s -w"`（去符号表、去 DWARF 调试信息、去文件路径），这是无损的体积优化，Go 二进制通常能砍掉 25-35%；这次额外统一加了 `CGO_ENABLED=0`（纯静态、不依赖系统 libc，交叉编译也更可靠，之前只有 `build.sh`/`build.bat` 有这条，`make build`/`make build-min` 缺了，这次补齐）。

### 想要更小？UPX（可选，不是默认）

```bash
upx --best --lzma devia
```

没有默认接入 build 流程，因为 UPX 压缩后的二进制偶尔会被杀毒软件误报（壳特征），你自己权衡要不要用。

## CI / 发布（`.github/workflows/release.yml`）

三种触发方式：

- **每次 push / PR**：只跑 `go vet` + 用当前平台把两个变体（完整版、`-tags noserve`）都编译一遍，几十秒内出结果——这就是"打包是否失败"这个问题以后该由 CI 回答，而不是靠人肉审查代码。
- **手动触发（workflow_dispatch）**：去 GitHub 仓库的 Actions 标签页，选这个 workflow，点 "Run workflow"，可以勾选 "publish_release" 来顺便发一个带全部 3 个 zip 的 draft release。
- **推送 `v*` 标签**（比如 `git tag v1.0.0 && git push --tags`）：自动编译发布用的单一二进制（完整版，含 `serve`），打包成 3 个 zip 并发布 draft release，你在 GitHub 上确认后手动 publish：

  | zip | 平台 |
  |---|---|
  | `devia-darwin-arm64.zip` | macOS，Apple Silicon |
  | `devia-linux-amd64.zip` | Linux，x86_64 |
  | `devia-windows-x64.zip` | Windows，x86_64 |

  每个 zip 里只有一个文件——`devia`（Windows 下是 `devia.exe`），解压即用。这 3 个是目前唯一自动发布的平台；`make build-all` / `build.sh` 本地仍然可以出全部 6 个平台 × 2 个变体，CI 发布流程为了简单只挑了最常用的这 3 个和最完整的那个变体。想让 CI 也发布其他平台或 `-cli` 精简版，在 `release.yml` 的 `matrix.include` 里加一行就行。

## 体积参考

Go 运行时本身有固定开销（~1-1.5MB，跟你写多少代码无关），这是 Go 语言本身的性质，Rust 在这方面会更小。在这个前提下：

| 变体 | 大概量级 |
|---|---|
| `devia-cli`（`-tags noserve`，无 net/http） | 最小，纯 stdlib CLI |
| `devia`（含 `serve`，net/http + crypto/tls 链接进来） | 比 cli 版大，但仍然是个位数 MB |

具体字节数请用 `make size` 在你本地量，或者看 CI 里 "print sizes" 那一步的输出——我在沙箱里没法编译出真实数字，不想编一个看起来精确实则瞎猜的数字给你。

## 命令一览

```
devia [flags] <command> [subcommand] [flags] [args]

全局 flag: -h/--help  -v/--version  -q/--quiet  --json

hash        --algo=md5|sha1|sha256|sha512 [--hmac=key] [--base64] [--file=path] [text]
checksum    [--algo=..] [--compare=hash] <file>
base64      encode|decode [--file=path] [--out=path] [--dry-run] [--data-uri] [--url] [text]
json        format|minify|validate [--indent="..."] [--file=path] [text]
escape      json|url|url-path|html|unicode [text]
unescape    json|url|url-path|html|unicode [text]
uuid        [--count=N] [--upper]
text        <mode> [text]
            modes: lower upper sentence title camel pascal snake
                   constant kebab cobol train alternating inverse
lorem       [--type=word|sentence|paragraph] [--count=N] [--classic]
timestamp   now | to-date <unix> | from-date <date>   [--tz=..] [--format=..]
radix       [--from=N] <number>                        (自动识别 0x/0o/0b 前缀)
cron        [--next=N] <expr>                           (5 或 6 段表达式)
regex       test --pattern=.. [--flags=ims] [text]
            replace --pattern=.. --with=.. [text]
diff        --a=file --b=file | <textA> <textB>
cert        decode <file>                               (或 stdin 管道 PEM)
serve       [--host=127.0.0.1] [--port=7654]            启动 JSON API
completion  bash|zsh|fish                               打印 shell 补全脚本
```

`--dry-run` 目前只在 `base64 decode --out` 上有意义——devia 大部分命令本来就是无副作用的纯转换（跑一次就是结果，没有"预演"和"真跑"的区别），唯一真正写文件的操作是这条，加了 `--dry-run` 之后只报告"会写多少字节到哪"而不实际写。

`devia completion bash|zsh|fish` 打印一段对应 shell 的补全脚本（只补全一级命令名，不做 flag 级补全）：

```bash
source <(devia completion bash)          # 当前 shell 里试用
devia completion zsh > "${fpath[1]}/_devia"   # 持久化到 zsh
devia completion fish | source            # fish 当前 shell里试用
```

### 示例

```bash
devia hash --algo=sha256 "hello"
devia hash --file big.iso --algo=sha256
devia checksum installer.exe --compare abc123...   # exit 0=match 1=mismatch

devia base64 encode --file logo.png --data-uri      # data:image/png;base64,...
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

devia radix 0xFF                      # bin/oct/dec/hex 四进制同时输出
devia cron "*/15 * * * *" --next 3

devia regex test --pattern '\d+' "room 42, floor 3"
devia diff --a old.txt --b new.txt

devia cert decode server.crt
```

## `devia serve`：JSON API

```bash
devia serve                 # http://127.0.0.1:7654
```

浏览器打开根路径能看到端点表；所有端点 `POST` JSON，响应信封和 CLI 的 `--json` 模式完全一致：

```bash
curl -s localhost:7654/api/v1/hash \
  -d '{"text":"hello","algo":"sha256"}' | jq
```

完整端点列表见 `internal/server/server.go` 里嵌入的 `indexHTML`，或直接启动后访问首页。

`devia-cli`（`-tags noserve` 编译出来的版本）里 `serve` 命令仍然存在，但会明确报错说这个二进制没打包 API，而不是假装成功——避免"用户以为端口起来了但其实没有"的糊涂账。

## 项目结构

```
devia/
├── cmd/devia/
│   ├── main.go               入口：main() 调 main2()，main2() 调 internal/cli.Run()
│   └── main_test.go          黑盒测试：子进程重跑自身二进制，测 exit code/--json/stdin/--quiet
├── internal/
│   ├── cli/                  命令行层
│   │   ├── run.go             路由分发、--json/--quiet 提取、帮助文本
│   │   ├── run_test.go
│   │   ├── output.go           标准输出/错误(what+Try:)/退出码/stdin 读取
│   │   ├── flags.go             flag.FlagSet 封装
│   │   ├── hash.go               hash / checksum
│   │   ├── encode.go              base64（含 --dry-run）/ json / escape / unescape
│   │   ├── text.go                 text / uuid / lorem
│   │   ├── time.go                  timestamp / cron
│   │   ├── data.go                   radix / regex / diff / cert
│   │   ├── serve.go                   serve（转发到 internal/server）
│   │   ├── completion.go               bash/zsh/fish 补全脚本生成
│   │   └── completion_test.go
│   ├── server/                HTTP API
│   │   ├── server.go           (build tag: !noserve) 真实实现
│   │   └── stub.go               (build tag: noserve)  桩实现
│   ├── core/                  纯业务逻辑，零外部依赖，CLI 和 API 共用；每个 .go 都有对应 _test.go
│   │   ├── core.go / core_test.go             错误码体系
│   │   ├── hash.go / hash_test.go
│   │   ├── checksum.go / checksum_test.go
│   │   ├── base64.go / base64_test.go
│   │   ├── mime.go / mime_test.go
│   │   ├── jsonfmt.go / jsonfmt_test.go
│   │   ├── escape.go / escape_test.go
│   │   ├── uuid.go / uuid_test.go
│   │   ├── text.go / text_test.go
│   │   ├── timestamp.go / timestamp_test.go
│   │   ├── radix.go / radix_test.go
│   │   ├── cron.go / cron_test.go
│   │   ├── regex.go / regex_test.go
│   │   ├── diff.go / diff_test.go
│   │   ├── cert.go / cert_test.go
│   │   └── lorem.go / lorem_test.go
│   └── version/               单一版本号常量，避免 cli/server 循环依赖
│       └── version.go
├── .github/workflows/
│   └── release.yml            CI：push/PR 跑 vet+test+build；tag 或手动触发时编译 3 个平台 zip 并发布
├── Makefile                  test/test-race/build/build-min/build-all/size/vet/clean
├── build.sh / build.bat
├── RELEASE_NOTES.md           GitHub Release 正文，release.yml 自动引用
└── go.mod
```

`internal/core` 包里的每个函数都不知道自己是被 CLI 调用还是被 HTTP handler 调用——这是刻意的：逻辑只写一遍，两个入口都是薄适配层。

⚠️ 一个行为变化：`make build`/`build-min`/`build-all` 现在都依赖 `test` 这个 target，也就是说本地 `make build` 会先跑一遍完整测试套件再编译——编译变慢了，但也意味着一个跑不过测试的版本不会被 `make build` 悄悄编出来。只想编译不想等测试，直接跑 `go build ./cmd/devia`（跳过 Makefile）或者 `CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o devia ./cmd/devia`（Makefile 里 `build` target 展开后的原始命令）。

## 已知的取舍（不是 bug，是选择）

- **正则用 Go 的 `regexp`（RE2）**，不支持反向引用和环视断言。换来的是保证线性时间匹配、没有 ReDoS 风险——对一个可能拿去测试不可信输入的工具，这个权衡是对的。
- **Diff 是行级的**，不是字符级的。O(n·m)，几千行封顶，超过会报错而不是卡死或吃满内存。
- **Cron 的下次运行时间是暴力步进查找**（按秒或按分钟），设了 4 年的安全上限。对绝大多数表达式几毫秒内出结果；对"2 月 30 日"这种永不匹配的表达式会在 4 年上限后返回 3 号错误而不是死循环。
- **API 不做文件上传（multipart）**，二进制内容一律走 JSON 里的 base64 字段。保持 API 纯 JSON 是刻意的，代价是大文件走 API 会因为 base64 膨胀 33% 体积——大文件建议用 CLI 的 `--file` 直接读盘（会流式处理，不会整个读进内存）。

## License

MIT
