package app

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/opso-code/sonovel-go/internal/crawler"
	"github.com/opso-code/sonovel-go/internal/exporter"
	"github.com/opso-code/sonovel-go/internal/httpx"
	"github.com/opso-code/sonovel-go/internal/model"
	"github.com/opso-code/sonovel-go/internal/parser"
	"github.com/opso-code/sonovel-go/internal/rule"
)

type Service struct {
	Cfg   model.Config
	Rules []model.Rule
}

func New(cfg model.Config) (*Service, error) {
	rules, err := rule.LoadRules(cfg.RulesFile)
	if err != nil {
		return nil, err
	}
	return &Service{Cfg: cfg, Rules: rules}, nil
}

func (s *Service) Search(keyword string) ([]model.SearchResult, error) {
	r, err := s.pickRuleForSearch()
	if err != nil {
		return nil, err
	}
	sp := &parser.SearchParser{Client: httpx.New(20 * time.Second), Rule: r, Cfg: s.Cfg}
	return sp.Parse(keyword)
}

func (s *Service) DownloadByURL(bookURL string) (string, error) {
	r, err := s.pickRuleForBookURL(bookURL)
	if err != nil {
		return "", err
	}
	client := httpx.New(25 * time.Second)
	c := &crawler.Crawler{
		Cfg:           s.Cfg,
		BookParser:    &parser.BookParser{Client: client, Rule: r},
		TocParser:     &parser.TocParser{Client: client, Rule: r},
		ChapterParser: &parser.ChapterParser{Client: client, Rule: r},
	}
	meta, chapters, err := c.Crawl(bookURL)
	if err != nil {
		return "", err
	}
	ex, err := pickExporter(s.Cfg.Format)
	if err != nil {
		return "", err
	}
	return ex.Export(meta, chapters, filepath.Clean(s.Cfg.OutputDir))
}

func (s *Service) pickRuleForSearch() (*model.Rule, error) {
	if s.Cfg.SourceID <= 0 {
		return nil, fmt.Errorf("source-id is required for search")
	}
	r, err := rule.GetRuleByID(s.Rules, s.Cfg.SourceID)
	if err != nil {
		return nil, err
	}
	if r.Search == nil || r.Search.Disabled {
		return nil, fmt.Errorf("source %d does not support search", r.ID)
	}
	if strings.HasPrefix(strings.TrimSpace(r.Search.Result), "/") {
		return nil, fmt.Errorf("source %d search uses xpath, not supported yet", r.ID)
	}
	return r, nil
}

func (s *Service) pickRuleForBookURL(bookURL string) (*model.Rule, error) {
	if s.Cfg.SourceID > 0 {
		return rule.GetRuleByID(s.Rules, s.Cfg.SourceID)
	}
	return rule.GetRuleByBookURL(s.Rules, bookURL)
}

func pickExporter(format string) (exporter.Exporter, error) {
	switch strings.ToLower(format) {
	case "txt":
		return &exporter.TXTExporter{}, nil
	case "html":
		return &exporter.HTMLExporter{}, nil
	case "epub", "":
		return &exporter.EPUBExporter{}, nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}
