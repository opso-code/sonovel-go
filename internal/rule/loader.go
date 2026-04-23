package rule

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/opso-code/sonovel-go/internal/model"
)

func LoadRules(path string) ([]model.Rule, error) {
	p := strings.TrimSpace(path)
	if p == "" {
		p = "bundle/rules/main.json"
	}
	info, err := os.Stat(p)
	if err != nil {
		return nil, err
	}

	var files []string
	if info.IsDir() {
		entries, err := os.ReadDir(p)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			files = append(files, filepath.Join(p, e.Name()))
		}
		sort.Strings(files)
	} else {
		files = []string{p}
	}

	var all []model.Rule
	for _, file := range files {
		b, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		var arr []model.Rule
		if err := json.Unmarshal(b, &arr); err != nil {
			return nil, fmt.Errorf("parse %s: %w", file, err)
		}
		all = append(all, arr...)
	}

	for i := range all {
		all[i].ID = i + 1
		applyDefaults(&all[i])
	}
	return all, nil
}

func GetRuleByID(rules []model.Rule, id int) (*model.Rule, error) {
	for i := range rules {
		if rules[i].ID == id {
			return &rules[i], nil
		}
	}
	return nil, fmt.Errorf("rule id=%d not found", id)
}

func GetRuleByBookURL(rules []model.Rule, bookURL string) (*model.Rule, error) {
	for i := range rules {
		if strings.HasPrefix(bookURL, rules[i].URL) {
			return &rules[i], nil
		}
	}
	return nil, fmt.Errorf("no rule matched for %s", bookURL)
}

func applyDefaults(r *model.Rule) {
	if r.Search != nil {
		if r.Search.BaseURI == "" {
			r.Search.BaseURI = r.URL
		}
		if r.Search.Timeout == 0 {
			r.Search.Timeout = 15
		}
	}
	if r.Book != nil {
		if r.Book.BaseURI == "" {
			r.Book.BaseURI = r.URL
		}
		if r.Book.Timeout == 0 {
			r.Book.Timeout = 15
		}
		if r.Book.BookName == "" {
			r.Book.BookName = `meta[property="og:novel:book_name"]`
		}
		if r.Book.Author == "" {
			r.Book.Author = `meta[property="og:novel:author"]`
		}
		if r.Book.Intro == "" {
			r.Book.Intro = `meta[name="description"]`
		}
		if r.Book.CoverURL == "" {
			r.Book.CoverURL = `meta[property="og:image"]`
		}
		if r.Book.Category == "" {
			r.Book.Category = `meta[property="og:novel:category"]`
		}
		if r.Book.LatestChapter == "" {
			r.Book.LatestChapter = `meta[property="og:novel:latest_chapter_name"]`
		}
		if r.Book.LastUpdateTime == "" {
			r.Book.LastUpdateTime = `meta[property="og:novel:update_time"]`
		}
		if r.Book.Status == "" {
			r.Book.Status = `meta[property="og:novel:status"]`
		}
	}
	if r.Toc != nil {
		if r.Toc.BaseURI == "" {
			r.Toc.BaseURI = r.URL
		}
		if r.Toc.Timeout == 0 {
			r.Toc.Timeout = 60
		}
	}
	if r.Chapter != nil {
		if r.Chapter.BaseURI == "" {
			r.Chapter.BaseURI = r.URL
		}
		if r.Chapter.Timeout == 0 {
			r.Chapter.Timeout = 15
		}
	}
}
