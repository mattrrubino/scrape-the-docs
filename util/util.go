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

func SafeFilepath(path string) string {
	re, err := regexp.Compile("[<>:\"|?*]")
	Check(err)

	safePath := re.ReplaceAllString(path, "%")
	if safePath != "" && filepath.Ext(safePath) == "" && !strings.HasSuffix(safePath, "/") {
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
