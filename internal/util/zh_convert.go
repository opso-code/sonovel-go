package util

import (
	"strings"
	"github.com/opso-code/sonovel-go/internal/model"
)

var (
	simpToTradMap map[rune]rune
	tradToSimpMap map[rune]rune
)

type ChineseConverter struct{}

func NewChineseConverter() *ChineseConverter {
	return &ChineseConverter{}
}

func (c *ChineseConverter) Convert(text, sourceLang, targetLang string) string {
	if text == "" || sourceLang == targetLang {
		return text
	}

	sourceLang = normalizeLang(sourceLang)
	targetLang = normalizeLang(targetLang)

	convertFunc := getConversionFunction(sourceLang, targetLang)
	if convertFunc == nil {
		return text
	}

	return convertFunc(text)
}

func (c *ChineseConverter) ConvertBook(book *model.Book, sourceLang, targetLang string) *model.Book {
	if book == nil || sourceLang == targetLang {
		return book
	}

	convertFunc := getConversionFunction(sourceLang, targetLang)
	if convertFunc == nil {
		return book
	}

	if book.BookName != "" {
		book.BookName = c.convert(book.BookName, convertFunc)
	}
	if book.Author != "" {
		book.Author = c.convert(book.Author, convertFunc)
	}
	if book.Intro != "" {
		book.Intro = c.convert(book.Intro, convertFunc)
	}
	if book.Category != "" {
		book.Category = c.convert(book.Category, convertFunc)
	}
	if book.LatestChapter != "" {
		book.LatestChapter = c.convert(book.LatestChapter, convertFunc)
	}
	if book.LastUpdateTime != "" {
		book.LastUpdateTime = c.convert(book.LastUpdateTime, convertFunc)
	}
	if book.Status != "" {
		book.Status = c.convert(book.Status, convertFunc)
	}

	return book
}

func (c *ChineseConverter) ConvertChapter(chapter *model.Chapter, sourceLang, targetLang string) *model.Chapter {
	if chapter == nil || sourceLang == targetLang {
		return chapter
	}

	convertFunc := getConversionFunction(sourceLang, targetLang)
	if convertFunc == nil {
		return chapter
	}

	if chapter.Title != "" {
		chapter.Title = c.convert(chapter.Title, convertFunc)
	}
	if chapter.Content != "" {
		chapter.Content = c.convert(chapter.Content, convertFunc)
	}

	return chapter
}

func normalizeLang(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	switch lang {
	case "zh-cn", "zhcn", "简体中文", "简中":
		return "zh-cn"
	case "zh-tw", "zhtw", "zh-hant", "繁体中文", "繁中":
		return "zh-tw"
	default:
		return lang
	}
}

func getConversionFunction(sourceLang, targetLang string) func(string) string {
	switch sourceLang + ">" + targetLang {
	case "zh-cn>zh-tw", "zh-hans>zh-tw", "zhcn>zhtw":
		return c2t
	case "zh-tw>zh-cn", "zh-hant>zh-cn", "zhtw>zhcn":
		return t2c
	default:
		return nil
	}
}

func c2t(text string) string {
	result := make([]rune, 0, len(text))
	for _, r := range text {
		if trad, ok := simpToTradMap[r]; ok {
			result = append(result, trad)
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

func t2c(text string) string {
	result := make([]rune, 0, len(text))
	for _, r := range text {
		if simp, ok := tradToSimpMap[r]; ok {
			result = append(result, simp)
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

func (c *ChineseConverter) convert(text string, convertFunc func(string) string) string {
	if convertFunc == nil {
		return text
	}
	return convertFunc(text)
}

func buildSimpToTradMap() map[rune]rune {
	m := make(map[rune]rune, 2000)
	return m
}

func buildTradToSimpMap() map[rune]rune {
	m := make(map[rune]rune, 2000)
	return m
}
