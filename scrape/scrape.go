package scrape

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"matthewrubino.com/scrape/crawl"
	"matthewrubino.com/scrape/util"

	"golang.org/x/net/html"
)

func CreateDocDirectory(rootUrl *url.URL) (string, error) {
	subdir := rootUrl.Hostname() + rootUrl.EscapedPath() + "/"
	docDirectoryPath := filepath.Join("documentation", util.SafeFilepath(subdir))

	if _, err := os.Stat(docDirectoryPath); !os.IsNotExist(err) {
		return docDirectoryPath, os.ErrExist
	}

	err := os.MkdirAll(docDirectoryPath, os.ModeDir)
	util.Check(err)

	return docDirectoryPath, nil
}

func GetLocalFilepath(context ScrapeContext) string {
	safePath := util.SafeFilepath(context.String())

	basePath := context.GetDocDirectoryPath()
	if context.IsAbs() {
		return filepath.Join(basePath, "__abs__", safePath)
	} else {
		return filepath.Join(basePath, safePath)
	}
}

func GetRelativeFilepath(context ScrapeContext, targetContext ScrapeContext) string {
	localFilepath := GetLocalFilepath(context)
	localDirpath := filepath.Dir(localFilepath)
	targetFilepath := GetLocalFilepath(targetContext)

	rel, err := filepath.Rel(localDirpath, targetFilepath)
	util.Check(err)

	return rel
}

func ShouldUpdateCrawlKeys(context ScrapeContext) bool {
	return context.crawler.ShouldCrawl(context.PageContext)
}

func UpdateCrawlKeys(context ScrapeContext, node *html.Node) {
	if node.Data == "body" {
		InjectJavaScript(context, node)
	}

	crawlAttr := util.GetCrawlAttr(node)
	UpdateCrawlKey(context, crawlAttr)

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		UpdateCrawlKeys(context, child)
	}
}

func UpdateCrawlKey(context ScrapeContext, a *html.Attribute) {
	if a == nil {
		return
	}

	url := util.ValidateRawUrl(a.Val)
	targetContext := context.NextScrapeContext(url)
	a.Val = GetRelativeFilepath(context, *targetContext)
}

func WriteScrapeContextToFile(context ScrapeContext, file *os.File) {
	body := context.GetResponseBody()

	if context.IsHtml() {
		doc := util.ValidateRawHtml(body)

		if ShouldUpdateCrawlKeys(context) {
			UpdateCrawlKeys(context, doc)
		}

		err := html.Render(file, doc)
		util.Check(err)
	} else {
		_, err := io.Copy(file, body)
		util.Check(err)
	}
}

func InjectJavaScript(context ScrapeContext, node *html.Node) {
	child := &html.Node{
		Type: html.ElementNode,
		Data: "script",
		Attr: []html.Attribute{
			{Key: "src", Val: "/__inject__.js"},
			{Key: "type", Val: "text/javascript"},
			{Key: "defer"},
		},
	}

	node.AppendChild(child)
}

func CreateInjectJavaScriptFile(docDirectoryPath string) (*os.File, error) {
	path := filepath.Join(docDirectoryPath, "__inject__.js")
	return os.Create(path)
}

func ScrapeOnVisit(docDirectoryPath string) func(*crawl.Crawler, crawl.PageContext) {
	onVisit := func(crawler *crawl.Crawler, context crawl.PageContext) {
		log.Printf("Scraping %s\n", crawler.GetUrlString(context))

		scrapeContext := ScrapeContext{context, crawler, docDirectoryPath}
		localFilepath := GetLocalFilepath(scrapeContext)

		directories := filepath.Dir(localFilepath)
		err := os.MkdirAll(directories, os.ModeDir)
		util.Check(err)

		file, err := os.Create(localFilepath)
		util.Check(err)

		WriteScrapeContextToFile(scrapeContext, file)
	}

	return onVisit
}

// Scrapes documentation for the input documentation root.
// Function panics if the input string is a malformed URL
// or if the supplied URL does not respond.
func ScrapeDocumentation(rawRootUrl string, maxDepth int, maxGoroutines int) {
	rootUrl := util.ValidateRawUrl(rawRootUrl)
	docDirectoryPath, err := CreateDocDirectory(rootUrl)

	fmt.Println()
	fmt.Println("Scrape the Docs!")
	fmt.Println("Version 1.0")

	if os.IsExist(err) {
		fmt.Printf("Documentation directory for %s already exists: %s\n", rawRootUrl, docDirectoryPath)
		fmt.Println("If you intend to rescrape this endpoint, you must first delete this directory")
		return
	}

	fmt.Printf("Scraping documentation at %s to %s\n\n", rawRootUrl, docDirectoryPath)
	fmt.Println("Creating __inject__.js")

	_, err = CreateInjectJavaScriptFile(docDirectoryPath)
	util.Check(err)

	onVisit := ScrapeOnVisit(docDirectoryPath)
	crawler := crawl.NewCrawler(rootUrl, onVisit, maxDepth, maxGoroutines)
	crawler.Run()

	fmt.Println()
	fmt.Println("Scraping complete!")
	fmt.Printf("Documentation root available at %s", filepath.Join(docDirectoryPath, "index.html"))
	fmt.Println()
}
