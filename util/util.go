package util

import (
	"io"
	"log"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/h2non/filetype"
	"golang.org/x/net/html"
)

func init() {
	filetype.AddType("html", "text/html")
	filetype.AddType("htm", "text/html")
	filetype.AddType("csv", "text/csv")
	filetype.AddType("js", "text/javascript")
	filetype.AddType("mjs", "text/javascript")
	filetype.AddType("css", "text/css")
	filetype.AddType("txt", "text/plain")
	filetype.AddType("xhtml", "application/xhtml+xml")
	filetype.AddType("xml", "application/xml")
	filetype.AddType("zip", "application/zip")
	filetype.AddType("json", "application/json")
	filetype.AddType("jpeg", "image/jpeg")
	filetype.AddType("svg", "image/svg+xml")
}

func Check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func IsDir(path string) bool {
	return path == "" || strings.HasSuffix(path, "/")
}

func KnownFiletype(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))

	// Remove period to match extension format used by filetype package
	if ext != "" {
		ext = ext[1:]
	}
	ftype := filetype.GetType(ext)

	return ftype != filetype.Unknown || IsDir(path)
}

func SafeFilepath(path string) string {
	re, err := regexp.Compile("[<>:\"|?*]")
	Check(err)

	noQuery := strings.Split(path, "?")[0]
	safePath := re.ReplaceAllString(noQuery, "%")

	// Need explicit .html extension in local filesystem
	if !KnownFiletype(safePath) {
		safePath += ".html"
	}

	return safePath
}

func ValidateRawUrl(rawUrl string) *url.URL {
	parsedUrl, err := url.Parse(rawUrl)
	Check(err)

	return parsedUrl
}

func ValidateRawHtml(rawHtml io.Reader) *html.Node {
	node, err := html.Parse(rawHtml)
	Check(err)

	return node
}

func GetCrawlAttr(node *html.Node) *html.Attribute {
	crawlKey := ""

	if node.Data == "link" || node.Data == "a" {
		crawlKey = "href"
	} else if node.Data == "script" {
		crawlKey = "src"
	}

	for i, a := range node.Attr {
		if a.Key == crawlKey {
			return &node.Attr[i]
		}
	}

	return nil
}
