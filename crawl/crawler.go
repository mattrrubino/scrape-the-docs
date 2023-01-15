package crawl

import (
	"net/http"
	"net/url"
	"strings"
	"sync"

	"matthewrubino.com/scrape/util"

	"golang.org/x/net/html"
)

// Container for all crawler information required for crawling all pages
type Crawler struct {
	*url.URL                                  // Root URL embedding
	visited       *SafeSet[string]            // Set tracking visited URL strings
	wg            *sync.WaitGroup             // Wait group for tracking goroutine execution
	onVisit       func(*Crawler, PageContext) // Function to perform on each crawled page
	maxDepth      int                         // Max crawl depth
	maxGoroutines int                         // Maximum number of concurrent goroutines used to crawl
	resource      chan int                    // Channel used for regulating number of concurrent goroutines
}

func PreprocessRootUrl(rootUrl *url.URL) *url.URL {
	rootUrlString := rootUrl.String()

	// Crawler requires root URL to specify a directory (indicated by the suffix "/")
	if !strings.HasSuffix(rootUrlString, "/") {
		rootUrlString += "/"
	}

	newRootUrl, err := url.Parse(rootUrlString)
	util.Check(err)

	return newRootUrl
}

func NewCrawler(rootUrl *url.URL, onVisit func(*Crawler, PageContext), maxDepth int, maxGoroutines int) *Crawler {
	visited := NewSafeSet[string]()
	wg := &sync.WaitGroup{}
	resource := make(chan int, maxGoroutines)

	return &Crawler{PreprocessRootUrl(rootUrl), visited, wg, onVisit, maxDepth, maxGoroutines, resource}
}

func (crawler *Crawler) Run() *Crawler {
	context := NewPageContext("index.html", 0)
	crawler.VisitPageContext(*context)
	crawler.WaitForGoroutines()

	return crawler
}

func (crawler *Crawler) OnVisit(context PageContext) {
	crawler.onVisit(crawler, context)
}

func (crawler *Crawler) GetUrlString(context PageContext) string {
	crawlerString := crawler.String()
	contextString := context.String()

	if context.IsAbs() {
		return contextString
	}

	urlString, err := url.JoinPath(crawlerString, contextString)
	util.Check(err)

	return urlString
}

func (crawler *Crawler) GetMaxDepth() int {
	return crawler.maxDepth
}

func (crawler *Crawler) Claim(context PageContext) bool {
	urlString := context.String()
	return !crawler.visited.Add(urlString)
}

func (crawler *Crawler) WaitForGoroutines() {
	crawler.wg.Wait()
}

func (crawler *Crawler) OpenGoroutine() {
	crawler.wg.Add(1)
}

func (crawler *Crawler) WaitForResource() {
	crawler.resource <- 0
}

func (crawler *Crawler) CloseGoroutine() {
	<-crawler.resource
	crawler.wg.Done()
}

func (crawler *Crawler) VisitPageContext(context PageContext) {
	if !crawler.ShouldVisit(context) {
		return
	}

	crawler.OpenGoroutine()
	go crawler.VisitPageContextGoroutine(context)
}

func (crawler *Crawler) RequestPageContext(context *PageContext) {
	urlString := crawler.GetUrlString(*context)
	resp, err := http.Get(urlString)
	util.Check(err)

	context.SetResponseBody(resp.Body)
}

func (crawler *Crawler) VisitPageContextGoroutine(context PageContext) {
	defer crawler.CloseGoroutine()

	crawler.WaitForResource()
	crawler.RequestPageContext(&context)
	crawler.OnVisit(context)

	if crawler.ShouldCrawl(context) {
		crawler.CrawlPageContext(context)
	}
}

func (crawler *Crawler) CrawlPageContext(context PageContext) {
	body := context.GetResponseBody()
	node := util.ValidateRawHtml(body)

	crawler.CrawlHtmlPageContext(context, node)
}

func (crawler *Crawler) ExceedsMaxDepth(context PageContext) bool {
	maxDepth := crawler.GetMaxDepth()
	if !context.IsHtml() {
		maxDepth += 1
	}

	return context.GetDepth() > maxDepth
}

func (crawler *Crawler) EscapesRootUrl(context PageContext) bool {
	crawlerString := crawler.String()
	contextString := context.String()

	urlString, err := url.JoinPath(crawlerString, contextString)
	util.Check(err)

	return !strings.Contains(urlString, crawlerString)
}

func (crawler *Crawler) IsDir(context PageContext) bool {
	hrefString := context.String()
	return hrefString == "" || strings.HasSuffix(hrefString, ".") || strings.HasSuffix(hrefString, "/")
}

func (crawler *Crawler) ShouldVisit(context PageContext) bool {
	if crawler.ExceedsMaxDepth(context) {
		return false
	}

	if !context.IsAbs() && (crawler.IsDir(context) || crawler.EscapesRootUrl(context)) {
		return false
	}

	return crawler.Claim(context)
}

func (crawler *Crawler) ShouldCrawl(context PageContext) bool {
	return !context.IsAbs() && context.IsHtml() && !crawler.ExceedsMaxDepth(context)
}

func (crawler *Crawler) CrawlHtmlPageContext(context PageContext, node *html.Node) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		crawler.CrawlHtmlPageContext(context, child)
	}

	crawlKey := util.GetCrawlKey(node)
	if crawlKey == "" {
		return
	}

	for _, a := range node.Attr {
		if a.Key == crawlKey {
			url := util.ValidateRawUrl(a.Val)
			newContext := context.NextPageContext(url)
			crawler.VisitPageContext(*newContext)

			break
		}
	}
}
