# devia

一个零依赖、单二进制的开发者工具集。CLI 优先设计，兼带一个可选的本地 JSON API。

## ⚠️ 关于这份代码的诚实声明

这份代码是在一个**没有网络、没有装 Go** 的沙箱里写的——我没法在这里 `go build` 或 `go vet` 来实际验证它。所有函数逻辑我都手工过了一遍（括号配对、import 是否用到、边界情况），但**你拿到手的第一件事应该是本地跑一遍**：

```bash
go vet ./...   # 静态检查，比 build 更容易先抓到问题
go build -o devia .
./devia hash --algo=sha256 "test"
```

如果报错，把错误信息发我，我立刻改。没有为了显得"完成度高"而假装这是编译验证过的成品。

## 设计上和上一版的区别

上一版用了 `urfave/cli` 和 `gin`——这是错的，等于给一个体积敏感的工具集背了 gin 的整条依赖链（json-iterator、validator、protobuf、yaml.v3……）。这次重写：

- **`go.mod` 里没有一行 `require`。** 全部标准库：`flag` 做命令行解析，`net/http`（Go 1.22 内置的方法路由 `mux.HandleFunc("POST /path", ...)`）做 API，不需要任何路由框架。
- **`net/http` 只存在于 `server.go`，且被 build tag 隔离。** `core` 包（所有业务逻辑）完全不碰 `net/http`——连 base64 图片的 MIME 检测都是自己写的十几行魔数嗅探（`core/mime.go`），不用 `http.DetectContentType`，因为那个函数会把整个 `net/http`（以及它牵连的 `crypto/tls`）链接进本该是纯 CLI 的二进制里。这就是为什么下面的 `devia-cli`（`-tags noserve`）能明显小于完整版——不是靠事后裁剪，是从依赖图设计阶段就切开的。
- **UUID 用 `crypto/rand`**，不是 `math/rand` 拼出来的伪随机——它标识东西，就该是真随机。
- **Cron、Diff 都是手写的**，没有引 `robfig/cron` 或 `sergi/go-diff`。Cron 是标准的字段解析 + 步进查找；Diff 是经典 LCS 回溯，O(n·m)，对几千行以内的文本足够快，超过阈值会直接报错而不是吃满内存。

## 脚本调用契约（这是这次的重点）

### 输出分离
- **stdout**：纯结果，没有多余的提示语。`devia hash --algo=md5 "x"` 输出就是那一行哈希，可以直接 `X=$(devia hash ...)`。
- **stderr**：所有错误和诊断信息，不管是不是 `--json` 模式都会写。
- **`--json` 模式**：stdout 变成单行 JSON，成功 `{"ok":true,"result":...}`，失败 `{"ok":false,"error":"...","code":N}`。可以直接接 `jq`。

```bash
devia --json hash --algo=sha256 "test" | jq -r .result
```

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

需要 [Go 1.22+](https://go.dev/dl/)（用了 `net/http` 的内置方法路由，早期版本没有这个 API）。

```bash
git clone <this repo>
cd devia

make build       # 完整版：devia   (CLI + serve)
make build-min   # 精简版：devia-cli (无 API，体积最小)
make build-all   # 全平台交叉编译，两个变体各 6 个目标
```

Windows 用 `build.bat`。

编译参数固定加了 `-trimpath -ldflags="-s -w"`（去符号表、去 DWARF 调试信息、去文件路径），这是无损的体积优化，Go 二进制通常能砍掉 25-35%。

### 想要更小？UPX（可选，不是默认）

```bash
upx --best --lzma devia
```

没有默认接入 build 流程，因为 UPX 压缩后的二进制偶尔会被杀毒软件误报（壳特征），你自己权衡要不要用。

## 体积参考

Go 运行时本身有固定开销（~1-1.5MB，跟你写多少代码无关），这是 Go 语言本身的性质，Rust 在这方面会更小。在这个前提下：

| 变体 | 大概量级 |
|---|---|
| `devia-cli`（`-tags noserve`，无 net/http） | 最小，纯 stdlib CLI |
| `devia`（含 `serve`，net/http + crypto/tls 链接进来） | 比 cli 版大，但仍然是个位数 MB |

具体字节数请用 `make size` 在你本地量——我在沙箱里没法编译出真实数字，不想编一个看起来精确实则瞎猜的数字给你。

## 命令一览

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
radix       [--from=N] <number>                        (自动识别 0x/0o/0b 前缀)
cron        [--next=N] <expr>                           (5 或 6 段表达式)
regex       test --pattern=.. [--flags=ims] [text]
            replace --pattern=.. --with=.. [text]
diff        --a=file --b=file | <textA> <textB>
cert        decode <file>                               (或 stdin 管道 PEM)
serve       [--host=127.0.0.1] [--port=7654]            启动 JSON API
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

完整端点列表见 `server.go` 里嵌入的 `indexHTML`，或直接启动后访问首页。

`devia-cli`（`-tags noserve` 编译出来的版本）里 `serve` 命令仍然存在，但会明确报错说这个二进制没打包 API，而不是假装成功——避免"用户以为端口起来了但其实没有"的糊涂账。

## 项目结构

```
devia/
├── main.go              路由分发、--json 提取、帮助文本
├── output.go             标准输出/错误/退出码/stdin 读取
├── flags.go               flag.FlagSet 封装
├── cmd_*.go                各命令的 CLI 适配层
├── server.go              (build tag: !noserve) HTTP API
├── server_stub.go         (build tag: noserve)  API 桩实现
└── core/                  纯业务逻辑，零外部依赖，CLI 和 API 共用
    ├── core.go             错误码体系
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

`core` 包里的每个函数都不知道自己是被 CLI 调用还是被 HTTP handler 调用——这是刻意的：逻辑只写一遍，两个入口都是薄适配层。

## 已知的取舍（不是 bug，是选择）

- **正则用 Go 的 `regexp`（RE2）**，不支持反向引用和环视断言。换来的是保证线性时间匹配、没有 ReDoS 风险——对一个可能拿去测试不可信输入的工具，这个权衡是对的。
- **Diff 是行级的**，不是字符级的。O(n·m)，几千行封顶，超过会报错而不是卡死或吃满内存。
- **Cron 的下次运行时间是暴力步进查找**（按秒或按分钟），设了 4 年的安全上限。对绝大多数表达式几毫秒内出结果；对"2 月 30 日"这种永不匹配的表达式会在 4 年上限后返回 3 号错误而不是死循环。
- **API 不做文件上传（multipart）**，二进制内容一律走 JSON 里的 base64 字段。保持 API 纯 JSON 是刻意的，代价是大文件走 API 会因为 base64 膨胀 33% 体积——大文件建议用 CLI 的 `--file` 直接读盘（会流式处理，不会整个读进内存）。

## License

MIT
