package web

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"

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
	BookName    string `json:"bookName"`
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
	ID             string `json:"id"`
	State          string `json:"state"`
	Message        string `json:"message"`
	BookName       string `json:"bookName,omitempty"`
	CurrentChapter string `json:"currentChapter,omitempty"`
	Done           int    `json:"done"`
	Total          int    `json:"total"`
	Path           string `json:"path,omitempty"`
	CancelReq      bool   `json:"cancelRequested"`
	StartedAt      int64  `json:"startedAt"`
	UpdatedAt      int64  `json:"updatedAt"`
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
	s.mux.HandleFunc("/api/jobs", s.handleJobs)
	s.mux.HandleFunc("/api/job/cancel", s.handleJobCancel)
	s.mux.HandleFunc("/api/files", s.handleFiles)
	s.mux.HandleFunc("/api/open-output", s.handleOpenOutput)
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
	limit := parseSearchLimit(r.URL.Query().Get("limit"))
	if kw == "" {
		writeErr(w, http.StatusBadRequest, errors.New("kw is required"))
		return
	}

	var (
		items []model.SearchResult
		err   error
	)
	if sid > 0 {
		cfg := s.BaseCfg
		cfg.SourceID = sid
		cfg.SearchLimit = limit
		svc, e := app.New(cfg)
		if e != nil {
			writeErr(w, http.StatusInternalServerError, e)
			return
		}
		items, err = svc.Search(kw)
	} else {
		items, err = s.searchAllSources(kw, limit)
	}
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	if len(items) > limit {
		items = items[:limit]
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

	job := s.newJob("pending", "任务已创建，等待执行", req.BookName)
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

func (s *Server) handleJobs(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("limit")))
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	writeOK(w, s.listJobs(limit))
}

func (s *Server) handleJobCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	if strings.TrimSpace(req.ID) == "" {
		writeErr(w, http.StatusBadRequest, errors.New("id is required"))
		return
	}
	job, err := s.requestCancel(req.ID)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
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
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		items = append(items, localFile{
			Name:    name,
			IsDir:   e.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().UnixMilli(),
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ModTime > items[j].ModTime })
	writeOK(w, items)
}

func (s *Server) handleOpenOutput(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}
	dir := filepath.Clean(s.BaseCfg.OutputDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	cmd, err := openDirCommand(abs)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	if err := cmd.Start(); err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeOK(w, map[string]string{"path": abs})
}

func openDirCommand(path string) (*exec.Cmd, error) {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", path), nil
	case "windows":
		return exec.Command("cmd", "/c", "start", "", path), nil
	default:
		return exec.Command("xdg-open", path), nil
	}
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

func (s *Server) newJob(state, msg, bookName string) *jobStatus {
	id := fmt.Sprintf("job-%d", atomic.AddUint64(&s.jobSeq, 1))
	now := time.Now().UnixMilli()
	job := &jobStatus{
		ID:        id,
		State:     state,
		Message:   msg,
		BookName:  strings.TrimSpace(bookName),
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
		if state == "error" {
			msg = compactErrMessage(msg)
		}
		job.Message = msg
	}
	if path != "" {
		job.Path = path
	}
	if job.State == "success" || job.State == "error" || job.State == "canceled" {
		job.CancelReq = false
	}
	job.UpdatedAt = time.Now().UnixMilli()
}

func compactErrMessage(msg string) string {
	msg = strings.TrimSpace(msg)
	msg = strings.ReplaceAll(msg, "\n", " ")
	msg = strings.ReplaceAll(msg, "\t", " ")
	msg = strings.Join(strings.Fields(msg), " ")
	const max = 180
	if len([]rune(msg)) <= max {
		return msg
	}
	rs := []rune(msg)
	return string(rs[:max]) + "..."
}

func (s *Server) setJobProgress(id string, done, total int, chapter string) {
	s.jobsMu.Lock()
	defer s.jobsMu.Unlock()
	job, ok := s.jobs[id]
	if !ok {
		return
	}
	if job.State == "success" || job.State == "error" || job.State == "canceled" {
		return
	}
	if done >= 0 {
		job.Done = done
	}
	if total > 0 {
		job.Total = total
	}
	chapter = strings.TrimSpace(chapter)
	if chapter != "" {
		job.CurrentChapter = chapter
	}
	if job.Total > 0 {
		job.Message = fmt.Sprintf("正在下载章节 %d/%d", job.Done, job.Total)
	} else {
		job.Message = "正在下载章节..."
	}
	if job.CurrentChapter != "" {
		job.Message = job.Message + " · " + job.CurrentChapter
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

func (s *Server) listJobs(limit int) []jobStatus {
	s.jobsMu.RLock()
	defer s.jobsMu.RUnlock()
	out := make([]jobStatus, 0, len(s.jobs))
	for _, job := range s.jobs {
		out = append(out, *job)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].UpdatedAt == out[j].UpdatedAt {
			return out[i].StartedAt > out[j].StartedAt
		}
		return out[i].UpdatedAt > out[j].UpdatedAt
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

func (s *Server) requestCancel(id string) (jobStatus, error) {
	s.jobsMu.Lock()
	defer s.jobsMu.Unlock()
	job, ok := s.jobs[id]
	if !ok {
		return jobStatus{}, errors.New("job not found")
	}
	switch job.State {
	case "success", "error", "canceled":
		return jobStatus{}, errors.New("job already finished")
	}
	job.CancelReq = true
	if job.State == "pending" {
		job.State = "canceled"
		job.Message = "任务已取消"
	} else {
		job.Message = "正在取消任务..."
	}
	job.UpdatedAt = time.Now().UnixMilli()
	return *job, nil
}

func (s *Server) isCancelRequested(id string) bool {
	s.jobsMu.RLock()
	defer s.jobsMu.RUnlock()
	job, ok := s.jobs[id]
	if !ok {
		return false
	}
	return job.CancelReq
}

func (s *Server) runDownloadJob(jobID string, svc *app.Service, url string) {
	s.setJob(jobID, "pending", "任务排队中...", "")
	// 序列化下载任务，避免并发写同目录导致混杂
	s.downloadMu.Lock()
	if s.isCancelRequested(jobID) {
		s.downloadMu.Unlock()
		s.setJob(jobID, "canceled", "任务已取消", "")
		return
	}
	s.setJob(jobID, "running", "开始下载章节...", "")
	svc.Cfg.ShouldCancel = func() bool {
		return s.isCancelRequested(jobID)
	}
	svc.Cfg.OnChapter = func(done, total int, title string) {
		s.setJobProgress(jobID, done, total, title)
	}
	svc.Cfg.OnProgress = func(done, total int) {
		s.setJobProgress(jobID, done, total, "")
	}
	path, err := svc.DownloadByURL(url)
	s.downloadMu.Unlock()
	if err != nil {
		if s.isCancelRequested(jobID) {
			s.setJob(jobID, "canceled", "任务已取消", "")
			return
		}
		s.setJob(jobID, "error", err.Error(), "")
		return
	}
	if s.isCancelRequested(jobID) {
		s.setJob(jobID, "canceled", "任务已取消", "")
		return
	}
	s.setJobProgress(jobID, -1, 0, "")
	s.setJob(jobID, "success", "下载完成", path)
}

func parseSearchLimit(raw string) int {
	limit, _ := strconv.Atoi(strings.TrimSpace(raw))
	if limit <= 0 {
		limit = 60
	}
	if limit > 200 {
		limit = 200
	}
	return limit
}

func (s *Server) searchAllSources(kw string, limit int) ([]model.SearchResult, error) {
	baseSvc, err := app.New(s.BaseCfg)
	if err != nil {
		return nil, err
	}
	type task struct {
		index int
		id    int
	}
	type result struct {
		index int
		items []model.SearchResult
		err   error
	}

	tasks := make([]task, 0, len(baseSvc.Rules))
	for i, r := range baseSvc.Rules {
		if r.Disabled || r.Search == nil || r.Search.Disabled {
			continue
		}
		tasks = append(tasks, task{index: i, id: r.ID})
	}
	if len(tasks) == 0 {
		return nil, errors.New("no searchable sources")
	}

	sem := make(chan struct{}, 8)
	ch := make(chan result, len(tasks))
	kwNorm := normalizeSearchText(kw)
	const aggregateTimeout = 4 * time.Second

	for _, t := range tasks {
		sem <- struct{}{}
		go func(t task) {
			defer func() { <-sem }()

			cfg := s.BaseCfg
			cfg.SourceID = t.id
			cfg.SearchLimit = limit
			svc, e := app.New(cfg)
			if e != nil {
				ch <- result{index: t.index, err: e}
				return
			}
			items, e := svc.Search(kw)
			if e != nil {
				ch <- result{index: t.index, err: e}
				return
			}
			ch <- result{index: t.index, items: items}
		}(t)
	}

	type scored struct {
		score      int
		sourceRank int
		itemRank   int
		item       model.SearchResult
	}
	list := make([]scored, 0, limit*2)
	deadline := time.NewTimer(aggregateTimeout)
	defer deadline.Stop()
	completed := 0

	for completed < len(tasks) {
		select {
		case r := <-ch:
			completed++
			for i, item := range r.items {
				list = append(list, scored{
					score:      scoreSearchResult(item, kwNorm),
					sourceRank: r.index,
					itemRank:   i,
					item:       item,
				})
			}
		case <-deadline.C:
			completed = len(tasks)
		}
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i].score != list[j].score {
			return list[i].score > list[j].score
		}
		if list[i].sourceRank != list[j].sourceRank {
			return list[i].sourceRank < list[j].sourceRank
		}
		return list[i].itemRank < list[j].itemRank
	})

	out := make([]model.SearchResult, 0, minInt(limit, len(list)))
	for i := 0; i < len(list) && len(out) < limit; i++ {
		out = append(out, list[i].item)
	}
	return out, nil
}

func scoreSearchResult(item model.SearchResult, kw string) int {
	if kw == "" {
		return 0
	}
	name := normalizeSearchText(item.BookName)
	author := normalizeSearchText(item.Author)
	latest := normalizeSearchText(item.LatestChapter)

	score := 0
	switch {
	case name == kw:
		score += 1200
	case strings.HasPrefix(name, kw):
		score += 930
	case strings.Contains(name, kw):
		score += 760
	}
	switch {
	case author == kw:
		score += 420
	case strings.Contains(author, kw):
		score += 260
	}
	if strings.Contains(latest, kw) {
		score += 60
	}

	// 对短书名做轻微加权，优先展示更“像标题”的结果
	nameLen := len([]rune(name))
	if nameLen > 0 {
		score += maxInt(0, 50-minInt(50, nameLen))
	}
	return score
}

func normalizeSearchText(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(s))
	lastSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !lastSpace && b.Len() > 0 {
				b.WriteByte(' ')
				lastSpace = true
			}
			continue
		}
		if isPunctRune(r) {
			continue
		}
		b.WriteRune(r)
		lastSpace = false
	}
	return strings.TrimSpace(b.String())
}

func isPunctRune(r rune) bool {
	if unicode.IsPunct(r) || unicode.IsSymbol(r) {
		return true
	}
	switch r {
	case '《', '》', '【', '】', '（', '）', '“', '”', '·':
		return true
	default:
		return false
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
