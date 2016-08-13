// Package crawler implments routines for crawling a domain
package crawler

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

const (
	atag = "a"
	stag = "script"
	itag = "img"
	ltag = "link"
	href = "href"
	src  = "src"
)

// SiteMap represents a crawled site/domain
type SiteMap struct {
	URL   string
	Pages []Page
}

func (sm SiteMap) String() string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("Site map for %s\n", sm.URL))
	for _, v := range sm.Pages {
		buffer.WriteString(fmt.Sprintf("%s", v))
	}

	return buffer.String()
}

// Page represents a resource that has been crawled
type Page struct {
	Path   string
	Links  []string
	Assets []string
}

func (pg Page) String() string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("Path: \n\t%s\n\tLinks:\n", pg.Path))
	for _, l := range pg.Links {
		buffer.WriteString(fmt.Sprintf("\t\t%s\n", l))
	}

	buffer.WriteString(fmt.Sprintf("\tAssets:\n"))
	for _, a := range pg.Assets {
		buffer.WriteString(fmt.Sprintf("\t\t%s\n", a))
	}

	return buffer.String()
}

type linkCache struct {
	links map[string]struct{}
	mux   sync.RWMutex
}

var (
	lc = linkCache{
		links: make(map[string]struct{}),
	}
	attrs = map[string]string{
		atag: href,
		stag: src,
		itag: src,
		ltag: href,
	}
)

// Crawl crawls all links in the domain of the supplied URL,
// and returns a collection of Page structs containing links
// and static assets for each page that was crawled
func Crawl(url url.URL) (sm SiteMap, err error) {
	if url.Host == "" {
		return sm, errors.New("crawler: host missing from URL")
	} else if url.Path == "" {
		return sm, errors.New("crawler: path missing from URL ")
	}

	ch := make(chan SiteMap)

	go crawl(url, ch)

	sm = <-ch
	log.Printf("Finished crawling URL: %s\n", url.String())

	return sm, err
}

func crawl(url url.URL, ch chan<- SiteMap) {
	var sm SiteMap

	// make sure this link hasn't been crawled
	lc.mux.Lock()
	_, ok := lc.links[url.RequestURI()]

	if !ok {
		lc.links[url.RequestURI()] = struct{}{}
		lc.mux.Unlock()
		log.Printf("Crawling URL: %s\n", url.String())

		b, err := fetch(url)
		if err != nil {
			log.Printf("Error fetching page for URL: %s, err: %s\n", url.String(), err)
			ch <- SiteMap{}
		}

		ln, as := parse(b, url)

		// create new sitemap and add indexed page
		sm = SiteMap{URL: url.String()}
		pg := Page{Path: url.Path, Links: ln, Assets: as}
		sm.Pages = append(sm.Pages, pg)

		c := make(chan SiteMap)
		chLnks := 0
		for _, v := range ln {
			chURL, err := url.Parse(v)
			if err != nil {
				log.Printf("Unable parse url for child link %s, error: %s", v, err)
				continue
			}
			go crawl(*chURL, c)
			chLnks++
		}
		// results of child link crawls
		for i := 0; i < chLnks; i++ {
			chsm := <-c
			if len(chsm.Pages) > 0 {
				sm.Pages = append(sm.Pages, chsm.Pages...)
			}
		}
	} else {
		lc.mux.Unlock()
	}
	// send a SiteMap, even if empty
	ch <- sm
}

func fetch(url url.URL) (string, error) {
	resp, err := http.Get(url.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func parse(s string, base url.URL) (ln []string, as []string) {
	z := html.NewTokenizer(strings.NewReader(s))
	lnm := make(map[string]struct{})
	asm := make(map[string]struct{})

	// anonymous func used to get attribute values
	attr := func(a string) string {
		var av string
		for {
			if k, v, ha := z.TagAttr(); string(k) == a {
				av = string(v)
				break
			} else if ha == false {
				break
			}
		}
		return av
	}

	// convert map to slice
	slc := func(m map[string]struct{}) []string {
		var v []string

		for k := range m {
			v = append(v, k)
		}
		sort.Strings(v)
		return v
	}

	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return slc(lnm), slc(asm)
		case html.StartTagToken, html.EndTagToken:
			if tn, ha := z.TagName(); ha {
				tg := string(tn)
				if av := attr(attrs[tg]); av != "" {
					switch tg {
					case atag:
						if url, err := base.Parse(av); err == nil {
							if url.Host == base.Host && url.RequestURI() != base.RequestURI() {
								lnm[url.RequestURI()] = struct{}{}
							}
						}
					case stag, itag, ltag:
						asm[av] = struct{}{}
					}
				}
			}
		}
	}
}
