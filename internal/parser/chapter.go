package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/so-novel/sonovel-go/internal/httpx"
	"github.com/so-novel/sonovel-go/internal/model"
	"github.com/so-novel/sonovel-go/internal/util"
)

type ChapterParser struct {
	Client *httpx.Client
	Rule   *model.Rule
}

func (p *ChapterParser) Parse(ch model.ChapterItem) (model.ChapterItem, error) {
	if p.Rule.Chapter == nil {
		return ch, fmt.Errorf("chapter rule is empty")
	}
	r := p.Rule.Chapter
	body, err := p.Client.Get(ch.URL, r.Timeout, "")
	if err != nil {
		return ch, err
	}
	doc, err := util.NewDocument(body, r.BaseURI)
	if err != nil {
		return ch, err
	}
	title := selectField(doc.Selection, r.Title, doc.Url)
	if title == "" {
		title = ch.Title
	}
	content := util.SelectHTML(doc.Selection, r.Content)
	if content == "" {
		return ch, fmt.Errorf("empty chapter content: %s", ch.URL)
	}
	content = cleanContent(content, r)
	ch.Title = title
	ch.Content = content
	return ch, nil
}

func cleanContent(html string, rule *model.Chapter) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader("<div id='root'>" + html + "</div>"))
	if err == nil {
		if strings.TrimSpace(rule.FilterTag) != "" {
			for _, s := range strings.Fields(rule.FilterTag) {
				doc.Find(s).Remove()
			}
		}
		h, _ := doc.Find("#root").Html()
		html = h
	}

	if strings.TrimSpace(rule.FilterTxt) != "" {
		if re, e := regexp.Compile(rule.FilterTxt); e == nil {
			html = re.ReplaceAllString(html, "")
		}
	}
	if !rule.ParagraphTagClosed && strings.TrimSpace(rule.ParagraphTag) != "" {
		re := regexp.MustCompile(rule.ParagraphTag)
		html = re.ReplaceAllString(html, "</p><p>")
	}
	html = strings.TrimSpace(html)
	if !strings.HasPrefix(html, "<p>") {
		html = "<p>" + html
	}
	if !strings.HasSuffix(html, "</p>") {
		html += "</p>"
	}
	return html
}
