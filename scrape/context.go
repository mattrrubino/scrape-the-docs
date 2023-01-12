package scrape

import (
	"net/url"

	"matthewrubino.com/scrape/crawl"
)

type ScrapeContext struct {
	crawl.PageContext
	crawler          *crawl.Crawler
	docDirectoryPath string
}

func (context *ScrapeContext) GetDocDirectoryPath() string {
	return context.docDirectoryPath
}

func (context *ScrapeContext) NextScrapeContext(url *url.URL) *ScrapeContext {
	nextContext := *context
	nextContext.PageContext = *context.NextPageContext(url)

	return &nextContext
}
