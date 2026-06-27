# AGENTS.md

## 1. 项目概述

so-novel-go 是 so-novel 的 Go 重写版，支持书源解析、并发爬虫、TXT/EPUB/PDF 导出。核心架构：Source(书源定义) → Parser(解析器) → Crawler(并发爬虫) → Exporter(导出器)。技术栈：Go 1.24 无 CGO，依赖 goquery/cascadia/xpath/cssquery。

## 2. 快速命令

```bash
make build          # 构建当前平台
make build-all      # 构建所有平台
make test           # 运行测试
make verify         # 验证构建和测试
make clean          # 清理
```

环境变量：无特殊配置，`CGO_ENABLED=0` 默认禁用。

## 3. 后端架构

```
cmd/sonovel/        # 命令行入口
internal/
  core/             # 核心功能：封面更新、本地书源、BookFetcher
  crawler/          # 并发爬虫
  exporter/         # 导出器：TXT/EPUB/PDF
  model/            # 数据模型
  parser/           # 解析器：Book/Chapter/Toc/Search
  rule/             # 规则加载
  source/           # 书源定义
  util/             # 工具：选择器、XPath、JS 执行、中文转换、文本处理
  web/              # Web 服务
rules/              # 规则文件
docs/               # 文档
```

核心子系统：
- **Source**: 书源定义与初始化
- **Parser**: 4 大解析器（Book/Chapter/Toc/Search）
- **Crawler**: 并发爬虫 + 进度回调
- **Exporter**: TXT/EPUB/PDF 导出
- **Rule**: JSON 规则驱动解析

详细文档：[README.md](README.md)

## 5. 关键约定

1. 规则驱动解析：JSON 规则文件配置 CSS/XPath 选择器
2. 并发爬虫：worker pool 模式 + 进度回调
3. 内容清洗：正则过滤 + HTML 标签移除
4. 无外部 HTML 解析器：仅 goquery/cascadia
5. 使用 goja 执行 JS（替代 V8Runtime）
6. PDF 导出使用文本文件替代（需 htmltopdf）
7. 二进制统一命名为 go-sonovel
8. 使用语义化版本号（vX.Y.Z）
9. 禁止自动 push，等待明确指令
10. 禁止提交 secrets/keys

详细文档：[AGENTS.md](AGENTS.md)

## 7. 质量检查

| 命令 | 说明 |
|------|------|
| `go mod download` | 下载依赖 |
| `go test ./...` | 运行测试 |
| `go build ./...` | 构建 |
| `make verify` | 完整验证 |

## 8. 参考项目约定

- [so-novel](https://github.com/so-novel/so-novel) - 原版 Java 项目
- [goquery](https://github.com/PuerkitoBio/goquery) - HTML 解析
- [chromedp](https://github.com/chromedp/chromedp) - Chrome 自动化

优先级：so-novel > goquery > chromedp

## 9. 文档导航

| 文档 | 说明 |
|------|------|
| [README.md](README.md) | 项目概述和使用示例 |
| [config.toml](config.toml) | 运行时配置 |
| [rules/main.json](rules/main.json) | 主解析规则 |
| [Makefile](Makefile) | 构建脚本 |
| [.github/workflows/build.yml](.github/workflows/build.yml) | GitHub Actions |
| [.github/release.yml](.github/release.yml) | Release 变更日志 |
| [docs/release.md](docs/release.md) | 发布流程 |
