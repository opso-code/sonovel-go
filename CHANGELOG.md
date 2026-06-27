# Changelog

All notable changes to this project will be documented in this file.

## [0.1.0] - 2024-06-27

### ✨ Features
- 支持 XPath 查询（移除所有"xpath not supported"错误）
- 添加 CoverUpdater 封面更新器（支持起点/纵横/七猫）
- 添加 BookFetcher 统一接口
- 添加本地书源支持
- 添加 JS 执行器（使用 goja）
- 添加中文转换器（繁简转换框架）
- 添加 PDF 导出器（文本替代方案）
- 添加 Web 服务扩展（SSE 进度推送、书籍详情、书名补全、删除）
- 添加 GitHub Actions 自动构建发布流程
- 二进制优化（CGO_ENABLED=0 + ldflags 优化）
- 统一二进制命名（go-sonovel）

### 🐛 Bug Fixes
- 修复 XPath 查询不支持的错误
- 修复选择器匹配错误
- 修复并发爬虫内存泄漏
- 修复 Web 服务 CORS 配置

### 📦 Dependencies
- goquery v1.10.2
- golang.org/x/net v0.40.0
- github.com/antchfx/xpath v1.3.1
- github.com/antchfx/cssquery v1.3.0
- github.com/gosimple/slug v1.13.0
- github.com/tdewolff/minify/v2 v2.12.9
- github.com/tdewolff/parse/v2 v2.6.5
- github.com/andybalholm/cascadia v1.3.2
- github.com/antchfx/htmlquery v1.3.5
- github.com/itchyny/gojq v0.12.15
- github.com/jaytaylor/html2text v0.0.0-20230321000545-44aa0b420618
- github.com/chromedp/cdproto v0.0.0-20240901102400-b9d66e514ad3
- github.com/chromedp/chromedp v0.9.2
- github.com/iancoleman/orderedmap v0.3.0

### ⚡️ Performance
- 二进制大小从 22MB 降至 16MB（静态编译）
- 并发爬虫 worker pool 优化
- 流式处理减少内存占用

### 📝 Documentation
- 添加 AGENTS.md AI 开发指南
- 完善 README.md 使用示例
- 添加 docs/release.md 发布流程文档
