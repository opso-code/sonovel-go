package util

import (
	"bytes"
	"io"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
)

func CleanSpaces(s string) string {
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = strings.TrimSpace(s)
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	return s
}

func SplitQueryAndJS(query string) (string, string) {
	parts := strings.SplitN(query, "@js:", 2)
	if len(parts) == 1 {
		return strings.TrimSpace(query), ""
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

func ApplyInlineJS(jsCode, input string) string {
	if strings.TrimSpace(jsCode) == "" {
		return input
	}
	if out, err := RunInlineJS(jsCode, input); err == nil {
		return strings.TrimSpace(out)
	}
	res := input

	if strings.Contains(jsCode, `r='`) && strings.Contains(jsCode, `'+r`) {
		start := strings.Index(jsCode, `r='`)
		end := strings.Index(jsCode[start+3:], `'`)
		if end > -1 {
			prefix := jsCode[start+3 : start+3+end]
			res = prefix + res
		}
	}

	if strings.Contains(jsCode, "replaceAll(") || strings.Contains(jsCode, "replace(") {
		re := regexp.MustCompile(`replace(All)?\(\s*['"](.*?)['"]\s*,\s*['"](.*?)['"]\s*\)`)
		for _, m := range re.FindAllStringSubmatch(jsCode, -1) {
			oldV := m[2]
			newV := m[3]
			res = strings.ReplaceAll(res, oldV, newV)
		}
	}

	return strings.TrimSpace(res)
}

func CleanHtmlTag(html string) string {
	re := regexp.MustCompile(`(?i)<[^>]*>`)
	return re.ReplaceAllString(html, "")
}

func NewDocument(raw []byte, baseURI string) (*goquery.Document, error) {
	reader, err := charset.NewReader(bytes.NewReader(raw), detectContentType(raw))
	if err != nil {
		reader = bytes.NewReader(raw)
	}
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, err
	}
	if baseURI != "" {
		doc.Url, _ = url.Parse(baseURI)
	}
	return doc, nil
}

func detectContentType(raw []byte) string {
	head := strings.ToLower(string(raw[:min(2048, len(raw))]))
	if strings.Contains(head, "charset=gbk") || strings.Contains(head, "charset=gb2312") || strings.Contains(head, "charset=gb18030") {
		return "text/html; charset=gbk"
	}
	return "text/html; charset=utf-8"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func SelectText(sel *goquery.Selection, query string) string {
	q, js := SplitQueryAndJS(query)
	if q == "" {
		return ""
	}
	
	var s *goquery.Selection
	if strings.HasPrefix(q, "/") {
		s, _ = selXPath(sel, q)
	} else {
		s = sel.Find(q)
	}
	
	if s == nil || s.Length() == 0 {
		return ""
	}
	
	text := s.First().Text()
	return ApplyInlineJS(js, text)
}

func SelectHTML(sel *goquery.Selection, query string) string {
	q, js := SplitQueryAndJS(query)
	if q == "" {
		return ""
	}
	
	var s *goquery.Selection
	if strings.HasPrefix(q, "/") {
		s, _ = selXPath(sel, q)
	} else {
		s = sel.Find(q)
	}
	
	if s == nil || s.Length() == 0 {
		return ""
	}
	
	html, _ := s.First().Html()
	return ApplyInlineJS(js, html)
}

func SelectAttr(sel *goquery.Selection, query, attr string, base *url.URL) string {
	q, js := SplitQueryAndJS(query)
	if q == "" {
		return ""
	}
	
	var s *goquery.Selection
	if strings.HasPrefix(q, "/") {
		s, _ = selXPath(sel, q)
	} else {
		s = sel.Find(q)
	}
	
	if s == nil || s.Length() == 0 {
		return ""
	}
	
	val, _ := s.First().Attr(attr)
	if val == "" {
		return ""
	}
	
	val = strings.TrimSpace(val)
	
	if (attr == "href" || attr == "src") && base != nil && val != "" {
		u, err := url.Parse(val)
		if err == nil && u.Scheme == "" && u.Host == "" {
			val = base.ResolveReference(u).String()
		}
	}
	
	return ApplyInlineJS(js, val)
}

func ReadAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}

func selXPath(sel *goquery.Selection, xpath string) (*goquery.Selection, error) {
	return sel.Find(xpath), nil
}
