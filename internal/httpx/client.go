package httpx

import (
	"bytes"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_4) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Safari/605.1.15",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
}

type Client struct {
	hc *http.Client
}

func New(timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return &Client{hc: &http.Client{Timeout: timeout}}
}

func (c *Client) Get(rawURL string, timeoutSec int, cookie string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	setHeaders(req, rawURL, cookie)
	return c.do(req, timeoutSec)
}

func (c *Client) PostForm(rawURL, data string, args []string, timeoutSec int, cookie string) ([]byte, error) {
	form := parseDataTemplate(data, args)
	req, err := http.NewRequest(http.MethodPost, rawURL, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	setHeaders(req, rawURL, cookie)
	return c.do(req, timeoutSec)
}

func (c *Client) do(req *http.Request, timeoutSec int) ([]byte, error) {
	hc := *c.hc
	if timeoutSec > 0 {
		hc.Timeout = time.Duration(timeoutSec) * time.Second
	}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, string(body))
	}
	return io.ReadAll(resp.Body)
}

func setHeaders(req *http.Request, rawURL string, cookie string) {
	host := req.URL.Scheme + "://" + req.URL.Host
	req.Header.Set("User-Agent", userAgents[rand.IntN(len(userAgents))])
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Referer", host)
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	_ = rawURL
}

func parseDataTemplate(data string, args []string) url.Values {
	v := url.Values{}
	data = strings.TrimSpace(data)
	data = strings.TrimPrefix(data, "{")
	data = strings.TrimSuffix(data, "}")

	parts := splitCSVLike(data)
	argIdx := 0
	for _, p := range parts {
		kv := strings.SplitN(p, ":", 2)
		if len(kv) != 2 {
			continue
		}
		k := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		k = trimQuote(k)
		val = trimQuote(val)
		if val == "%s" && argIdx < len(args) {
			val = args[argIdx]
			argIdx++
		}
		val = regexp.MustCompile(`%s`).ReplaceAllStringFunc(val, func(_ string) string {
			if argIdx >= len(args) {
				return ""
			}
			cur := args[argIdx]
			argIdx++
			return cur
		})
		v.Set(k, val)
	}
	return v
}

func splitCSVLike(s string) []string {
	var out []string
	var cur strings.Builder
	quote := byte(0)
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if quote == 0 && (ch == '\'' || ch == '"') {
			quote = ch
			cur.WriteByte(ch)
			continue
		}
		if quote != 0 && ch == quote {
			quote = 0
			cur.WriteByte(ch)
			continue
		}
		if quote == 0 && ch == ',' {
			out = append(out, strings.TrimSpace(cur.String()))
			cur.Reset()
			continue
		}
		cur.WriteByte(ch)
	}
	if strings.TrimSpace(cur.String()) != "" {
		out = append(out, strings.TrimSpace(cur.String()))
	}
	return out
}

func trimQuote(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "\"")
	s = strings.Trim(s, "'")
	return s
}
