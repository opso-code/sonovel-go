package parser

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/so-novel/sonovel-go/internal/util"
)

func selectField(sel *goquery.Selection, query string, base *url.URL) string {
	q, _ := util.SplitQueryAndJS(query)
	if q == "" {
		return ""
	}
	if strings.HasPrefix(q, "/") {
		return ""
	}
	if strings.HasPrefix(strings.TrimSpace(q), "meta[") {
		return util.SelectAttr(sel, query, "content", base)
	}
	if strings.Contains(strings.ToLower(q), "img") {
		if v := util.SelectAttr(sel, query, "src", base); v != "" {
			return v
		}
	}
	return util.SelectText(sel, query)
}
