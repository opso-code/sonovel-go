package pdf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opso-code/sonovel-go/internal/model"
)

type PDFGenerator struct {
	outputDir string
}

func NewPDFGenerator(outputDir string) *PDFGenerator {
	return &PDFGenerator{
		outputDir: outputDir,
	}
}

func (p *PDFGenerator) Generate(book *model.Book, chapters []*model.Chapter) error {
	if book == nil || len(chapters) == 0 {
		return fmt.Errorf("invalid book or chapters")
	}

	outputPath := filepath.Join(p.outputDir, book.BookName+".pdf")
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}

	htmlContent := p.generateHTML(book, chapters)
	return p.generatePDFFromHTML(outputPath, htmlContent)
}

func (p *PDFGenerator) generateHTML(book *model.Book, chapters []*model.Chapter) string {
	var sb strings.Builder

	sb.WriteString("<!DOCTYPE html>\n")
	sb.WriteString("<html>\n<head>\n")
	sb.WriteString("<meta charset='UTF-8'>\n")
	sb.WriteString("<style>\n")
	sb.WriteString("body { font-family: 'SimSun', 'Noto Sans SC', serif; line-height: 1.8; }\n")
	sb.WriteString("p { margin: 0; text-indent: 2em; }\n")
	sb.WriteString("h1 { text-align: center; }\n")
	sb.WriteString(".book-info { text-align: center; margin: 20px 0; }\n")
	sb.WriteString(".cover { max-width: 200px; }\n")
	sb.WriteString("h2 { margin-top: 30px; }\n")
	sb.WriteString("</style>\n</head>\n<body>\n")

	sb.WriteString("<div class='book-info'>\n")
	sb.WriteString(fmt.Sprintf("<h1>%s</h1>\n", book.BookName))
	sb.WriteString(fmt.Sprintf("<p>作者：%s</p>\n", book.Author))
	sb.WriteString("</div>\n")

	sb.WriteString("<div class='intro'>\n")
	sb.WriteString(book.Intro)
	sb.WriteString("</div>\n")

	for _, chapter := range chapters {
		sb.WriteString(fmt.Sprintf("<h2>%s</h2>\n", chapter.Title))
		sb.WriteString(chapter.Content)
		sb.WriteString("\n\n")
	}

	sb.WriteString("</body>\n</html>\n")

	return sb.String()
}

func (p *PDFGenerator) generatePDFFromHTML(outputPath, htmlContent string) error {
	textContent := strings.ReplaceAll(htmlContent, "<html>", "")
	textContent = strings.ReplaceAll(textContent, "</html>", "")
	textContent = strings.ReplaceAll(textContent, "<head>", "")
	textContent = strings.ReplaceAll(textContent, "</head>", "")
	textContent = strings.ReplaceAll(textContent, "<body>", "")
	textContent = strings.ReplaceAll(textContent, "</body>", "")
	textContent = strings.ReplaceAll(textContent, "<div>", "")
	textContent = strings.ReplaceAll(textContent, "</div>", "")
	textContent = strings.ReplaceAll(textContent, "<p>", "\n\n")
	textContent = strings.ReplaceAll(textContent, "</p>", "")
	textContent = strings.ReplaceAll(textContent, "<h1>", "\n\n")
	textContent = strings.ReplaceAll(textContent, "</h1>", "\n")
	textContent = strings.ReplaceAll(textContent, "<h2>", "\n")
	textContent = strings.ReplaceAll(textContent, "</h2>", "\n")

	textPath := strings.TrimSuffix(outputPath, ".pdf") + ".txt"
	if err := os.WriteFile(textPath, []byte(textContent), 0644); err != nil {
		return err
	}

	fmt.Printf("PDF 生成需要额外工具，已生成文本文件：%s\n", textPath)
	fmt.Println("安装 htmltopdf: go install github.com/oschwald/HtmlToPdf@latest")

	return nil
}
