package core

import (
	"github.com/opso-code/sonovel-go/internal/model"
)

type Source interface {
	FetchBook(bookID, bookName string) (*model.Book, error)
	FetchChapter(bookID, chapterID string) (*model.Chapter, error)
	FetchToc(bookID string) (*model.Toc, error)
	FetchSearch(keyword string) ([]*model.Book, error)
}
