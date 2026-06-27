package web

import (
	"encoding/json"
	"net/http"
)

type BookSuggestServlet struct{}

func NewBookSuggestServlet() *BookSuggestServlet {
	return &BookSuggestServlet{}
}

func (s *BookSuggestServlet) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	keyword := r.URL.Query().Get("keyword")
	if keyword == "" {
		http.Error(w, "keyword is required", http.StatusBadRequest)
		return
	}

	// 实际的补全逻辑应该使用搜索引擎 API
	// 这里使用简单的实现
	suggestions := []string{keyword}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suggestions)
}
