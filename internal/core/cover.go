package core

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type CoverUpdater struct {
	httpClient *http.Client
	baseURL    string
}

func NewCoverUpdater() *CoverUpdater {
	return &CoverUpdater{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    "https://bookcover.yuewen.com/qdbimg/no-cover",
	}
}

func (c *CoverUpdater) FetchCover(bookName, bookAuthor, coverUrl string) string {
	if coverUrl == "" {
		coverUrl = c.baseURL
	}

	if isValidURL(coverUrl) {
		resp, err := c.httpClient.Get(coverUrl)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return coverUrl
		}
	}

	sources := []struct {
		name  string
		fetch func(bookName, bookAuthor string) (string, error)
	}{
		{"起点中文网", c.fetchQidian},
		{"纵横中文网", c.fetchZongheng},
		{"七猫小说网", c.fetchQimao},
	}

	bestCover := coverUrl
	bestScore := 0

	for _, source := range sources {
		cover, err := source.fetch(bookName, bookAuthor)
		if err == nil && cover != "" {
			if score := c.validateCover(cover); score > bestScore {
				bestScore = score
				bestCover = cover
			}
		}
	}

	return bestCover
}

func (c *CoverUpdater) fetchQidian(bookName, bookAuthor string) (string, error) {
	urlStr := fmt.Sprintf("https://www.qidian.com/so/%s.html", url.QueryEscape(bookName))
	resp, err := c.httpClient.Get(urlStr)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("起点请求失败：%s", resp.Status)
	}

	html, _ := io.ReadAll(resp.Body)
	return extractQidianCover(html, bookName, bookAuthor), nil
}

func (c *CoverUpdater) fetchZongheng(bookName, bookAuthor string) (string, error) {
	formData := url.Values{}
	formData.Set("keyword", bookName)
	formData.Set("pageNo", "1")
	formData.Set("pageNum", "20")
	formData.Set("isFromHuayu", "0")

	resp, err := c.httpClient.PostForm("https://search.zongheng.com/search/book", formData)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("纵横请求失败：%s", resp.Status)
	}

	return extractZonghengCover(resp.Body), nil
}

func (c *CoverUpdater) fetchQimao(bookName, bookAuthor string) (string, error) {
	formData := url.Values{}
	formData.Set("keyword", bookName)
	formData.Set("count", "0")
	formData.Set("page", "1")
	formData.Set("page_size", "15")

	resp, err := c.httpClient.PostForm("https://www.qimao.com/qimaoapi/api/search/result", formData)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("七猫请求失败：%s", resp.Status)
	}

	return extractQimaoCover(resp.Body), nil
}

func (c *CoverUpdater) validateCover(coverUrl string) int {
	if !isValidURL(coverUrl) {
		return 0
	}

	resp, err := c.httpClient.Head(coverUrl)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0
	}

	return 1
}

func isValidURL(urlStr string) bool {
	if urlStr == "" {
		return false
	}
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return false
	}
	return true
}

func extractQidianCover(html []byte, bookName, bookAuthor string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(html)))
	if err != nil {
		return ""
	}

	if cover, ok := doc.Find(".book-info img.cover").First().Attr("src"); ok && cover != "" {
		return cover
	}
	if cover, ok := doc.Find(".book-cover img").First().Attr("src"); ok && cover != "" {
		return cover
	}
	if cover, ok := doc.Find(".book-img img").First().Attr("src"); ok && cover != "" {
		return cover
	}
	return ""
}

func extractZonghengCover(body io.Reader) string {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return ""
	}

	if cover, ok := doc.Find(".book-cover img").First().Attr("src"); ok && cover != "" {
		return cover
	}
	if cover, ok := doc.Find(".book-img img").First().Attr("src"); ok && cover != "" {
		return cover
	}
	return ""
}

func extractQimaoCover(body io.Reader) string {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return ""
	}

	if cover, ok := doc.Find(".book-cover img").First().Attr("src"); ok && cover != "" {
		return cover
	}
	if cover, ok := doc.Find(".book-img img").First().Attr("src"); ok && cover != "" {
		return cover
	}
	return ""
}
