package web

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type SSEClient struct {
	id       int
	eventChan chan struct{}
	ctx      context.Context
}

type SSEServer struct {
	calls     int
	mu        sync.Mutex
	callsChan chan int

	clients   map[int]*SSEClient
	clientsMu sync.Mutex
}

func NewSSEServer() *SSEServer {
	return &SSEServer{
		calls:     0,
		callsChan: make(chan int, 100),
		clients:   make(map[int]*SSEClient),
	}
}

func (s *SSEServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	accept := r.Header.Get("Accept")
	if accept != "text/event-stream" {
		http.Error(w, "only SSE supported", http.StatusNotAcceptable)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Accept")

	clientID := time.Now().UnixNano()
	client := &SSEClient{
		id:       int(clientID),
		eventChan: make(chan struct{}, 1),
		ctx:      context.Background(),
	}

	s.clientsMu.Lock()
	s.clients[int(clientID)] = client
	s.clientsMu.Unlock()

	fmt.Fprintf(w, "event: connected\ndata: {\"id\": %d}\n\n", int(clientID))

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-client.ctx.Done():
			s.clientsMu.Lock()
			delete(s.clients, int(clientID))
			s.clientsMu.Unlock()
			return
		case <-ticker.C:
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

func (s *SSEServer) NotifyProgress(done, total int, title string) {
	s.mu.Lock()
	calls := s.calls
	s.calls = calls + 1
	s.mu.Unlock()

	s.mu.Lock()
	select {
	case s.callsChan <- calls:
	default:
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	for _, client := range s.clients {
		select {
		case client.eventChan <- struct{}{}:
		default:
		}
	}
}

func (s *SSEServer) GetProgress() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}
