package main

import (
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/opso-code/sonovel-go/internal/app"
	"github.com/opso-code/sonovel-go/internal/model"
	"github.com/opso-code/sonovel-go/internal/tui"
	"github.com/opso-code/sonovel-go/internal/web"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	sub := os.Args[1]
	switch sub {
	case "search":
		runSearch(os.Args[2:])
	case "download":
		runDownload(os.Args[2:])
	case "web":
		runWeb(os.Args[2:])
	case "tui":
		runTUI(os.Args[2:])
	default:
		usage()
		os.Exit(1)
	}
}

func runSearch(args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	kw := fs.String("kw", "", "keyword")
	rules := fs.String("rules", "../bundle/rules/main.json", "rule file path")
	sourceID := fs.Int("source-id", 1, "source id")
	limit := fs.Int("limit", 20, "search limit")
	_ = fs.Parse(args)

	if strings.TrimSpace(*kw) == "" {
		fmt.Println("kw is required")
		os.Exit(2)
	}

	svc, err := app.New(model.Config{
		RulesFile:   *rules,
		SourceID:    *sourceID,
		SearchLimit: *limit,
	})
	must(err)

	results, err := svc.Search(*kw)
	must(err)
	if len(results) == 0 {
		fmt.Println("no result")
		return
	}
	for i, r := range results {
		fmt.Printf("%d. [%d:%s] %s - %s\n   %s\n", i+1, r.SourceID, r.SourceName, r.BookName, r.Author, r.URL)
	}
}

func runDownload(args []string) {
	fs := flag.NewFlagSet("download", flag.ExitOnError)
	url := fs.String("url", "", "book url")
	rules := fs.String("rules", "../bundle/rules/main.json", "rule file path")
	sourceID := fs.Int("source-id", 0, "source id, optional when url can auto match")
	format := fs.String("format", "txt", "output format: txt|epub|html")
	out := fs.String("out", "./downloads", "output directory")
	concurrency := fs.Int("concurrency", 12, "download concurrency")
	start := fs.Int("start", 1, "chapter start index (1-based)")
	end := fs.Int("end", 0, "chapter end index (0 means all)")
	_ = fs.Parse(args)

	if strings.TrimSpace(*url) == "" {
		fmt.Println("url is required")
		os.Exit(2)
	}

	svc, err := app.New(model.Config{
		RulesFile:     *rules,
		SourceID:      *sourceID,
		Concurrency:   *concurrency,
		MinIntervalMS: 200,
		MaxIntervalMS: 450,
		EnableRetry:   true,
		MaxRetries:    3,
		RetryMinMS:    1000,
		RetryMaxMS:    2000,
		OutputDir:     *out,
		Format:        *format,
		ChapterStart:  *start,
		ChapterEnd:    *end,
	})
	must(err)

	path, err := svc.DownloadByURL(*url)
	must(err)
	size, e := pathSize(path)
	if e != nil {
		fmt.Printf("done: %s (size unknown: %v)\n", path, e)
		return
	}
	fmt.Printf("done: %s (%s)\n", path, humanSize(size))
}

func runWeb(args []string) {
	fs := flag.NewFlagSet("web", flag.ExitOnError)
	rules := fs.String("rules", "../bundle/rules/main.json", "rule file path")
	out := fs.String("out", "./downloads", "output directory")
	format := fs.String("format", "txt", "default output format")
	port := fs.Int("port", 7765, "web port")
	_ = fs.Parse(args)

	cfg := model.Config{
		RulesFile:     *rules,
		Concurrency:   12,
		MinIntervalMS: 200,
		MaxIntervalMS: 450,
		EnableRetry:   true,
		MaxRetries:    3,
		RetryMinMS:    1000,
		RetryMaxMS:    2000,
		OutputDir:     *out,
		Format:        *format,
		ChapterStart:  1,
		ChapterEnd:    0,
		SearchLimit:   20,
	}

	srv := web.New(cfg)
	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Web UI: http://localhost%s\n", addr)
	must(http.ListenAndServe(addr, srv.Handler()))
}

func runTUI(args []string) {
	fs := flag.NewFlagSet("tui", flag.ExitOnError)
	rules := fs.String("rules", "../bundle/rules/main.json", "rule file path")
	out := fs.String("out", "./downloads", "output directory")
	_ = fs.Parse(args)

	cfg := model.Config{
		RulesFile:     *rules,
		Concurrency:   12,
		MinIntervalMS: 200,
		MaxIntervalMS: 450,
		EnableRetry:   true,
		MaxRetries:    3,
		RetryMinMS:    1000,
		RetryMaxMS:    2000,
		OutputDir:     *out,
		Format:        "txt",
		ChapterStart:  1,
		ChapterEnd:    0,
		SearchLimit:   20,
	}
	must(tui.New(cfg).Run())
}

func usage() {
	fmt.Println(`sonovel-go
  search   --kw "xx" --source-id 1 [--rules path]
  download --url "https://..." [--source-id 1] [--format txt|epub|html] [--out ./downloads]
  tui      [--rules path] [--out ./downloads]
  web      [--port 7765] [--rules path] [--out ./downloads] [--format txt]`)
}

func must(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
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
