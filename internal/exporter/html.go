package exporter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/so-novel/sonovel-go/internal/model"
	"github.com/so-novel/sonovel-go/internal/util"
)

type HTMLExporter struct{}

func (e *HTMLExporter) Export(meta model.BookMeta, chapters []model.ChapterItem, outputDir string) (string, error) {
	dir := filepath.Join(outputDir, fmt.Sprintf("%s (%s) HTML", util.SanitizeName(meta.BookName), util.SanitizeName(meta.Author)))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	for _, ch := range chapters {
		name := fmt.Sprintf("%04d_%s.html", ch.Order, util.SanitizeName(ch.Title))
		path := filepath.Join(dir, name)
		content := renderChapterHTML(meta, ch)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return "", err
		}
	}
	return dir, nil
}

func renderChapterHTML(meta model.BookMeta, ch model.ChapterItem) string {
	return fmt.Sprintf(`<!doctype html>
<html lang="zh-CN"><head><meta charset="UTF-8"><title>%s</title></head>
<body><h1>%s</h1>%s</body></html>`,
		escapeHTML(ch.Title), escapeHTML(ch.Title), ch.Content)
}

func escapeHTML(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return r.Replace(s)
}
