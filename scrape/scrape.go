package scrape

import (
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
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		UpdateCrawlKeys(context, child)
	}

	crawlKey := util.GetCrawlKey(node)
	if crawlKey == "" {
		return
	}

	for i, a := range node.Attr {
		if a.Key == crawlKey {
			UpdateCrawlKey(context, &node.Attr[i])
			break
		}
	}
}

func UpdateCrawlKey(context ScrapeContext, a *html.Attribute) {
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
func ScrapeDocumentation(rawRootUrl string) {
	rootUrl := util.ValidateRawUrl(rawRootUrl)
	docDirectoryPath, err := CreateDocDirectory(rootUrl)

	if os.IsExist(err) {
		log.Printf("Documentation directory for %s already exists: %s\n", rawRootUrl, docDirectoryPath)
		log.Printf("If you intend to rescrape this endpoint, you must first delete this directory\n")
		return
	}

	log.Printf("\n")
	log.Printf("Scrape the Docs!\n")
	log.Printf("Version 1.0\n\n")
	log.Printf("Scraping documentation at %s to %s\n\n", rawRootUrl, docDirectoryPath)

	onVisit := ScrapeOnVisit(docDirectoryPath)
	crawler := crawl.NewCrawler(rootUrl, onVisit, 3, 5)
	crawler.Run()

	log.Printf("\n")
	log.Printf("Scraping complete!\n")
	log.Printf("Documentation root available at %s", filepath.Join(docDirectoryPath, "index.html"))
	log.Printf("\n")
}
