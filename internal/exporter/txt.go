package exporter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/so-novel/sonovel-go/internal/model"
	"github.com/so-novel/sonovel-go/internal/util"
)

type TXTExporter struct{}

func (e *TXTExporter) Export(meta model.BookMeta, chapters []model.ChapterItem, outputDir string) (string, error) {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", err
	}
	filename := fmt.Sprintf("%s(%s).txt", util.SanitizeName(meta.BookName), util.SanitizeName(meta.Author))
	path := filepath.Join(outputDir, filename)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("书名：%s\n作者：%s\n简介：%s\n\n", meta.BookName, meta.Author, meta.Intro))
	for _, ch := range chapters {
		sb.WriteString(ch.Title)
		sb.WriteString("\n\n")
		sb.WriteString(htmlToText(ch.Content))
		sb.WriteString("\n\n")
	}
	if err := os.WriteFile(path, []byte(sb.String()), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func htmlToText(html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader("<div>" + html + "</div>"))
	if err != nil {
		return html
	}
	text := doc.Find("div").Text()
	return util.CleanSpaces(strings.ReplaceAll(text, "\u3000", "  "))
}
