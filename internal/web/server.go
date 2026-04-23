package web

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/opso-code/sonovel-go/internal/app"
	"github.com/opso-code/sonovel-go/internal/model"
)

//go:embed static/*
var staticFiles embed.FS

type Server struct {
	BaseCfg    model.Config
	mux        *http.ServeMux
	downloadMu sync.Mutex
	jobsMu     sync.RWMutex
	jobs       map[string]*jobStatus
	jobSeq     uint64
}

type response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type sourceItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type downloadReq struct {
	URL         string `json:"url"`
	SourceID    int    `json:"sourceId"`
	Format      string `json:"format"`
	Concurrency int    `json:"concurrency"`
	Start       int    `json:"start"`
	End         int    `json:"end"`
}

type localFile struct {
	Name    string `json:"name"`
	IsDir   bool   `json:"isDir"`
	Size    int64  `json:"size"`
	ModTime int64  `json:"modTime"`
}

type jobStatus struct {
	ID        string `json:"id"`
	State     string `json:"state"`
	Message   string `json:"message"`
	Path      string `json:"path,omitempty"`
	StartedAt int64  `json:"startedAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

func New(base model.Config) *Server {
	s := &Server{BaseCfg: base, mux: http.NewServeMux(), jobs: map[string]*jobStatus{}}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler { return s.mux }

func (s *Server) routes() {
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/api/sources", s.handleSources)
	s.mux.HandleFunc("/api/search", s.handleSearch)
	s.mux.HandleFunc("/api/download", s.handleDownload)
	s.mux.HandleFunc("/api/job", s.handleJob)
	s.mux.HandleFunc("/api/files", s.handleFiles)
	s.mux.HandleFunc("/api/file", s.handleFile)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	b, err := fs.ReadFile(sub, "index.html")
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	_, _ = w.Write(b)
}

func (s *Server) handleSources(w http.ResponseWriter, _ *http.Request) {
	svc, err := app.New(s.BaseCfg)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	out := make([]sourceItem, 0, len(svc.Rules))
	for _, r := range svc.Rules {
		if r.Disabled {
			continue
		}
		out = append(out, sourceItem{ID: r.ID, Name: r.Name})
	}
	writeOK(w, out)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	kw := strings.TrimSpace(r.URL.Query().Get("kw"))
	sid, _ := strconv.Atoi(r.URL.Query().Get("sourceId"))
	if kw == "" || sid <= 0 {
		writeErr(w, http.StatusBadRequest, errors.New("kw and sourceId are required"))
		return
	}
	cfg := s.BaseCfg
	cfg.SourceID = sid
	svc, err := app.New(cfg)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	items, err := svc.Search(kw)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	writeOK(w, items)
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	var req downloadReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	if strings.TrimSpace(req.URL) == "" {
		writeErr(w, http.StatusBadRequest, errors.New("url is required"))
		return
	}

	cfg := s.BaseCfg
	if req.SourceID > 0 {
		cfg.SourceID = req.SourceID
	}
	if req.Concurrency > 0 {
		cfg.Concurrency = req.Concurrency
	}
	if req.Format != "" {
		cfg.Format = req.Format
	}
	if req.Start > 0 {
		cfg.ChapterStart = req.Start
	}
	cfg.ChapterEnd = req.End

	svc, err := app.New(cfg)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}

	job := s.newJob("pending", "任务已创建，等待执行")
	go s.runDownloadJob(job.ID, svc, req.URL)
	writeOK(w, job)
}

func (s *Server) handleJob(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		writeErr(w, http.StatusBadRequest, errors.New("id is required"))
		return
	}
	job, ok := s.getJob(id)
	if !ok {
		writeErr(w, http.StatusNotFound, errors.New("job not found"))
		return
	}
	writeOK(w, job)
}

func (s *Server) handleFiles(w http.ResponseWriter, _ *http.Request) {
	dir := filepath.Clean(s.BaseCfg.OutputDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			writeOK(w, []localFile{})
			return
		}
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	items := make([]localFile, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		items = append(items, localFile{
			Name:    e.Name(),
			IsDir:   e.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().UnixMilli(),
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ModTime > items[j].ModTime })
	writeOK(w, items)
}

func (s *Server) handleFile(w http.ResponseWriter, r *http.Request) {
	name := filepath.Base(strings.TrimSpace(r.URL.Query().Get("name")))
	if name == "." || name == "" {
		writeErr(w, http.StatusBadRequest, errors.New("invalid file name"))
		return
	}
	baseDir := filepath.Clean(s.BaseCfg.OutputDir)
	full := filepath.Join(baseDir, name)
	realBase, _ := filepath.Abs(baseDir)
	realFull, _ := filepath.Abs(full)
	if !strings.HasPrefix(realFull, realBase+string(os.PathSeparator)) && realFull != realBase {
		writeErr(w, http.StatusBadRequest, errors.New("invalid path"))
		return
	}
	info, err := os.Stat(realFull)
	if err != nil {
		writeErr(w, http.StatusNotFound, err)
		return
	}
	if info.IsDir() {
		writeErr(w, http.StatusBadRequest, errors.New("directory download is not supported"))
		return
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", name))
	http.ServeFile(w, r, realFull)
}

func writeOK(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, response{Code: 200, Message: "OK", Data: data})
}

func writeErr(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, response{Code: status, Message: err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, v response) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (s *Server) newJob(state, msg string) *jobStatus {
	id := fmt.Sprintf("job-%d", atomic.AddUint64(&s.jobSeq, 1))
	now := time.Now().UnixMilli()
	job := &jobStatus{
		ID:        id,
		State:     state,
		Message:   msg,
		StartedAt: now,
		UpdatedAt: now,
	}
	s.jobsMu.Lock()
	s.jobs[id] = job
	s.jobsMu.Unlock()
	return job
}

func (s *Server) setJob(id, state, msg, path string) {
	s.jobsMu.Lock()
	defer s.jobsMu.Unlock()
	job, ok := s.jobs[id]
	if !ok {
		return
	}
	if state != "" {
		job.State = state
	}
	if msg != "" {
		job.Message = msg
	}
	if path != "" {
		job.Path = path
	}
	job.UpdatedAt = time.Now().UnixMilli()
}

func (s *Server) getJob(id string) (jobStatus, bool) {
	s.jobsMu.RLock()
	defer s.jobsMu.RUnlock()
	job, ok := s.jobs[id]
	if !ok {
		return jobStatus{}, false
	}
	return *job, true
}

func (s *Server) runDownloadJob(jobID string, svc *app.Service, url string) {
	s.setJob(jobID, "running", "下载中...", "")
	// 序列化下载任务，避免并发写同目录导致混杂
	s.downloadMu.Lock()
	path, err := svc.DownloadByURL(url)
	s.downloadMu.Unlock()
	if err != nil {
		s.setJob(jobID, "error", err.Error(), "")
		return
	}
	s.setJob(jobID, "success", "下载完成", path)
}
