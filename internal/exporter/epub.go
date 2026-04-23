package exporter

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opso-code/sonovel-go/internal/model"
	"github.com/opso-code/sonovel-go/internal/util"
)

type EPUBExporter struct{}

func (e *EPUBExporter) Export(meta model.BookMeta, chapters []model.ChapterItem, outputDir string) (string, error) {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", err
	}
	name := fmt.Sprintf("%s(%s).epub", util.SanitizeName(meta.BookName), util.SanitizeName(meta.Author))
	path := filepath.Join(outputDir, name)

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	if err := writeStored(zw, "mimetype", []byte("application/epub+zip")); err != nil {
		return "", err
	}
	if err := writeFile(zw, "META-INF/container.xml", []byte(containerXML)); err != nil {
		return "", err
	}
	if err := writeFile(zw, "OEBPS/style.css", []byte(defaultCSS)); err != nil {
		return "", err
	}

	manifest := make([]string, 0, len(chapters)+2)
	spine := make([]string, 0, len(chapters))
	for _, ch := range chapters {
		id := fmt.Sprintf("ch%04d", ch.Order)
		fname := fmt.Sprintf("chapter-%04d.xhtml", ch.Order)
		xhtml := renderXHTML(ch.Title, ch.Content)
		if err := writeFile(zw, "OEBPS/"+fname, []byte(xhtml)); err != nil {
			return "", err
		}
		manifest = append(manifest, fmt.Sprintf(`<item id="%s" href="%s" media-type="application/xhtml+xml"/>`, id, fname))
		spine = append(spine, fmt.Sprintf(`<itemref idref="%s"/>`, id))
	}

	opf := fmt.Sprintf(contentOPFTemplate,
		escapeXML(meta.BookName),
		escapeXML(meta.Author),
		strings.Join(manifest, "\n    "),
		strings.Join(spine, "\n    "))
	if err := writeFile(zw, "OEBPS/content.opf", []byte(opf)); err != nil {
		return "", err
	}
	return path, nil
}

func writeStored(zw *zip.Writer, name string, data []byte) error {
	h := &zip.FileHeader{Name: name, Method: zip.Store}
	w, err := zw.CreateHeader(h)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func writeFile(zw *zip.Writer, name string, data []byte) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func renderXHTML(title, content string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" lang="zh-CN"><head>
<meta charset="utf-8"/><title>%s</title><link rel="stylesheet" type="text/css" href="style.css"/>
</head><body><h1>%s</h1>%s</body></html>`, escapeXML(title), escapeXML(title), content)
}

func escapeXML(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&apos;")
	return r.Replace(s)
}

const containerXML = `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`

const contentOPFTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="BookID" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="BookID">sonovel-go</dc:identifier>
    <dc:title>%s</dc:title>
    <dc:creator>%s</dc:creator>
    <dc:language>zh-CN</dc:language>
  </metadata>
  <manifest>
    <item id="css" href="style.css" media-type="text/css"/>
    %s
  </manifest>
  <spine>
    %s
  </spine>
</package>`

const defaultCSS = `body{font-family: serif;line-height:1.7;} h1{font-size:1.3em;}`
