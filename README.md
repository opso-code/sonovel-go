# sonovel-go

Go 语言实现的 So Novel 核心功能版本（规则驱动抓取）。

## 功能概览

- 规则加载：读取 `./rules/*.json` 或指定规则文件。
- 搜索：按书源规则搜索书名/作者。
- 搜索后直下：`search` 命令搜索完成后可直接输入序号下载。
- 抓取：解析书籍详情、目录、章节正文。
- 下载：支持并发抓章节、失败重试、章节范围下载。
- 导出：`txt`（默认）、`epub`、`html`。
- 交互方式：默认 Web UI（可选 TUI/CLI 高级模式）。

## 当前限制

- 优先支持 CSS 选择器规则。
- `xpath` 规则当前未实现（后续可补 `htmlquery` 方案）。
- `@js:` 已支持规则内联 JS 执行（用于正文解密/清洗等场景）。
- 暂未实现 PDF 导出。

## 快速开始

```bash
cd sonovel-go
make test
```

## 命令说明

### 1) 默认启动（推荐）

```bash
go run ./cmd/sonovel --config ./config.toml --port 7765
```

程序会启动 Web，并按配置自动打开浏览器（可用 `--no-browser` 关闭）。

### 2) 搜索（CLI，支持选行直接下载）

```bash
go run ./cmd/sonovel search --kw "斗罗大陆" --config ./config.toml
```

### 3) 下载（CLI）

```bash
go run ./cmd/sonovel download \
  --url "https://www.shuhaige.net/70475/" \
  --format txt \
  --config ./config.toml \
  --concurrency 12
```

### 4) 章节范围下载

```bash
go run ./cmd/sonovel download \
  --url "https://www.shuhaige.net/70475/" \
  --start 1 --end 100 \
  --format txt \
  --config ./config.toml
```

### 5) 终端交互模式（TUI）

```bash
go run ./cmd/sonovel --tui --config ./config.toml
```

## Makefile

```bash
make help
make run-tui
make run-web
make build
make build-all
```

`build-all` 默认生成：
- linux: amd64
- darwin: amd64
- windows: amd64

产物目录：`./dist`

压缩包内包含：
- 可执行文件
- `config.toml`
- `rules/`

用户下载对应平台压缩包后，解压即可直接运行，无需执行 `make init`。

## 参数要点

- `--config`：配置文件路径（默认 `./config.toml`）。
- `--rules`：规则文件路径（默认 `./rules/main.json`）。
- `--format`：`txt|epub|html`，默认 `txt`。
- `--out`：下载输出目录，默认 `./downloads`。
- `--concurrency`：并发数，默认 `12`。
- `--start --end`：章节范围（`end=0` 表示到最后一章）。

## 配置文件

默认读取 `./config.toml`。示例：

```toml
rules_file = "./rules/main.json"
output_dir = "./downloads"
format = "txt"
source_id = 1
search_limit = 20
concurrency = 12
chapter_start = 1
chapter_end = 0
min_interval_ms = 200
max_interval_ms = 450
enable_retry = true
max_retries = 3
retry_min_ms = 1000
retry_max_ms = 2000

[web]
port = 7765
open_browser = true
```

优先级：命令行参数 > `config.toml` > 程序默认值。

## 目录结构

- `cmd/sonovel`: 程序入口（默认 Web；可用 `--tui` / `search` / `download` / `init`）。
- `internal/rule`: 规则加载与默认值处理。
- `internal/parser`: 搜索/详情/目录/章节解析。
- `internal/crawler`: 并发抓取与重试调度。
- `internal/exporter`: `txt/html/epub` 导出。
- `internal/tui`: 终端交互界面。
- `internal/web`: Web 服务与静态页面。
- `internal/httpx`: HTTP 客户端与 POST 模板解析。
- `rules`: 书源规则目录。

## License

本项目基于 `so-novel` 进行翻译与改造，遵循其 AGPL 许可要求。
