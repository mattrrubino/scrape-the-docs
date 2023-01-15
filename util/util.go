package util

import (
	"io"
	"log"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

func Check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func IsDir(path string) bool {
	return path == "" || strings.HasSuffix(path, "/")
}

// TODO: Use filetype package
// TODO: Fix links not gaining .html extension (despite being gained by local files)
// TODO: Fix crawling of absolute contexts (or verify that it didn't happen)
func ShouldAddHtmlExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	knownExt := map[string]bool{
		".js":    true,
		".css":   true,
		".html":  true,
		".xhtml": true,
		".xml":   true,
		".csv":   true,
		".gif":   true,
		".jpg":   true,
		".jpeg":  true,
		".png":   true,
	}

	return !IsDir(path) && !knownExt[ext]
}

func SafeFilepath(path string) string {
	re, err := regexp.Compile("[<>:\"|?*]")
	Check(err)

	noQuery := strings.Split(path, "?")[0]
	safePath := re.ReplaceAllString(noQuery, "%")
	if ShouldAddHtmlExtension(safePath) {
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

func GetCrawlKey(node *html.Node) string {
	crawlKey := ""

	if node.Data == "link" || node.Data == "a" {
		crawlKey = "href"
	} else if node.Data == "script" {
		crawlKey = "src"
	}

	return crawlKey
}
