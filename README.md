# Gobot

A web crawler written in [Go](https://golang.org/)

## Building and Running

Build gobot by running the `go build` command. Once built you can run:

`gobot crawl -domain foo.com`

to crawl the foo.com domain. The results are sent to stdout.

## Testing

To run tests, navigate to the root folder and run:

`go test ./...`

The output from the tests will be sent to the terminal.

## Assumptions

Gobot was built to do the following:

1. Crawl a single domain specified by the user, and do not follow links to subdomains
2. Maintain a collection of links (as identified by the 'href' attribute of the 'a' tag), and static assets (as identified by the 'link'script', and 'img' tags)
3. Output this information to stdout, with each crawled path and related links/static assets
4. Remove the hash fragment (i.e. everything after '#') when crawling pages
5. Consider query parameters (i.e. everything after '?') as identifying a unique URL. For example, /foo and /foo?bar=baz will both be crawled and considered unique URLs

## Enhancements

Gobot has some remove from improvement. The following are a collection of features that would improve Gobot:

1. Add benchmarks to the existing tests to maintain perf
2. Add aditional tests for edge cases, and improve test coverage
3. Add feature to honor 'robots.txt'
4. Deal with throttling better; currently just log an error

Ideas and thoughts are always welcome

## License

MIT
