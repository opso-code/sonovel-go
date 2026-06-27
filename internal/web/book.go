package web

import (
	"encoding/json"
	"net/http"

	"github.com/opso-code/sonovel-go/internal/core"
)

type BookFetchServlet struct {
	source core.Source
}

func NewBookFetchServlet(source core.Source) *BookFetchServlet {
	return &BookFetchServlet{
		source: source,
	}
}

func (s *BookFetchServlet) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bookID := r.URL.Query().Get("bookId")
	bookName := r.URL.Query().Get("bookName")

	if bookID == "" && bookName == "" {
		http.Error(w, "bookId or bookName is required", http.StatusBadRequest)
		return
	}

	book, err := s.source.FetchBook(bookID, bookName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}
