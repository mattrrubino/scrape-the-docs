package crawl

import (
	"bytes"
	"io"
	"net/url"
	"path/filepath"
	"strings"

	"matthewrubino.com/scrape/util"
)

// Container for all stateful information required for crawling one page
type PageContext struct {
	*url.URL            // Relative URL embedding
	depth        int    // Current crawling depth
	responseBody []byte // HTTP response body data
}

func (context *PageContext) GetDepth() int {
	return context.depth
}

func (context *PageContext) GetResponseBody() io.Reader {
	return bytes.NewReader(context.responseBody)
}

func (context *PageContext) SetResponseBody(bodyReader io.ReadCloser) {
	defer bodyReader.Close()

	// Read in body stream so it can be consumed multiple times
	responseBody, err := io.ReadAll(bodyReader)
	util.Check(err)

	context.responseBody = responseBody
}

func (context *PageContext) IsHtml() bool {
	urlString := context.String()
	ext := filepath.Ext(urlString)

	return ext == ".html" || ext == ""
}

func (context *PageContext) IsRel() bool {
	return !context.IsAbs() && !strings.HasPrefix(context.String(), "/")
}

func (context *PageContext) NextPageContext(href *url.URL) *PageContext {
	nextContext := *context
	nextContext.depth += 1

	if href.IsAbs() {
		nextContext.URL = href
	} else if strings.HasPrefix(href.String(), "/") {
		rawUrl, err := url.JoinPath(href.String()[1:])
		util.Check(err)

		nextContext.URL = util.ValidateRawUrl(rawUrl)
	} else {
		dirPath := filepath.Dir(context.String())
		rawUrl, err := url.JoinPath(dirPath, href.String())
		util.Check(err)

		nextContext.URL = util.ValidateRawUrl(rawUrl)
	}

	return &nextContext
}

func NewPageContext(rel string, depth int) *PageContext {
	context := PageContext{util.ValidateRawUrl(rel), depth, nil}
	return &context
}
