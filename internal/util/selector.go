package util

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
)

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
	s := sel.Find(q)
	if s.Length() == 0 {
		return ""
	}
	res := CleanSpaces(s.First().Text())
	return ApplyInlineJS(js, res)
}

func SelectHTML(sel *goquery.Selection, query string) string {
	q, js := SplitQueryAndJS(query)
	if q == "" {
		return ""
	}
	if strings.HasPrefix(q, "/") {
		n := strings.ToLower(strings.TrimSpace(q))
		switch n {
		case "/html", "//html", "(//html)[1]":
			h, err := sel.Html()
			if err != nil {
				return ""
			}
			return ApplyInlineJS(js, strings.TrimSpace(h))
		default:
			return ""
		}
	}
	s := sel.Find(q)
	if s.Length() == 0 {
		return ""
	}
	h, _ := s.First().Html()
	return ApplyInlineJS(js, strings.TrimSpace(h))
}

func SelectAttr(sel *goquery.Selection, query, attr string, base *url.URL) string {
	q, js := SplitQueryAndJS(query)
	if q == "" {
		return ""
	}
	s := sel.Find(q)
	if s.Length() == 0 {
		return ""
	}
	val, _ := s.First().Attr(attr)
	val = strings.TrimSpace(val)
	if (attr == "href" || attr == "src") && base != nil {
		u, err := url.Parse(val)
		if err == nil {
			val = base.ResolveReference(u).String()
		}
	}
	return ApplyInlineJS(js, val)
}

func NodeToSelection(node *goquery.Selection) *goquery.Selection {
	return node
}

func SelectList(sel *goquery.Selection, query string) (*goquery.Selection, error) {
	q, _ := SplitQueryAndJS(query)
	if q == "" {
		return nil, fmt.Errorf("empty query")
	}
	return sel.Find(q), nil
}

func ReadAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}
