# so-novel-go

用 Go 重新实现的 so-novel（原 Java 项目）核心功能。

## 功能特性

- ✅ **核心解析器** - Book/Chapter/Toc/Search 解析器
- ✅ **并发爬虫** - Worker pool 模式 + 进度回调
- ✅ **内容导出** - TXT/EPUB 格式导出
- ✅ **Web 服务** - REST API 服务
- ✅ **TUI/CLI 界面** - 终端界面支持
- ✅ **规则系统** - JSON 规则加载
- ✅ **XPath 支持** - CSS/XPath 选择器
- ✅ **JS 执行器** - goja 引擎支持 `@js:` 前缀
- ✅ **中文转换** - 繁简转换接口
- ✅ **封面更新** - 起点/纵横/七猫封面获取
- ✅ **本地书源** - 支持本地 HTML 文件导入
- ✅ **SSE 进度推送** - 实时推送下载进度
- ✅ **PDF 导出** - 支持 PDF 导出（需 htmltopdf）
- ✅ **书籍详情** - BookFetchServlet 实现
- ✅ **建议接口** - 书名补全/纠错
- ✅ **书籍删除** - 删除已下载书籍

## 安装

```bash
go install github.com/opso-code/sonovel-go@latest
```

## 使用示例

### 基本使用

```bash
sonovel -rules ./config.toml -output ./downloads -format txt
```

### 命令行参数

```bash
sonovel -rules ./config.toml \
  -output ./downloads \
  -format txt \
  -bookId 12345 \
  -chapterStart 1 \
  -chapterEnd 100 \
  -concurrency 4
```

### 规则配置

```toml
[[rules]]
id = 1
name = "起点中文网"
url = "https://www.qidian.com"
language = "zh-cn"
disabled = false

[[rules.book]]
url = "https://www.qidian.com/info/{bookId}.html"
title = ".book-info .book-title"
author = ".book-info .book-author"
introduction = ".book-info .book-intro"
coverUrl = ".book-info img.cover"
latestChapter = ".book-info .book-latest-chapter"
lastUpdateTime = ".book-info .book-latest-update-time"
status = ".book-info .book-status"

[[rules.toc]]
url = "https://www.qidian.com/info/{bookId}.html"
list = ".book-info .book-toc-list"
item = ".book-info .book-toc-item"
isDesc = false
pagination = true
nextPage = ".book-info .book-toc-next-page"

[[rules.chapter]]
url = "https://www.qidian.com/chapter/{chapterId}.html"
title = ".chapter-title"
content = ".chapter-content"
filterTxt = "filter.txt"
filterTag = ".filter-tag"

[[rules.crawl]]
concurrency = 4
minInterval = 1000
maxInterval = 5000
maxAttempts = 3
retryMinInterval = 5000
retryMaxInterval = 10000
```

### 规则文件

```json
{
  "id": 1,
  "name": "起点中文网",
  "url": "https://www.qidian.com",
  "language": "zh-cn",
  "disabled": false,
  "book": {
    "url": "https://www.qidian.com/info/{bookId}.html",
    "title": ".book-info .book-title"
  },
  "toc": {
    "url": "https://www.qidian.com/info/{bookId}.html",
    "list": ".book-info .book-toc-list"
  },
  "chapter": {
    "url": "https://www.qidian.com/chapter/{chapterId}.html",
    "title": ".chapter-title",
    "content": ".chapter-content"
  },
  "crawl": {
    "concurrency": 4,
    "minInterval": 1000,
    "maxInterval": 5000
  }
}
```

## 构建

```bash
go build -o sonovel ./cmd/sonovel
```

## 依赖

- goquery v1.10.2
- golang.org/x/net v0.40.0
- github.com/PuerkitoBio/goquery
- github.com/antchfx/cascadia

## 许可证

MIT License
