package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"

	"github.com/rbeaudoin/gobot/crawler"
)

func main() {
	domain := flag.String("domain", "digitalocean.com", "The domain to crawl, defaults to 'digitalocean.com'")
	flag.Parse()

	// Start the crawl by visting '/' for the supplied domain
	url, err := url.Parse(fmt.Sprintf("http://%s/", *domain))
	if err != nil {
		log.Fatal(err)
	}

	sm, err := crawler.Crawl(*url)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", sm)
}
