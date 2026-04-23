package exporter

import "github.com/so-novel/sonovel-go/internal/model"

type Exporter interface {
	Export(meta model.BookMeta, chapters []model.ChapterItem, outputDir string) (string, error)
}
