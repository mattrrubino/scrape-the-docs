package main

import (
	"matthewrubino.com/scrape/scrape"
	"matthewrubino.com/scrape/util"
)

func main() {
	util.Init()
	scrape.ScrapeDocumentation("http://localhost")
}
