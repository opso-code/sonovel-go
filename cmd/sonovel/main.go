package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/opso-code/sonovel-go/internal/app"
	"github.com/opso-code/sonovel-go/internal/appcfg"
	"github.com/opso-code/sonovel-go/internal/model"
	"github.com/opso-code/sonovel-go/internal/tui"
	"github.com/opso-code/sonovel-go/internal/web"
)

const defaultConfigPath = "./config.toml"

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
	case "init":
		runInit(os.Args[2:])
	default:
		usage()
		os.Exit(1)
	}
}

func runSearch(args []string) {
	configPath, fileCfg, err := loadRuntimeConfig(args)
	must(err)

	fs := flag.NewFlagSet("search", flag.ExitOnError)
	_ = fs.String("config", configPath, "config toml path")
	kw := fs.String("kw", "", "keyword")
	rules := fs.String("rules", fileCfg.RulesFile, "rule file path")
	sourceID := fs.Int("source-id", fileCfg.SourceID, "source id")
	limit := fs.Int("limit", fileCfg.SearchLimit, "search limit")
	format := fs.String("format", fileCfg.Format, "download format after selecting row: txt|epub|html")
	out := fs.String("out", fileCfg.OutputDir, "output directory")
	concurrency := fs.Int("concurrency", fileCfg.Concurrency, "download concurrency")
	start := fs.Int("start", fileCfg.ChapterStart, "chapter start index (1-based)")
	end := fs.Int("end", fileCfg.ChapterEnd, "chapter end index (0 means all)")
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
	tui.PrintSearchResultsTable(results)
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("输入序号直接下载（回车跳过）: ")
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	pick, err := strconv.Atoi(line)
	if err != nil || pick < 1 || pick > len(results) {
		fmt.Println("无效序号，已跳过下载")
		return
	}
	chosen := results[pick-1]
	dcfg := modelConfigFromFile(fileCfg)
	dcfg.RulesFile = *rules
	dcfg.OutputDir = *out
	dcfg.SourceID = chosen.SourceID
	dcfg.Format = *format
	dcfg.Concurrency = *concurrency
	dcfg.ChapterStart = *start
	dcfg.ChapterEnd = *end
	dsvc, err := app.New(dcfg)
	must(err)
	path, err := dsvc.DownloadByURL(chosen.URL)
	must(err)
	size, e := pathSize(path)
	if e != nil {
		fmt.Printf("done: %s (size unknown: %v)\n", path, e)
		return
	}
	fmt.Printf("done: %s (%s)\n", path, humanSize(size))
}

func runDownload(args []string) {
	configPath, fileCfg, err := loadRuntimeConfig(args)
	must(err)

	fs := flag.NewFlagSet("download", flag.ExitOnError)
	_ = fs.String("config", configPath, "config toml path")
	url := fs.String("url", "", "book url")
	rules := fs.String("rules", fileCfg.RulesFile, "rule file path")
	sourceID := fs.Int("source-id", fileCfg.SourceID, "source id, optional when url can auto match")
	format := fs.String("format", fileCfg.Format, "output format: txt|epub|html")
	out := fs.String("out", fileCfg.OutputDir, "output directory")
	concurrency := fs.Int("concurrency", fileCfg.Concurrency, "download concurrency")
	start := fs.Int("start", fileCfg.ChapterStart, "chapter start index (1-based)")
	end := fs.Int("end", fileCfg.ChapterEnd, "chapter end index (0 means all)")
	_ = fs.Parse(args)

	if strings.TrimSpace(*url) == "" {
		fmt.Println("url is required")
		os.Exit(2)
	}

	cfg := modelConfigFromFile(fileCfg)
	cfg.RulesFile = *rules
	cfg.SourceID = *sourceID
	cfg.OutputDir = *out
	cfg.Format = *format
	cfg.Concurrency = *concurrency
	cfg.ChapterStart = *start
	cfg.ChapterEnd = *end

	svc, err := app.New(cfg)
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
	configPath, fileCfg, err := loadRuntimeConfig(args)
	must(err)

	fs := flag.NewFlagSet("web", flag.ExitOnError)
	_ = fs.String("config", configPath, "config toml path")
	rules := fs.String("rules", fileCfg.RulesFile, "rule file path")
	out := fs.String("out", fileCfg.OutputDir, "output directory")
	format := fs.String("format", fileCfg.Format, "default output format")
	port := fs.Int("port", fileCfg.WebPort, "web port")
	openBrowser := fs.Bool("open", fileCfg.WebOpen, "auto open default browser")
	_ = fs.Parse(args)

	cfg := modelConfigFromFile(fileCfg)
	cfg.RulesFile = *rules
	cfg.OutputDir = *out
	cfg.Format = *format

	srv := web.New(cfg)
	addr := fmt.Sprintf(":%d", *port)
	url := fmt.Sprintf("http://localhost%s", addr)
	fmt.Printf("Web UI: %s\n", url)
	if *openBrowser {
		if err := openURL(url); err != nil {
			fmt.Printf("open browser failed: %v\n", err)
		}
	}
	must(http.ListenAndServe(addr, srv.Handler()))
}

func runTUI(args []string) {
	configPath, fileCfg, err := loadRuntimeConfig(args)
	must(err)

	fs := flag.NewFlagSet("tui", flag.ExitOnError)
	_ = fs.String("config", configPath, "config toml path")
	rules := fs.String("rules", fileCfg.RulesFile, "rule file path")
	out := fs.String("out", fileCfg.OutputDir, "output directory")
	format := fs.String("format", fileCfg.Format, "default output format")
	_ = fs.Parse(args)

	cfg := modelConfigFromFile(fileCfg)
	cfg.RulesFile = *rules
	cfg.OutputDir = *out
	cfg.Format = *format
	must(tui.New(cfg).Run())
}

func runInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	configPath := fs.String("config", defaultConfigPath, "config toml path")
	rulesDir := fs.String("rules-dir", "./rules", "rules directory")
	outDir := fs.String("out", "./downloads", "output directory")
	force := fs.Bool("force", false, "overwrite config file if exists")
	_ = fs.Parse(args)

	if err := os.MkdirAll(*rulesDir, 0o755); err != nil {
		must(err)
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		must(err)
	}

	if _, err := os.Stat(*configPath); err == nil && !*force {
		fmt.Printf("config already exists: %s (use --force to overwrite)\n", *configPath)
		return
	}
	if err := os.MkdirAll(filepath.Dir(*configPath), 0o755); err != nil {
		must(err)
	}
	tpl := appcfg.Template()
	tpl = strings.Replace(tpl, "./rules/main.json", filepath.ToSlash(filepath.Join(*rulesDir, "main.json")), 1)
	tpl = strings.Replace(tpl, "./downloads", filepath.ToSlash(*outDir), 1)
	if err := os.WriteFile(*configPath, []byte(tpl), 0o644); err != nil {
		must(err)
	}

	fmt.Printf("initialized: %s\n", *configPath)
	fmt.Printf("rules dir : %s\n", *rulesDir)
	fmt.Printf("output dir: %s\n", *outDir)
	if _, err := os.Stat(filepath.Join(*rulesDir, "main.json")); err != nil {
		fmt.Printf("warning: %s not found, please place rule files there\n", filepath.Join(*rulesDir, "main.json"))
	}
}

func usage() {
	fmt.Println(`sonovel-go
  init     [--config ./config.toml] [--rules-dir ./rules] [--out ./downloads]
  search   --kw "xx" [--config ./config.toml] [--source-id 1] [--rules path] [可直接选行下载]
  download --url "https://..." [--config ./config.toml] [--source-id 1] [--format txt|epub|html] [--out ./downloads]
  tui      [--config ./config.toml] [--rules path] [--out ./downloads]
  web      [--config ./config.toml] [--port 7765] [--rules path] [--out ./downloads] [--format txt] [--open true|false]`)
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

func openURL(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func modelConfigFromFile(c appcfg.Config) model.Config {
	return model.Config{
		RulesFile:     c.RulesFile,
		SourceID:      c.SourceID,
		Concurrency:   c.Concurrency,
		MinIntervalMS: c.MinIntervalMS,
		MaxIntervalMS: c.MaxIntervalMS,
		EnableRetry:   c.EnableRetry,
		MaxRetries:    c.MaxRetries,
		RetryMinMS:    c.RetryMinMS,
		RetryMaxMS:    c.RetryMaxMS,
		OutputDir:     c.OutputDir,
		Format:        c.Format,
		ChapterStart:  c.ChapterStart,
		ChapterEnd:    c.ChapterEnd,
		SearchLimit:   c.SearchLimit,
	}
}

func loadRuntimeConfig(args []string) (string, appcfg.Config, error) {
	path := resolveConfigPath(args, defaultConfigPath)
	cfg, err := appcfg.Load(path)
	if err != nil {
		return "", appcfg.Config{}, err
	}
	return path, cfg, nil
}

func resolveConfigPath(args []string, fallback string) string {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "-config" || a == "--config" {
			if i+1 < len(args) && strings.TrimSpace(args[i+1]) != "" {
				return args[i+1]
			}
			continue
		}
		if strings.HasPrefix(a, "--config=") {
			v := strings.TrimSpace(strings.TrimPrefix(a, "--config="))
			if v != "" {
				return v
			}
		}
		if strings.HasPrefix(a, "-config=") {
			v := strings.TrimSpace(strings.TrimPrefix(a, "-config="))
			if v != "" {
				return v
			}
		}
	}
	return fallback
}
