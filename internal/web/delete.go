package web

import (
	"net/http"
	"os"
	"path/filepath"
)

type BookDeleteServlet struct {
	basePath string
}

func NewBookDeleteServlet(basePath string) *BookDeleteServlet {
	return &BookDeleteServlet{
		basePath: basePath,
	}
}

func (s *BookDeleteServlet) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bookID := r.URL.Query().Get("bookId")
	if bookID == "" {
		http.Error(w, "bookId is required", http.StatusBadRequest)
		return
	}

	bookPath := filepath.Join(s.basePath, bookID)
	if err := os.RemoveAll(bookPath); err != nil {
		http.Error(w, "delete failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("book deleted successfully"))
}
