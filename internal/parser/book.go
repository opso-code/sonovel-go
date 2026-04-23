package parser

import (
	"fmt"

	"github.com/opso-code/sonovel-go/internal/httpx"
	"github.com/opso-code/sonovel-go/internal/model"
	"github.com/opso-code/sonovel-go/internal/util"
)

type BookParser struct {
	Client *httpx.Client
	Rule   *model.Rule
}

func (p *BookParser) Parse(bookURL string) (model.BookMeta, error) {
	if p.Rule.Book == nil {
		return model.BookMeta{}, fmt.Errorf("book rule is empty")
	}
	r := p.Rule.Book
	body, err := p.Client.Get(bookURL, r.Timeout, "")
	if err != nil {
		return model.BookMeta{}, err
	}
	doc, err := util.NewDocument(body, r.BaseURI)
	if err != nil {
		return model.BookMeta{}, err
	}
	root := doc.Selection

	meta := model.BookMeta{
		URL:            bookURL,
		BookName:       selectField(root, r.BookName, doc.Url),
		Author:         selectField(root, r.Author, doc.Url),
		Intro:          selectField(root, r.Intro, doc.Url),
		Category:       selectField(root, r.Category, doc.Url),
		CoverURL:       selectField(root, r.CoverURL, doc.Url),
		LatestChapter:  selectField(root, r.LatestChapter, doc.Url),
		LastUpdateTime: selectField(root, r.LastUpdateTime, doc.Url),
		Status:         selectField(root, r.Status, doc.Url),
	}
	if meta.BookName == "" || meta.Author == "" {
		return meta, fmt.Errorf("book name or author empty")
	}
	return meta, nil
}
