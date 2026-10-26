<p align="right"><a href="README.md">English</a> · <a href="README.de.md">Deutsch</a> · <b>简体中文</b></p>

<p align="center">
  <a href="https://maximilianfeix.github.io/shortfree/">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="brand/banner-dark.svg">
      <img src="brand/banner-light.svg" alt="shortfree：所有未注册的 1–4 字符域名，每日核验" width="800">
    </picture>
  </a>
</p>

<p align="center">
  <b>.com .net .org .io .ai .dev</b> 及<a href="docs/tlds.md">更多后缀</a>上所有未注册的 1–4 字符域名，<br>
  每天直接向注册局的权威域名服务器和 RDAP 服务核实。
</p>

<p align="center">
  <a href="https://maximilianfeix.github.io/shortfree/"><img src="https://img.shields.io/badge/%E6%90%9C%E7%B4%A2%E5%88%97%E8%A1%A8-1FC16B?style=for-the-badge" alt="搜索列表"></a>
  <a href="#lists"><img src="https://img.shields.io/badge/%E4%B8%8B%E8%BD%BD%E5%88%97%E8%A1%A8-0B0C0E?style=for-the-badge" alt="下载列表"></a>
  <a href="#try-it"><img src="https://img.shields.io/badge/%E5%AE%89%E8%A3%85-0B0C0E?style=for-the-badge" alt="安装"></a>
  <a href="#mcp"><img src="https://img.shields.io/badge/MCP-0B0C0E?style=for-the-badge" alt="MCP 服务器"></a>
  <a href="#api"><img src="https://img.shields.io/badge/API-0B0C0E?style=for-the-badge" alt="JSON 和 CSV API"></a>
</p>

<p align="center">
  <a href="https://github.com/maximilianfeix/shortfree/actions/workflows/scan.yml"><img src="https://github.com/maximilianfeix/shortfree/actions/workflows/scan.yml/badge.svg" alt="每日扫描"></a>
  <a href="https://github.com/maximilianfeix/shortfree/actions/workflows/ci.yml"><img src="https://github.com/maximilianfeix/shortfree/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/maximilianfeix/shortfree/releases"><img src="https://img.shields.io/github/v/release/maximilianfeix/shortfree?color=1FC16B" alt="最新版本"></a>
  <a href="https://pkg.go.dev/github.com/maximilianfeix/shortfree"><img src="https://pkg.go.dev/badge/github.com/maximilianfeix/shortfree.svg" alt="Go reference"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/maximilianfeix/shortfree?color=0B0C0E" alt="MIT 许可证"></a>
  <a href="https://maximilianfeix.github.io/shortfree/api/v1/badge.json"><img src="https://img.shields.io/endpoint?url=https%3A%2F%2Fmaximilianfeix.github.io%2Fshortfree%2Fapi%2Fv1%2Fbadge.json" alt="经 RDAP 核验的空闲短域名数量"></a>
</p>

<!-- demo -->

<p align="center">
  <a href="#try-it">试用</a> ·
  <a href="#numbers">数据</a> ·
  <a href="#lists">列表</a> ·
  <a href="#limits">局限</a> ·
  <a href="#how-it-works">原理</a> ·
  <a href="#mcp">MCP</a> ·
  <a href="#api">API</a> ·
  <a href="#cli">CLI</a> ·
  <a href="#faq">常见问题</a> ·
  <a href="#contributing">参与贡献</a>
</p>

<a id="try-it"></a>
## 试用

**无需安装：** 打开 [maximilianfeix.github.io/shortfree](https://maximilianfeix.github.io/shortfree/)，输入 1–4 个字符，即可看到它在所有已扫描后缀上的状态。完整列表可以按后缀、长度和字符类型筛选。

**命令行**（需要 Go 1.27 或更高版本）：

```sh
go install github.com/maximilianfeix/shortfree/cmd/shortfree@latest
shortfree check x7.dev acme.io
```

**Docker：**

```sh
docker run -i --rm ghcr.io/maximilianfeix/shortfree check x7.dev
```

**Claude Code**（MCP）：

```sh
claude mcp add shortfree -- docker run -i --rm ghcr.io/maximilianfeix/shortfree
```

每个 [Release](https://github.com/maximilianfeix/shortfree/releases) 都附带 Linux、macOS 和 Windows 的二进制文件。

<a id="numbers"></a>
## 当前数据

按后缀和长度统计的数量表在每次扫描后自动更新，见[英文 README](README.md#numbers)。原始数据见 [`stats.json`](https://maximilianfeix.github.io/shortfree/api/v1/stats.json)。

<a id="lists"></a>
## 只想要列表？

配套仓库 [free-short-domains](https://github.com/maximilianfeix/free-short-domains) 提供纯 CSV 和 TXT 文件，每天扫描结束后更新。直接下载：

| 文件 | 内容 |
|---|---|
| [`csv/all.csv`](https://raw.githubusercontent.com/maximilianfeix/free-short-domains/main/csv/all.csv) | 所有已扫描后缀上经 RDAP 核验的空闲 1–3 字符域名 |
| `csv/<tld>.csv`，例如 [`csv/dev.csv`](https://raw.githubusercontent.com/maximilianfeix/free-short-domains/main/csv/dev.csv) | 1–3 字符名称及其状态：`available`、`dropping` 或 `unverified` |
| `txt/<tld>.txt`，例如 [`txt/io.txt`](https://raw.githubusercontent.com/maximilianfeix/free-short-domains/main/txt/io.txt) | 已核验的空闲 1–3 字符域名，每行一个 |
| `four-letter/<tld>.txt`，例如 [`four-letter/com.txt`](https://raw.githubusercontent.com/maximilianfeix/free-short-domains/main/four-letter/com.txt) | 不在区域文件中的 4 字母名称（a–z） |

包含数字和连字符在内的全部 4 字符名称，以 `four-<tld>.txt.gz` 的形式放在 [`data` Release](https://github.com/maximilianfeix/shortfree/releases/tag/data) 中。同样的数据也可以通过 [JSON/CSV API](#api) 获取。所有数据均采用 CC0：无需署名，无需密钥。

```sh
# 所有空闲的 2 字符 .dev 域名
curl -s https://raw.githubusercontent.com/maximilianfeix/free-short-domains/main/txt/dev.txt | awk 'length($0) == 6'
```

<a id="limits"></a>
## 能做什么，不能做什么

shortfree 会：

- 检查不超过 4 个字符的每一个合法标签（a–z、0–9、中间的连字符）：每个后缀 1–3 字符共 49,284 个，4 字符另有 1,774,224 个；
- 直接查询该后缀的权威域名服务器，不经过任何解析器缓存；
- 对每个 1–3 字符候选名称，向注册局的 RDAP 服务确认（`404` = 未注册）；
- 将处于赎回期（redemption）或待删除（pending delete）状态的已注册名称单独列出，并附上到期日期；
- 以 CC0 协议发布纯文件，无需账号、密钥，也没有任何跟踪。

shortfree 不能：

- **看到价格或保留状态。** 注册局会保留部分名称，或以溢价出售另一些名称，而 RDAP 对两者都报告为“未注册”。这在除 .com、.net、.de 以外大多数后缀的 1–2 字符名称上很常见。实际价格以注册商显示为准。
- **用 RDAP 核验 4 字符名称。** 它们只经过 DNS 检查。已注册但没有设置域名服务器的名称，可能出现在 4 字符列表中。决定之前，请先用 `shortfree check` 确认。
- **比最近一次扫描更新。** 列表是每天的快照，某个名称可能在扫描几分钟后就被注册。`shortfree check` 和 MCP 工具 `check_domains` 会实时查询注册局。

<a id="how-it-works"></a>
## 原理

1. **扫描。** shortfree 针对每个标签，向该后缀的权威域名服务器发送 `NS` 查询。`NXDOMAIN` 表示该名称不在区域中。单个后缀的全部 1–3 字符名称约 5 万次查询，几分钟即可完成。
2. **核验。** 不在区域中并不等于可注册：被锁定、未设置域名服务器或处于赎回期的名称同样不会出现在区域中。因此每个 1–3 字符候选名称都会向注册局的 RDAP 服务查询（Verisign、PIR、DENIC、Identity Digital、Google Registry、CentralNic 等，见 [docs/tlds.md](docs/tlds.md)）。短于注册局最小长度的名称会被跳过。`404` = 未注册；`pending delete` 或 `redemption period` = 即将释放。
3. **重复。** [`scan.yml`](.github/workflows/scan.yml) 每天 03:17 UTC 运行，每个后缀一个任务，扫描全部 1–3 字符名称，以及 4 字符空间七个分片中的一个，因此每个 4 字符名称每周复查一次。RDAP 结果最迟每 10 天复查一次：从未查过的名称优先，其次是即将释放的名称，然后是最久未查的名称。

每个 RDAP 主机只有一个总速率限制（每秒几次请求），由该主机上的所有后缀和所有并行任务共享（见 [`internal/rdap/rdap.go`](internal/rdap/rdap.go)），以保持合理使用。扫描器状态保存在 `data` Release 中，git 历史里只有体积很小的公开列表。

<a id="mcp"></a>
## MCP 服务器

`shortfree mcp` 是一个 [MCP](https://modelcontextprotocol.io) 服务器，可用于 Claude、Cursor、VS Code 等客户端。它提供四个只读工具：

| 工具 | 作用 |
|---|---|
| `check_domains` | 对最多 50 个域名做实时 DNS + RDAP 检查，长度和后缀不限（所有已扫描后缀支持 RDAP） |
| `find_free_domains` | 按后缀、长度、字符类型、包含的子串或前缀搜索每日快照 |
| `suggest_domains` | 把一个词或项目名转换成短候选名称，并返回其中空闲的 |
| `domain_stats` | 按后缀和长度统计的数量 |

```sh
claude mcp add shortfree -- docker run -i --rm ghcr.io/maximilianfeix/shortfree
```

Claude Desktop、Cursor 和 VS Code 的配置方法，以及 `suggest_domains` 的评分规则，见 [docs/mcp.md](docs/mcp.md)（英文）。

<a id="api"></a>
## API

托管在 GitHub Pages 上的静态只读 JSON/CSV API，每次扫描后重新生成。无需密钥，支持 CORS。

```
https://maximilianfeix.github.io/shortfree/api/v1/
```

| 路径 | 内容 |
|---|---|
| `stats.json` | 按后缀和长度统计的数量 |
| `{tld}.json`、`{tld}/{n}.json` | 某个后缀上的 1–3 字符名称，可只取某一长度 |
| `dropping.json` | 处于赎回期或待删除的名称，最早到期的排在前面 |
| `changes.json` | 最近 30 天内释放或被注册的名称 |
| `free.csv` | 所有经 RDAP 核验的空闲 1–3 字符域名 |
| `badge.json` | 带当前数量的 [shields.io endpoint](https://shields.io/badges/endpoint-badge) |

```sh
curl -s https://maximilianfeix.github.io/shortfree/api/v1/io/3.json | jq -r '.available[] + ".io"'
```

全部接口、字段说明以及 Python 和 JavaScript 示例见 [docs/api.md](docs/api.md)（英文）。OpenAPI 规范：[docs/openapi.yaml](docs/openapi.yaml)。

<a id="cli"></a>
## CLI

```sh
shortfree check q7.tools x4k.app abc.de
```

```
DOMAIN    DNS    RDAP        VERDICT
q7.tools  free   available   FREE
x4k.app   taken  registered  taken
abc.de    taken  registered  taken
```

`check` 适用于任意长度、任意后缀的域名。对于 shortfree 尚不支持 RDAP 的后缀，只给出基于 DNS 的结论（`probably free`）。加上 `-json` 后，每个域名输出一个对象：

```sh
shortfree check -json q7.tools | jq '.[0].verdict'
```

其他命令用于运行扫描本身：`scan`（1–3 字符，DNS + RDAP）、`scan4`（4 字符的一个分片）、`export`（写入 `data/` 和数量表）、`merge` 和 `mcp`。`shortfree <命令> -h` 可查看参数。

### 在 GitHub Actions 中使用

本仓库同时也是一个 GitHub Action。它检查一组域名，可以在某个域名变为可注册时让任务失败，或者创建一个 issue：

```yaml
- uses: maximilianfeix/shortfree@main
  id: watch
  with:
    domains: acme.dev acme.io
    fail-when: never   # free | taken | never
- run: echo '${{ steps.watch.outputs.free }}'   # ["acme.dev"] 或 []
```

输入、输出以及每日监控工作流示例见 [docs/action.md](docs/action.md)（英文）。

<a id="faq"></a>
## 常见问题

**列表里显示某个名称空闲，但注册商说不可注册或是溢价域名，为什么？**
注册局保留了它，或者按溢价出售。RDAP 只说明名称是否已注册，不提供价格。这主要发生在 .com、.net、.de 以外后缀的 1–2 字符名称上。网站和 MCP 工具会把这类名称标记为“可能溢价”。

**数据有多新？**
扫描每天 03:17 UTC 开始。每个 JSON 文件都有 `updated` 字段，记录扫描时间。1–3 字符名称每天扫描，4 字符名称每周一次。

**为什么 4 字符名称不做 RDAP 核验？**
每个后缀有 1,774,224 个 4 字符名称。按 shortfree 使用的速率限制（每个注册局每秒几次请求，由其各后缀共享），仅一个后缀查一遍就需要数周。而 DNS 能在几分钟内为绝大多数名称回答同样的问题。

**为什么只支持这些后缀？**
每个后缀都需要一个能在适度速率下稳定响应的注册局 RDAP 服务，而且每增加一个后缀，每天要多出约 30 万次 DNS 查询（全部 1–3 字符名称加上 4 字符空间的七分之一）。当前列表及注册局和来源见 [docs/tlds.md](docs/tlds.md)；如何添加新后缀，见 [CONTRIBUTING.md](CONTRIBUTING.md)。

**shortfree 卖域名或跟踪用户吗？**
不。没有推广链接，没有账号，也没有统计分析。网站和 API 都是 GitHub Pages 上的静态文件。代码采用 MIT 许可，数据采用 CC0。

<a id="contributing"></a>
## 参与贡献

欢迎提交 issue 和 pull request。[CONTRIBUTING.md](CONTRIBUTING.md) 说明了如何添加新后缀。提交 PR 之前请运行：

```sh
go test ./...
go vet ./...
gofmt -l .
```

在本地运行一次小规模扫描：

```sh
go run ./cmd/shortfree scan -tlds tools -rdap-time 1m   # 写入 state/state.json.gz
go run ./cmd/shortfree export -tlds tools -readme ''      # 写入 data/
```

安全问题请见 [SECURITY.md](SECURITY.md)。

## 许可证

代码：[MIT](LICENSE)。数据：[CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/)。注册局数据来自公开的 DNS 和 RDAP 服务。
