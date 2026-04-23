package parser

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/opso-code/sonovel-go/internal/httpx"
	"github.com/opso-code/sonovel-go/internal/model"
	"github.com/opso-code/sonovel-go/internal/util"
)

type TocParser struct {
	Client *httpx.Client
	Rule   *model.Rule
}

func (p *TocParser) ParseAll(bookURL string) ([]model.ChapterItem, error) {
	if p.Rule.Toc == nil {
		return nil, fmt.Errorf("toc rule is empty")
	}
	r := p.Rule.Toc
	targetURL := bookURL
	if r.URL != "" {
		id := extractBookID(p.Rule, bookURL)
		targetURL = fmt.Sprintf(r.URL, id)
	}
	body, err := p.Client.Get(targetURL, r.Timeout, "")
	if err != nil {
		return nil, err
	}
	doc, err := util.NewDocument(body, r.BaseURI)
	if err != nil {
		return nil, err
	}

	listSel := doc.Selection
	if r.List != "" {
		html := util.SelectHTML(doc.Selection, r.List)
		if html != "" {
			subDoc, e := goquery.NewDocumentFromReader(strings.NewReader(html))
			if e == nil {
				listSel = subDoc.Selection
			}
		}
	}

	q, _ := util.SplitQueryAndJS(r.Item)
	if strings.HasPrefix(q, "/") {
		return nil, fmt.Errorf("xpath not supported in toc.item: %s", r.Item)
	}

	items := make([]model.ChapterItem, 0)
	order := 1
	listSel.Find(q).Each(func(_ int, it *goquery.Selection) {
		title := strings.TrimSpace(it.Text())
		href, _ := it.Attr("href")
		href = strings.TrimSpace(href)
		if title == "" || href == "" {
			return
		}
		if doc.Url != nil {
			u, err := url.Parse(href)
			if err == nil {
				href = doc.Url.ResolveReference(u).String()
			}
		}
		items = append(items, model.ChapterItem{Order: order, Title: title, URL: href})
		order++
	})

	if r.Desc {
		reverseChapters(items)
	}
	return items, nil
}

func reverseChapters(in []model.ChapterItem) {
	for i, j := 0, len(in)-1; i < j; i, j = i+1, j-1 {
		in[i], in[j] = in[j], in[i]
	}
	for i := range in {
		in[i].Order = i + 1
	}
}

func extractBookID(rule *model.Rule, bookURL string) string {
	if rule.Book == nil || rule.Book.URL == "" {
		return ""
	}
	pattern, _ := util.SplitQueryAndJS(rule.Book.URL)
	re, err := regexp.Compile(pattern)
	if err != nil {
		return ""
	}
	m := re.FindStringSubmatch(bookURL)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}
