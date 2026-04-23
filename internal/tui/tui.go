package tui

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/opso-code/sonovel-go/internal/app"
	"github.com/opso-code/sonovel-go/internal/model"
)

type TUI struct {
	Base model.Config
	r    *bufio.Reader
}

func New(base model.Config) *TUI {
	return &TUI{Base: base, r: bufio.NewReader(os.Stdin)}
}

func (t *TUI) Run() error {
	for {
		fmt.Println("\n=== sonovel-go TUI ===")
		fmt.Println("1) 搜索并下载")
		fmt.Println("2) URL 直接下载")
		fmt.Println("3) 查看本地文件")
		fmt.Println("0) 退出")
		s := t.readLine("请选择: ")
		switch s {
		case "1":
			t.searchAndDownload()
		case "2":
			t.downloadByURL()
		case "3":
			t.listLocalFiles()
		case "0", "q", "quit", "exit":
			return nil
		default:
			fmt.Println("无效选择")
		}
	}
}

func (t *TUI) searchAndDownload() {
	svc, err := app.New(t.Base)
	if err != nil {
		fmt.Println("初始化失败:", err)
		return
	}
	sources := make([]model.Rule, 0, len(svc.Rules))
	for _, r := range svc.Rules {
		if !r.Disabled {
			sources = append(sources, r)
		}
	}
	if len(sources) == 0 {
		fmt.Println("没有可用书源")
		return
	}

	fmt.Println("\n可用书源:")
	for _, s := range sources {
		fmt.Printf("%d) %s\n", s.ID, s.Name)
	}
	sid := t.readSourceID(sources)
	kw := t.readLine("输入书名或作者: ")
	if kw == "" {
		fmt.Println("关键词不能为空")
		return
	}

	cfg := t.collectDownloadOptions()
	cfg.SourceID = sid
	cfg.SearchLimit = t.readIntDefault("搜索条数限制(默认20): ", 20)
	svc, err = app.New(cfg)
	if err != nil {
		fmt.Println("初始化失败:", err)
		return
	}

	results, err := svc.Search(kw)
	if err != nil {
		fmt.Println("搜索失败:", err)
		return
	}
	if len(results) == 0 {
		fmt.Println("未找到结果")
		return
	}

	t.printBookResults(results)
	idx := t.readInt(fmt.Sprintf("选择序号(1-%d): ", len(results)), 1)
	if idx < 1 || idx > len(results) {
		fmt.Println("序号越界")
		return
	}
	chosen := results[idx-1]

	cfg.SourceID = chosen.SourceID
	fmt.Println("开始下载:", chosen.BookName)
	path, elapsed, err := t.runDownload(cfg, chosen.URL)
	if err != nil {
		fmt.Println("下载失败:", err)
		return
	}
	t.printDownloadDone(path, elapsed)
}

func (t *TUI) downloadByURL() {
	url := t.readLine("输入书籍 URL: ")
	if url == "" {
		fmt.Println("URL 不能为空")
		return
	}
	cfg := t.collectDownloadOptions()
	sid := t.readIntDefault("source-id(可留空自动匹配): ", 0)
	cfg.SourceID = sid

	fmt.Println("开始下载...")
	path, elapsed, err := t.runDownload(cfg, url)
	if err != nil {
		fmt.Println("下载失败:", err)
		return
	}
	t.printDownloadDone(path, elapsed)
}

func (t *TUI) listLocalFiles() {
	dir := filepath.Clean(t.Base.OutputDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Println("读取目录失败:", err)
		return
	}
	type item struct {
		name  string
		size  int64
		mtime int64
		dir   bool
	}
	items := make([]item, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		items = append(items, item{name: e.Name(), size: info.Size(), mtime: info.ModTime().UnixMilli(), dir: e.IsDir()})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].mtime > items[j].mtime })

	fmt.Println("\n本地文件:")
	for i, f := range items {
		typeName := "FILE"
		if f.dir {
			typeName = "DIR"
		}
		fmt.Printf("%d) [%s] %s (%d bytes)\n", i+1, typeName, f.name, f.size)
	}
	if len(items) == 0 {
		fmt.Println("(空)")
	}
}

func (t *TUI) collectDownloadOptions() model.Config {
	cfg := t.Base
	fmt.Println("格式:")
	fmt.Println("1) TXT (默认)")
	fmt.Println("2) EPUB")
	fmt.Println("3) HTML")
	switch t.readIntDefault("请选择格式 [1-3，默认1]: ", 1) {
	case 2:
		cfg.Format = "epub"
	case 3:
		cfg.Format = "html"
	default:
		cfg.Format = "txt"
	}
	cfg.Concurrency = t.readIntDefault("并发(默认12): ", 12)
	cfg.ChapterStart = t.readIntDefault("起始章节(默认1): ", 1)
	cfg.ChapterEnd = t.readIntDefault("结束章节(默认0表示全部): ", 0)
	return cfg
}

func (t *TUI) printBookResults(results []model.SearchResult) {
	fmt.Println("\n搜索结果:")
	const (
		wNo     = 4
		wBook   = 20
		wAuthor = 12
		wLatest = 20
		wUpdate = 12
		wStatus = 8
		wSource = 10
		wURL    = 26
	)
	sep := buildTableSep([]int{wNo, wBook, wAuthor, wLatest, wUpdate, wStatus, wSource, wURL})
	fmt.Println(sep)
	fmt.Printf("| %s | %s | %s | %s | %s | %s | %s | %s |\n",
		padDisplay("序号", wNo),
		padDisplay("书名", wBook),
		padDisplay("作者", wAuthor),
		padDisplay("最新章节", wLatest),
		padDisplay("更新时间", wUpdate),
		padDisplay("状态", wStatus),
		padDisplay("来源", wSource),
		padDisplay("链接", wURL),
	)
	fmt.Println(sep)
	for i, r := range results {
		fmt.Printf("| %s | %s | %s | %s | %s | %s | %s | %s |\n",
			padDisplay(strconv.Itoa(i+1), wNo),
			padDisplay(truncateDisplay(orDash(r.BookName), wBook), wBook),
			padDisplay(truncateDisplay(orDash(r.Author), wAuthor), wAuthor),
			padDisplay(truncateDisplay(orDash(r.LatestChapter), wLatest), wLatest),
			padDisplay(truncateDisplay(orDash(r.LastUpdateTime), wUpdate), wUpdate),
			padDisplay(truncateDisplay(orDash(r.Status), wStatus), wStatus),
			padDisplay(truncateDisplay(orDash(r.SourceName), wSource), wSource),
			padDisplay(truncateDisplay(orDash(r.URL), wURL), wURL),
		)
		fmt.Println(sep)
	}
}

func truncateDisplay(s string, maxWidth int) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "-"
	}
	if displayWidth(s) <= maxWidth {
		return s
	}
	rs := []rune(s)
	cur := 0
	var out []rune
	for _, r := range rs {
		w := runeWidth(r)
		if cur+w > max(1, maxWidth-1) {
			break
		}
		out = append(out, r)
		cur += w
	}
	return string(out) + "…"
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return strings.TrimSpace(s)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func padDisplay(s string, width int) string {
	w := displayWidth(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

func displayWidth(s string) int {
	w := 0
	for _, r := range s {
		w += runeWidth(r)
	}
	return w
}

func runeWidth(r rune) int {
	if r == 0 {
		return 0
	}
	if unicode.Is(unicode.Mn, r) {
		return 0
	}
	// 常见 CJK 统一按双宽处理，提升终端对齐稳定性。
	if unicode.Is(unicode.Han, r) || unicode.Is(unicode.Hangul, r) || unicode.Is(unicode.Hiragana, r) || unicode.Is(unicode.Katakana, r) {
		return 2
	}
	// 全角符号区
	if r >= 0xFF01 && r <= 0xFF60 {
		return 2
	}
	return 1
}

func buildTableSep(widths []int) string {
	var b strings.Builder
	b.WriteString("+")
	for _, w := range widths {
		b.WriteString(strings.Repeat("-", w+2))
		b.WriteString("+")
	}
	return b.String()
}

func (t *TUI) readSourceID(sources []model.Rule) int {
	if len(sources) == 0 {
		return 1
	}
	defaultID := sources[0].ID
	in := t.readIntDefault(fmt.Sprintf("输入 source-id (默认%d): ", defaultID), defaultID)
	for _, s := range sources {
		if s.ID == in {
			return in
		}
	}
	fmt.Printf("source-id=%d 不存在，回退默认 %d\n", in, defaultID)
	return defaultID
}

func (t *TUI) runDownload(cfg model.Config, url string) (string, time.Duration, error) {
	start := time.Now()
	cfg.OnProgress = t.makeProgressBar(start)
	svc, err := app.New(cfg)
	if err != nil {
		return "", 0, err
	}
	path, err := svc.DownloadByURL(url)
	elapsed := time.Since(start).Round(time.Second)
	fmt.Print("\n")
	return path, elapsed, err
}

func (t *TUI) makeProgressBar(start time.Time) func(done, total int) {
	const width = 28
	return func(done, total int) {
		if total <= 0 {
			fmt.Printf("\r下载中... 已耗时 %s", time.Since(start).Round(time.Second))
			return
		}
		if done > total {
			done = total
		}
		ratio := float64(done) / float64(total)
		filled := int(ratio * width)
		if filled > width {
			filled = width
		}
		bar := strings.Repeat("=", filled) + strings.Repeat("-", width-filled)
		fmt.Printf("\r[%s] %3d%% (%d/%d) 已耗时 %s",
			bar, int(ratio*100), done, total, time.Since(start).Round(time.Second))
	}
}

func (t *TUI) printDownloadDone(path string, elapsed time.Duration) {
	size, err := pathSize(path)
	if err != nil {
		fmt.Printf("下载完成: %s (耗时 %s，大小未知: %v)\n", path, elapsed, err)
		return
	}
	fmt.Printf("下载完成: %s\n", path)
	fmt.Printf("文件大小: %s | 耗时: %s\n", humanSize(size), elapsed)
}

func pathSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	if !info.IsDir() {
		return info.Size(), nil
	}
	var total int64
	err = filepath.WalkDir(path, func(_ string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			return nil
		}
		fi, e := d.Info()
		if e != nil {
			return e
		}
		total += fi.Size()
		return nil
	})
	return total, err
}

func humanSize(n int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case n >= GB:
		return fmt.Sprintf("%.2f GB", float64(n)/float64(GB))
	case n >= MB:
		return fmt.Sprintf("%.2f MB", float64(n)/float64(MB))
	case n >= KB:
		return fmt.Sprintf("%.2f KB", float64(n)/float64(KB))
	default:
		return fmt.Sprintf("%d B", n)
	}
}

func (t *TUI) readLine(prompt string) string {
	fmt.Print(prompt)
	s, _ := t.r.ReadString('\n')
	return strings.TrimSpace(s)
}

func (t *TUI) readInt(prompt string, fallback int) int {
	s := t.readLine(prompt)
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return fallback
	}
	return n
}

func (t *TUI) readIntDefault(prompt string, fallback int) int {
	s := t.readLine(prompt)
	if s == "" {
		return fallback
	}
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return fallback
	}
	return n
}
