package crawler

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"text/template"
)

const (
	// constants for paths, and test query str
	pg1 = "/"
	pg2 = "/foo"
	pg3 = "/bar"
	qs  = "?foo=bar&this=that"
)

var (
	// map of paths to expected data, links and assets must be sorted
	pgData = map[string]Page{
		pg1: {
			Path:  pg1,
			Links: []string{pg3 + qs, pg2},
			Assets: []string{
				"/css/doc1.css",
				"/images/doc1.jpg",
				"doc1.js",
				"http://images.com/doc1.jpg",
				"http://scripts.com/doc1.js",
			},
		},
		pg2: {
			Path:  pg2,
			Links: []string{pg1 + qs, pg3},
			Assets: []string{
				"/doc2/doc2.js",
				"/images/doc2.jpg",
				"doc2.css",
				"doc2.js",
			},
		},
		pg3: {
			Path:  pg3,
			Links: []string{pg2},
			Assets: []string{
				"/images/doc3.jpg",
				"doc3.css",
				"doc3.js",
			},
		},
	}

	// The doc tmpl for '/' requests. Tests embedded (non-first level)
	// links, and ignoring other domain links, and current path
	doc1Tmpl = `<!DOCTYPE html>
			<html lang="en">
				<head>
					<meta charset="utf-8">
					<title>doc1</title>
					<link rel="stylesheet" href="{{index .Assets 0}}">
					<script src="{{index .Assets 1}}"></script>
					<script src="{{index .Assets 2}}"></script>
				</head>
				<body>
					<a href="/">SELF</a>
					<a href="{{index .Links 0}}">FOO</a>
					<div><a href="{{index .Links 1}}">BAR</a></div>
					<a href="http://fake.com/users">FAKE</a>
					<img src="{{index .Assets 3}}">
					<img src="{{index .Assets 4}}">
				</body>
			</html>`

	// The document tmpl for '/foo' reqests. Tests query string,
	// hash URLs, multi level embedded link
	doc2Tmpl = `<!DOCTYPE html>
			<html lang="en">
				<head>
					<meta charset="utf-8">
					<title>doc2</title>
					<link rel="stylesheet" href="{{index .Assets 0}}">
					<script src="{{index .Assets 1}}"></script>
					<script src="{{index .Assets 2}}"></script>
				</head>
				<body>
					<a href="{{index .Links 0}}#test">BAR</a>
					<div><div><a href="{{index .Links 1}}">BASE</a></div></div>
					<img src="{{index .Assets 3}}">
				</body>
			</html>`

	// The document tmpl by for '/bar' requests. Tests skipping
	// of links already visited
	doc3Tmpl = `<!DOCTYPE html>
			<html lang="en">
				<head>
					<meta charset="utf-8">
					<title>doc3</title>
					<link rel="stylesheet" href="{{index .Assets 0}}">
					<script src="{{index .Assets 1}}"></script>
				</head>
				<body>
					<a href="{{index .Links 0}}">FOO</a>
					<img src="{{index .Assets 2}}">
				</body>
			</html>`
)

func TestAttrs(t *testing.T) {
	if len(attrs) != 4 {
		t.Fatalf("Expected 4 items in attribute map, found %d", len(attrs))
	}

	fn := func(tg string, at string) {
		if v, ok := attrs[tg]; !ok {
			t.Fatalf("The %s tag is missing an entry in the attribute map", tg)
		} else if v != at {
			t.Fatalf("The %s tag has value %s when value %s is expected in the attribute map", tg, v, at)
		}
	}

	fn(atag, href)
	fn(stag, src)
	fn(itag, src)
	fn(ltag, href)
}

func TestCrawl(t *testing.T) {
	url, _ := url.Parse("/test/users")
	if _, err := Crawl(*url); err == nil {
		t.Fatal("Expected error when omitting host from URL")
	}

	url, _ = url.Parse("http://test.com")
	if _, err := Crawl(*url); err == nil {
		t.Fatal("Expected error when omitting path from URL")
	}

	srv := mockSrvr()
	defer srv.Close()

	url, err := url.Parse(srv.URL + "/")
	if err != nil {
		t.Fatalf("Unable to parse URL from mock server %s", srv.URL)
	}
	sm, err := Crawl(*url)
	if err != nil {
		t.Fatal("Error when indexing test pages: " + err.Error())
	}
	for k := range pgData {
		fnd := false
		for x := range sm.Pages {
			if k == sm.Pages[x].Path {
				fnd = true
				if !slicesEqual(pgData[k].Links, sm.Pages[x].Links) {
					t.Fatal("Expected links do not match actual links", pgData[k].Links, "-->", sm.Pages[x].Links)
				} else if !slicesEqual(pgData[k].Assets, sm.Pages[x].Assets) {
					t.Fatal("Expected assets do not match actual assets", pgData[k].Assets, "-->", sm.Pages[x].Assets)
				}
				break
			}
		}
		if !fnd {
			t.Fatalf("No matching result found for expected url: %s", k)
		}
	}
}

func TestFetch(t *testing.T) {
	srv := mockSrvr()
	defer srv.Close()

	doc1, doc2, _ := getDocs()

	url, _ := url.Parse(srv.URL + "/")
	s, err := fetch(*url)
	if err != nil {
		t.Fatal("Call to 'fetch' returned error: ", err.Error())
	}
	if strings.TrimSpace(s) != strings.TrimSpace(doc1) {
		fmt.Println(len(s), " ", len(doc1))
		t.Fatal("Call to 'fetch' does not return expected value", s, doc1)
	}

	url, _ = url.Parse(srv.URL + "/foo")
	s, err = fetch(*url)
	if err != nil {
		t.Fatal("Call to 'fetch' returned error: ", err.Error())
	}
	if strings.TrimSpace(s) != strings.TrimSpace(doc2) {
		fmt.Println(len(s), " ", len(doc2))
		t.Fatal("Call to 'fetch' does not return expected value", s, doc2)
	}
}

func TestParse(t *testing.T) {
	base, _ := url.Parse("http://test.com/")
	_, doc2, _ := getDocs()
	ln, as := parse(doc2, *base)

	if !slicesEqual(pgData[pg2].Links, ln) {
		t.Fatal("Links do not match expected value", pgData[pg2].Links, "---->", ln)
	}

	if !slicesEqual(pgData[pg2].Assets, as) {
		t.Fatal("Assets do not match expected value", pgData[pg2].Assets, "---->", as)
	}

	// make sure the current path is excluded from links
	for _, v := range ln {
		if v == base.Path {
			t.Fatal("Links from base path should be excluded from parsed links")
		}
	}
}

// note: if pgData values change, this test must be updated
func TestSiteMapString(t *testing.T) {
	es := `Site map for http://test.com
Path: 
	/
	Links:
		/bar?foo=bar&this=that
		/foo
	Assets:
		/css/doc1.css
		/images/doc1.jpg
		doc1.js
		http://images.com/doc1.jpg
		http://scripts.com/doc1.js
Path: 
	/foo
	Links:
		/?foo=bar&this=that
		/bar
	Assets:
		/doc2/doc2.js
		/images/doc2.jpg
		doc2.css
		doc2.js
Path: 
	/bar
	Links:
		/foo
	Assets:
		/images/doc3.jpg
		doc3.css
		doc3.js
`
	sm := SiteMap{URL: "http://test.com"}
	sm.Pages = append(sm.Pages, pgData[pg1])
	sm.Pages = append(sm.Pages, pgData[pg2])
	sm.Pages = append(sm.Pages, pgData[pg3])

	s := sm.String()
	if es != s {
		t.Fatal("SiteMap String() does not match expected", s, "---->", es)
	}
}

// note: if pgData values change this test must be updated
func TestSitePageString(t *testing.T) {
	es1 := `Path: 
	/
	Links:
		/bar?foo=bar&this=that
		/foo
	Assets:
		/css/doc1.css
		/images/doc1.jpg
		doc1.js
		http://images.com/doc1.jpg
		http://scripts.com/doc1.js
`
	es2 := `Path: 
	/foo
	Links:
		/?foo=bar&this=that
		/bar
	Assets:
		/doc2/doc2.js
		/images/doc2.jpg
		doc2.css
		doc2.js
`
	es3 := `Path: 
	/bar
	Links:
		/foo
	Assets:
		/images/doc3.jpg
		doc3.css
		doc3.js
`

	s1 := pgData[pg1].String()
	s2 := pgData[pg2].String()
	s3 := pgData[pg3].String()

	if es1 != s1 {
		t.Fatal("SiteMap String() does not match expected", s1, "-->", es1)
	}

	if es2 != s2 {
		t.Fatal("SiteMap String() does not match expected", s2, "-->", es2)
	}

	if es3 != s3 {
		t.Fatal("SiteMap String() does not match expected", s3, "-->", es3)
	}
}

func slicesEqual(a []string, b []string) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Provides a mock server that serves requests for '/', '/foo', '/bar'
func mockSrvr() *httptest.Server {

	doc1, doc2, doc3 := getDocs()

	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		if r.URL.Path == pg1 {
			w.WriteHeader(200)
			fmt.Fprintln(w, doc1)
		} else if r.URL.Path == pg2 {
			w.WriteHeader(200)
			fmt.Fprintln(w, doc2)
		} else if r.URL.Path == pg3 {
			w.WriteHeader(200)
			fmt.Fprintln(w, doc3)
		} else {
			w.WriteHeader(404)
			fmt.Println(w, "Page not found")
		}
	}
	return httptest.NewServer(http.HandlerFunc(fn))
}

func getDocs() (doc1 string, doc2 string, doc3 string) {
	// anonymous fn for creating strings from tmpls
	fn := func(n string, t string, pg Page) string {
		var doc bytes.Buffer
		tmpl, err := template.New(n).Parse(t)
		if err != nil {
			panic(err)
		}
		err = tmpl.Execute(&doc, pg)
		if err != nil {
			panic(err)
		}
		return doc.String()
	}

	doc1 = fn("doc1", doc1Tmpl, pgData[pg1])
	doc2 = fn("doc2", doc2Tmpl, pgData[pg2])
	doc3 = fn("doc3", doc3Tmpl, pgData[pg3])
	return doc1, doc2, doc3
}
