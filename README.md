# sonovel-go

Go 语言实现的 So Novel 核心功能版本（规则驱动抓取）。

## 功能概览

- 规则加载：读取 `bundle/rules/*.json` 或指定规则文件。
- 搜索：按书源规则搜索书名/作者。
- 抓取：解析书籍详情、目录、章节正文。
- 下载：支持并发抓章节、失败重试、章节范围下载。
- 导出：`txt`（默认）、`epub`、`html`。
- 交互方式：CLI + TUI + Web UI。

## 当前限制

- 优先支持 CSS 选择器规则。
- `xpath` 规则当前未实现（后续可补 `htmlquery` 方案）。
- `@js:` 目前实现了常见字符串处理（如 `replace`、前缀拼接），复杂 JS 逻辑暂不完整。
- 暂未实现 PDF 导出。

## 快速开始

```bash
cd sonovel-go
GOPROXY=https://proxy.golang.org,direct go mod tidy
```

## 命令说明

### 1) 搜索（CLI）

```bash
go run ./cmd/sonovel search --kw "斗罗大陆" --source-id 1 --rules ../so-novel/bundle/rules/main.json
```

### 2) 下载（CLI）

```bash
go run ./cmd/sonovel download \
  --url "https://www.shuhaige.net/70475/" \
  --format txt \
  --out ./downloads \
  --rules ../so-novel/bundle/rules/main.json \
  --concurrency 12
```

### 3) 章节范围下载

```bash
go run ./cmd/sonovel download \
  --url "https://www.shuhaige.net/70475/" \
  --start 1 --end 100 \
  --format txt
```

### 4) 终端交互模式（TUI）

```bash
go run ./cmd/sonovel tui --rules ../so-novel/bundle/rules/main.json --out ./downloads
```

### 5) Web UI

```bash
go run ./cmd/sonovel web --port 7765 --rules ../so-novel/bundle/rules/main.json --out ./downloads
```

浏览器打开：`http://localhost:7765`

## 参数要点

- `--rules`：规则文件路径（默认 `../bundle/rules/main.json`，建议按你的目录改成 `../so-novel/bundle/rules/main.json`）。
- `--format`：`txt|epub|html`，默认 `txt`。
- `--out`：下载输出目录，默认 `./downloads`。
- `--concurrency`：并发数，默认 `12`。
- `--start --end`：章节范围（`end=0` 表示到最后一章）。

## 目录结构

- `cmd/sonovel`: 程序入口（`search` / `download` / `tui` / `web`）。
- `internal/rule`: 规则加载与默认值处理。
- `internal/parser`: 搜索/详情/目录/章节解析。
- `internal/crawler`: 并发抓取与重试调度。
- `internal/exporter`: `txt/html/epub` 导出。
- `internal/tui`: 终端交互界面。
- `internal/web`: Web 服务与静态页面。
- `internal/httpx`: HTTP 客户端与 POST 模板解析。
