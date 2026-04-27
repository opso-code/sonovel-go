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

type ChapterParser struct {
	Client *httpx.Client
	Rule   *model.Rule
}

func (p *ChapterParser) Parse(ch model.ChapterItem) (model.ChapterItem, error) {
	if p.Rule.Chapter == nil {
		return ch, fmt.Errorf("chapter rule is empty")
	}
	r := p.Rule.Chapter
	title, content, err := p.fetchContent(ch.URL, r)
	if err != nil {
		return ch, err
	}
	if title == "" {
		title = ch.Title
	}
	content = cleanContent(content, r)
	if !hasVisibleText(content) {
		return ch, fmt.Errorf("chapter content has no visible text: %s", ch.URL)
	}
	ch.Title = title
	ch.Content = content
	return ch, nil
}

func (p *ChapterParser) fetchContent(startURL string, r *model.Chapter) (string, string, error) {
	if !r.Pagination {
		doc, err := p.fetchDocument(startURL, r)
		if err != nil {
			return "", "", err
		}
		return p.extractFromPage(doc, startURL, r)
	}

	nextURL := startURL
	visited := map[string]struct{}{}
	var title string
	var contentBuilder strings.Builder

	for i := 0; i < 200; i++ {
		if _, ok := visited[nextURL]; ok {
			break
		}
		visited[nextURL] = struct{}{}

		doc, err := p.fetchDocument(nextURL, r)
		if err != nil {
			return "", "", err
		}

		pageTitle, pageContent, err := p.extractFromPage(doc, nextURL, r)
		if err != nil {
			return "", "", err
		}
		if title == "" {
			title = pageTitle
		}
		contentBuilder.WriteString(pageContent)

		candidate := resolveNextPageURL(doc, r)
		if !shouldFollowNextPage(candidate, doc, r) {
			break
		}
		nextURL = candidate
	}

	content := strings.TrimSpace(contentBuilder.String())
	if content == "" {
		return "", "", fmt.Errorf("empty chapter content: %s", startURL)
	}
	return title, content, nil
}

func (p *ChapterParser) fetchDocument(pageURL string, r *model.Chapter) (*goquery.Document, error) {
	body, err := p.Client.Get(pageURL, r.Timeout, "")
	if err != nil {
		return nil, err
	}
	doc, err := util.NewDocument(body, r.BaseURI)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func (p *ChapterParser) extractFromPage(doc *goquery.Document, pageURL string, r *model.Chapter) (string, string, error) {
	title := selectField(doc.Selection, r.Title, doc.Url)
	content := util.SelectHTML(doc.Selection, r.Content)
	if strings.TrimSpace(content) == "" {
		return "", "", fmt.Errorf("empty chapter content: %s", pageURL)
	}
	return title, content, nil
}

func resolveNextPageURL(doc *goquery.Document, r *model.Chapter) string {
	if strings.TrimSpace(r.NextPage) == "" {
		return ""
	}

	candidate := util.SelectAttr(doc.Selection, r.NextPage, "href", doc.Url)
	if candidate == "" {
		candidate = util.SelectAttr(doc.Selection, r.NextPage, "value", doc.Url)
	}
	if candidate == "" {
		candidate = util.SelectText(doc.Selection, r.NextPage)
	}
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return ""
	}

	if doc.Url != nil {
		if u, err := url.Parse(candidate); err == nil {
			candidate = doc.Url.ResolveReference(u).String()
		}
	}
	return candidate
}

func shouldFollowNextPage(nextURL string, doc *goquery.Document, r *model.Chapter) bool {
	if nextURL == "" {
		return false
	}
	if strings.TrimSpace(r.NextChapterLink) != "" {
		if re, err := regexp.Compile(r.NextChapterLink); err == nil && re.MatchString(nextURL) {
			return false
		}
	}

	btnText := util.SelectText(doc.Selection, r.NextPage)
	urlLikePage := regexp.MustCompile(`.*[-_]\d\.html`).MatchString(nextURL)
	genericChapterEnd := regexp.MustCompile(`(下一章|没有了|>>|书末页)`).MatchString(btnText)
	return urlLikePage || !genericChapterEnd
}

func hasVisibleText(html string) bool {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader("<div id='root'>" + html + "</div>"))
	if err != nil {
		return strings.TrimSpace(html) != ""
	}
	return strings.TrimSpace(doc.Find("#root").Text()) != ""
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
