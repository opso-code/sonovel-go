package parser

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/so-novel/sonovel-go/internal/httpx"
	"github.com/so-novel/sonovel-go/internal/model"
	"github.com/so-novel/sonovel-go/internal/util"
)

type SearchParser struct {
	Client *httpx.Client
	Rule   *model.Rule
	Cfg    model.Config
}

func (p *SearchParser) Parse(keyword string) ([]model.SearchResult, error) {
	if p.Rule.Search == nil || p.Rule.Search.Disabled {
		return nil, fmt.Errorf("source %d does not support search", p.Rule.ID)
	}
	r := p.Rule.Search
	searchURL := fmt.Sprintf(r.URL, url.QueryEscape(keyword))

	var body []byte
	var err error
	if strings.EqualFold(r.Method, "post") {
		body, err = p.Client.PostForm(searchURL, r.Data, []string{keyword}, r.Timeout, r.Cookies)
	} else {
		body, err = p.Client.Get(searchURL, r.Timeout, r.Cookies)
	}
	if err != nil {
		return nil, err
	}

	doc, err := util.NewDocument(body, r.BaseURI)
	if err != nil {
		return nil, err
	}
	return p.extractFromDoc(doc), nil
}

func (p *SearchParser) extractFromDoc(doc *goquery.Document) []model.SearchResult {
	r := p.Rule.Search
	res := make([]model.SearchResult, 0)
	query, _ := util.SplitQueryAndJS(r.Result)
	if strings.HasPrefix(query, "/") {
		return res
	}

	doc.Find(query).Each(func(_ int, item *goquery.Selection) {
		bookURL := util.SelectAttr(item, r.BookName, "href", doc.Url)
		bookName := util.SelectText(item, r.BookName)
		if bookName == "" || bookURL == "" {
			return
		}
		sr := model.SearchResult{
			SourceID:       p.Rule.ID,
			SourceName:     p.Rule.Name,
			URL:            bookURL,
			BookName:       bookName,
			Author:         util.SelectText(item, r.Author),
			Category:       util.SelectText(item, r.Category),
			LatestChapter:  util.SelectText(item, r.LatestChapter),
			LastUpdateTime: util.SelectText(item, r.LastUpdateTime),
			Status:         util.SelectText(item, r.Status),
			WordCount:      util.SelectText(item, r.WordCount),
		}
		res = append(res, sr)
	})

	if p.Cfg.SearchLimit > 0 && len(res) > p.Cfg.SearchLimit {
		return res[:p.Cfg.SearchLimit]
	}
	return res
}
