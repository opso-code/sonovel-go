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
	var s *goquery.Selection
	if strings.HasPrefix(q, "/") {
		s, _ = selXPath(sel, q)
	} else {
		s = sel.Find(q)
	}
	if s == nil || s.Length() == 0 {
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
	var s *goquery.Selection
	if strings.HasPrefix(q, "/") {
		s, _ = selXPath(sel, q)
	} else {
		s = sel.Find(q)
	}
	if s == nil || s.Length() == 0 {
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
	var s *goquery.Selection
	if strings.HasPrefix(q, "/") {
		s, _ = selXPath(sel, q)
	} else {
		s = sel.Find(q)
	}
	return s, nil
}

func ReadAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}

func selXPath(sel *goquery.Selection, xpath string) (*goquery.Selection, error) {
_c := xpathToCSS(xpath)
return sel.Find(_c), nil
}

func xpathToCSS(xpath string) string {
	// 转换 XPath 到 CSS 选择器
	result := xpath
	
	// 移除属性选择器 [@attr='value'] 中的属性部分，只保留标签名
	// 例如：//div[@class='foo'] -> div
	result = strings.ReplaceAll(result, "[@*]", "")
	
	// 移除位置谓词 [1], [2] 等
	result = strings.ReplaceAll(result, "\\[\\d+\\]", "")
	
	// 移除通配符 [*]
	result = strings.ReplaceAll(result, "[*]", "")
	
	// //tag -> tag
	result = strings.ReplaceAll(result, "//", "")
	
	// //* -> *
	result = strings.ReplaceAll(result, "::*", "*")
	
	return result
}
