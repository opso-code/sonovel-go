package tui

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/opso-code/sonovel-go/internal/model"
)

func PrintSearchResultsTable(results []model.SearchResult) {
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
	if unicode.Is(unicode.Han, r) || unicode.Is(unicode.Hangul, r) || unicode.Is(unicode.Hiragana, r) || unicode.Is(unicode.Katakana, r) {
		return 2
	}
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
