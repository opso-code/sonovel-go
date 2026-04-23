package crawler

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/so-novel/sonovel-go/internal/model"
	"github.com/so-novel/sonovel-go/internal/parser"
)

type Crawler struct {
	Cfg           model.Config
	BookParser    *parser.BookParser
	TocParser     *parser.TocParser
	ChapterParser *parser.ChapterParser
}

func (c *Crawler) Crawl(bookURL string) (model.BookMeta, []model.ChapterItem, error) {
	meta, err := c.BookParser.Parse(bookURL)
	if err != nil {
		return model.BookMeta{}, nil, err
	}
	toc, err := c.TocParser.ParseAll(bookURL)
	if err != nil {
		return model.BookMeta{}, nil, err
	}
	if len(toc) == 0 {
		return meta, nil, fmt.Errorf("toc empty")
	}

	toc = applyRange(toc, c.Cfg.ChapterStart, c.Cfg.ChapterEnd)
	chapters, err := c.fetchChapters(toc)
	if err != nil {
		return meta, nil, err
	}
	return meta, chapters, nil
}

func (c *Crawler) fetchChapters(toc []model.ChapterItem) ([]model.ChapterItem, error) {
	workers := c.Cfg.Concurrency
	if workers <= 0 {
		workers = min(20, len(toc))
	}
	type job struct{ idx int }
	jobs := make(chan job)
	results := make([]model.ChapterItem, len(toc))
	var wg sync.WaitGroup
	var firstErr error
	var errMu sync.Mutex

	worker := func() {
		defer wg.Done()
		for j := range jobs {
			ch := toc[j.idx]
			item, err := c.fetchWithRetry(ch)
			if err != nil {
				errMu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				errMu.Unlock()
				continue
			}
			results[j.idx] = item
		}
	}

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go worker()
	}
	for i := range toc {
		jobs <- job{idx: i}
	}
	close(jobs)
	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}
	return results, nil
}

func (c *Crawler) fetchWithRetry(ch model.ChapterItem) (model.ChapterItem, error) {
	maxTry := 1
	if c.Cfg.EnableRetry {
		if c.Cfg.MaxRetries > 0 {
			maxTry = c.Cfg.MaxRetries
		} else {
			maxTry = 3
		}
	}
	var lastErr error
	for i := 1; i <= maxTry; i++ {
		sleepRandom(c.Cfg.MinIntervalMS, c.Cfg.MaxIntervalMS)
		item, err := c.ChapterParser.Parse(ch)
		if err == nil {
			return item, nil
		}
		lastErr = err
		if i < maxTry {
			sleepRandom(c.Cfg.RetryMinMS*i, c.Cfg.RetryMaxMS*i)
		}
	}
	return ch, fmt.Errorf("fetch chapter failed: %s (%w)", ch.Title, lastErr)
}

func applyRange(in []model.ChapterItem, start, end int) []model.ChapterItem {
	if start <= 0 {
		start = 1
	}
	if end <= 0 || end > len(in) {
		end = len(in)
	}
	if start > end {
		start = 1
		end = len(in)
	}
	out := append([]model.ChapterItem{}, in[start-1:end]...)
	for i := range out {
		out[i].Order = i + 1
	}
	return out
}

func sleepRandom(minMS, maxMS int) {
	if minMS <= 0 {
		minMS = 100
	}
	if maxMS <= minMS {
		maxMS = minMS + 1
	}
	d := rand.IntN(maxMS-minMS) + minMS
	time.Sleep(time.Duration(d) * time.Millisecond)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
