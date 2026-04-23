# sonovel-go

Go 语言实现的 So Novel 核心功能版本（规则驱动抓取）。

## 已实现主要功能

- 规则加载：读取 `bundle/rules/*.json` 或指定规则文件。
- 搜索：按书源规则搜索书名/作者。
- 抓取：解析书籍详情、目录、章节正文。
- 并发下载：支持并发抓章节、失败重试、章节范围下载。
- 导出：`epub`、`txt`、`html`。

## 当前限制

- 优先支持 CSS 选择器规则。
- `xpath` 规则当前未实现（后续可补 `htmlquery` 方案）。
- `@js:` 目前实现了常见字符串处理（如 `replace`、前缀拼接），复杂 JS 逻辑暂不完整。
- 未实现 Java 版的 Web UI / TUI / PDF 导出。

## 快速开始

```bash
cd sonovel-go
GOPROXY=https://proxy.golang.org,direct go mod tidy
```

### 搜索

```bash
go run ./cmd/sonovel search --kw "斗罗大陆" --source-id 1 --rules ../bundle/rules/main.json
```

### 下载

```bash
go run ./cmd/sonovel download \
  --url "https://www.shuhaige.net/70475/" \
  --format epub \
  --out ./downloads \
  --rules ../bundle/rules/main.json \
  --concurrency 12
```

### 章节范围下载

```bash
go run ./cmd/sonovel download \
  --url "https://www.shuhaige.net/70475/" \
  --start 1 --end 100 \
  --format txt
```

## 目录结构

- `cmd/sonovel`: CLI 入口。
- `internal/rule`: 规则加载与默认值处理。
- `internal/parser`: 搜索/详情/目录/章节解析。
- `internal/crawler`: 并发抓取与重试调度。
- `internal/exporter`: `txt/html/epub` 导出。
- `internal/httpx`: HTTP 客户端与 POST 模板解析。

