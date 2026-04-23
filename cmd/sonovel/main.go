package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/so-novel/sonovel-go/internal/app"
	"github.com/so-novel/sonovel-go/internal/model"
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
	format := fs.String("format", "epub", "output format: epub|txt|html")
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
	fmt.Println("done:", path)
}

func usage() {
	fmt.Println(`sonovel-go
  search   --kw "xx" --source-id 1 [--rules path]
  download --url "https://..." [--source-id 1] [--format epub|txt|html] [--out ./downloads]`)
}

func must(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
