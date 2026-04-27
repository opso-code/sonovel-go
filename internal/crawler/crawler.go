package crawler

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/opso-code/sonovel-go/internal/model"
	"github.com/opso-code/sonovel-go/internal/parser"
)

type Crawler struct {
	Cfg           model.Config
	BookParser    *parser.BookParser
	TocParser     *parser.TocParser
	ChapterParser *parser.ChapterParser
}

var ErrCanceled = errors.New("download canceled")

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
	jobs := make(chan job, len(toc))
	results := make([]model.ChapterItem, len(toc))
	var wg sync.WaitGroup
	var cbMu sync.Mutex
	var completed int32
	var firstErr error
	var errMu sync.Mutex
	total := len(toc)
	if c.Cfg.OnProgress != nil {
		c.Cfg.OnProgress(0, total)
	}

	worker := func() {
		defer wg.Done()
		for j := range jobs {
			if c.isCanceled() {
				errMu.Lock()
				if firstErr == nil {
					firstErr = ErrCanceled
				}
				errMu.Unlock()
				continue
			}
			ch := toc[j.idx]
			if c.Cfg.OnChapter != nil {
				done := int(atomic.LoadInt32(&completed))
				c.Cfg.OnChapter(done, total, ch.Title)
			}
			item, err := c.fetchWithRetry(ch)
			if err != nil {
				if errors.Is(err, ErrCanceled) {
					errMu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					errMu.Unlock()
				} else {
					item = failedChapterPlaceholder(ch, err)
				}
			}
			if item.Title == "" {
				item.Title = ch.Title
			}
			if item.URL == "" {
				item.URL = ch.URL
			}
			if strings.TrimSpace(item.Content) == "" {
				item = failedChapterPlaceholder(ch, fmt.Errorf("empty content"))
			}
			results[j.idx] = item

			if c.Cfg.OnProgress != nil {
				cur := int(atomic.AddInt32(&completed, 1))
				cbMu.Lock()
				c.Cfg.OnProgress(cur, total)
				cbMu.Unlock()
			}
		}
	}

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go worker()
	}
	for i := range toc {
		if c.isCanceled() {
			break
		}
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
		if c.isCanceled() {
			return ch, ErrCanceled
		}
		if err := sleepRandom(c.Cfg.MinIntervalMS, c.Cfg.MaxIntervalMS, c.Cfg.ShouldCancel); err != nil {
			return ch, err
		}
		item, err := c.ChapterParser.Parse(ch)
		if err == nil {
			return item, nil
		}
		lastErr = err
		if i < maxTry {
			if err := sleepRandom(c.Cfg.RetryMinMS*i, c.Cfg.RetryMaxMS*i, c.Cfg.ShouldCancel); err != nil {
				return ch, err
			}
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

func sleepRandom(minMS, maxMS int, shouldCancel func() bool) error {
	if minMS <= 0 {
		minMS = 100
	}
	if maxMS <= minMS {
		maxMS = minMS + 1
	}
	d := rand.IntN(maxMS-minMS) + minMS
	deadline := time.Now().Add(time.Duration(d) * time.Millisecond)
	for {
		if shouldCancel != nil && shouldCancel() {
			return ErrCanceled
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil
		}
		if remaining > 150*time.Millisecond {
			remaining = 150 * time.Millisecond
		}
		time.Sleep(remaining)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (c *Crawler) isCanceled() bool {
	return c.Cfg.ShouldCancel != nil && c.Cfg.ShouldCancel()
}

func failedChapterPlaceholder(ch model.ChapterItem, err error) model.ChapterItem {
	msg := "unknown error"
	if err != nil {
		msg = err.Error()
	}
	return model.ChapterItem{
		Order: ch.Order,
		Title: ch.Title,
		URL:   ch.URL,
		Content: fmt.Sprintf(
			"<p>[章节下载失败]</p><p>章节：%s</p><p>原因：%s</p><p>链接：%s</p>",
			escapeHTML(ch.Title), escapeHTML(msg), escapeHTML(ch.URL),
		),
	}
}

func escapeHTML(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return r.Replace(s)
}
