package appcfg

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	RulesFile     string
	OutputDir     string
	Format        string
	SourceID      int
	SearchLimit   int
	Concurrency   int
	ChapterStart  int
	ChapterEnd    int
	MinIntervalMS int
	MaxIntervalMS int
	EnableRetry   bool
	MaxRetries    int
	RetryMinMS    int
	RetryMaxMS    int
	WebPort       int
	WebOpen       bool
}

func Defaults() Config {
	return Config{
		RulesFile:     "./rules/main.json",
		OutputDir:     "./downloads",
		Format:        "txt",
		SourceID:      1,
		SearchLimit:   20,
		Concurrency:   12,
		ChapterStart:  1,
		ChapterEnd:    0,
		MinIntervalMS: 200,
		MaxIntervalMS: 450,
		EnableRetry:   true,
		MaxRetries:    3,
		RetryMinMS:    1000,
		RetryMaxMS:    2000,
		WebPort:       7765,
		WebOpen:       true,
	}
}

func Load(path string) (Config, error) {
	cfg := Defaults()
	p := strings.TrimSpace(path)
	if p == "" {
		return cfg, nil
	}
	f, err := os.Open(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}
	defer f.Close()

	section := ""
	sc := bufio.NewScanner(f)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(stripComment(sc.Text()))
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(strings.TrimSpace(line[1 : len(line)-1]))
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			return cfg, fmt.Errorf("%s:%d invalid line: %s", p, lineNo, line)
		}
		key := strings.ToLower(strings.TrimSpace(k))
		val := strings.TrimSpace(v)
		full := key
		if section != "" {
			full = section + "." + key
		}
		if err := assign(&cfg, full, val, p, lineNo); err != nil {
			return cfg, err
		}
	}
	if err := sc.Err(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func Template() string {
	return `# sonovel-go 配置文件（TOML）
# 命令行参数优先级更高，会覆盖这里的值。

rules_file = "./rules/main.json"
output_dir = "./downloads"
format = "txt"            # txt | epub | html
source_id = 1
search_limit = 20

concurrency = 12
chapter_start = 1
chapter_end = 0           # 0 表示到最后一章

min_interval_ms = 200
max_interval_ms = 450
enable_retry = true
max_retries = 3
retry_min_ms = 1000
retry_max_ms = 2000

[web]
port = 7765
open_browser = true
`
}

func assign(cfg *Config, key, raw, path string, lineNo int) error {
	switch key {
	case "rules_file":
		v, err := parseString(raw)
		if err != nil {
			return parseErr(path, lineNo, key, err)
		}
		cfg.RulesFile = v
	case "output_dir":
		v, err := parseString(raw)
		if err != nil {
			return parseErr(path, lineNo, key, err)
		}
		cfg.OutputDir = v
	case "format":
		v, err := parseString(raw)
		if err != nil {
			return parseErr(path, lineNo, key, err)
		}
		cfg.Format = strings.ToLower(v)
	case "source_id":
		v, err := parseInt(raw)
		if err != nil {
			return parseErr(path, lineNo, key, err)
		}
		cfg.SourceID = v
	case "search_limit":
		v, err := parseInt(raw)
		if err != nil {
			return parseErr(path, lineNo, key, err)
		}
		cfg.SearchLimit = v
	case "concurrency":
		v, err := parseInt(raw)
		if err != nil {
			return parseErr(path, lineNo, key, err)
		}
		cfg.Concurrency = v
	case "chapter_start":
		v, err := parseInt(raw)
		if err != nil {
			return parseErr(path, lineNo, key, err)
		}
		cfg.ChapterStart = v
	case "chapter_end":
		v, err := parseInt(raw)
		if err != nil {
			return parseErr(path, lineNo, key, err)
		}
		cfg.ChapterEnd = v
	case "min_interval_ms":
		v, err := parseInt(raw)
		if err != nil {
			return parseErr(path, lineNo, key, err)
		}
		cfg.MinIntervalMS = v
	case "max_interval_ms":
		v, err := parseInt(raw)
		if err != nil {
			return parseErr(path, lineNo, key, err)
		}
		cfg.MaxIntervalMS = v
	case "enable_retry":
		v, err := parseBool(raw)
		if err != nil {
			return parseErr(path, lineNo, key, err)
		}
		cfg.EnableRetry = v
	case "max_retries":
		v, err := parseInt(raw)
		if err != nil {
			return parseErr(path, lineNo, key, err)
		}
		cfg.MaxRetries = v
	case "retry_min_ms":
		v, err := parseInt(raw)
		if err != nil {
			return parseErr(path, lineNo, key, err)
		}
		cfg.RetryMinMS = v
	case "retry_max_ms":
		v, err := parseInt(raw)
		if err != nil {
			return parseErr(path, lineNo, key, err)
		}
		cfg.RetryMaxMS = v
	case "web.port":
		v, err := parseInt(raw)
		if err != nil {
			return parseErr(path, lineNo, key, err)
		}
		cfg.WebPort = v
	case "web.open_browser":
		v, err := parseBool(raw)
		if err != nil {
			return parseErr(path, lineNo, key, err)
		}
		cfg.WebOpen = v
	}
	return nil
}

func parseErr(path string, lineNo int, key string, err error) error {
	return fmt.Errorf("%s:%d invalid %s: %w", path, lineNo, key, err)
}

func parseString(s string) (string, error) {
	v := strings.TrimSpace(s)
	if len(v) < 2 {
		return "", errors.New("string must be quoted")
	}
	if (v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'') {
		return v[1 : len(v)-1], nil
	}
	return "", errors.New("string must be quoted")
}

func parseInt(s string) (int, error) {
	i, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0, err
	}
	return i, nil
}

func parseBool(s string) (bool, error) {
	b, err := strconv.ParseBool(strings.TrimSpace(s))
	if err != nil {
		return false, err
	}
	return b, nil
}

func stripComment(s string) string {
	inQuote := byte(0)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inQuote == 0 {
			if c == '"' || c == '\'' {
				inQuote = c
				continue
			}
			if c == '#' {
				return s[:i]
			}
			continue
		}
		if c == inQuote {
			inQuote = 0
		}
	}
	return s
}
