package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"matthewrubino.com/scrape/scrape"
)

var maxDepth = flag.Int("d", 3, "Specify max default depth of the crawler. Default is 3.")
var maxGoroutines = flag.Int("c", 3, "Specify max number of concurrent requests. Default is 3.")

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Must specify a root URL. Usage: scrape {ROOT_URL} -d {MAX_DEPTH} -c {MAX_CONCURRENT_REQUESTS}")
	}
	rootUrl := os.Args[1]

	// flag.Parse() parses os.Args[1:] until first nonflag argument
	// => need to remove one argument to force parsing
	os.Args = os.Args[1:]
	flag.Parse()
	fmt.Println(*maxDepth, *maxGoroutines)

	scrape.ScrapeDocumentation(rootUrl, *maxDepth, *maxGoroutines)
}
