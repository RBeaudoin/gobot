# Gobot

A web crawler written in [Go](https://golang.org/)

## Running

Step one is to build Gobot from the source by following these steps:

1. Take a moment to walkthrough the [Getting Started](https://golang.org/doc/install) to install Go into your local development environment (Gobot was built/tested using Go v1.6)
3. Unzip the source to your local [GOPATH](https://github.com/golang/go/wiki/GOPATH) by running the command `tar -xzvf gobot.tar.gz -C GOPATH`
4. Navigate to the root source folder and run the command `go test ./...``

Once you have built and tested Gobot, it's ready to run.

Gobot accepts a single commandline parameter: the domain you'd like to crawl. The default value is 'digitalocean.com', but you can override this value by running:

`./gobot -domain <insert_your_domain_here>`

Gobot sends all of it's output to stdout, which can then be redirected to a file:

`./gobot -domain <insert_your_domain_here> > <output_file_name>`

Gobot logs info messages to stderr. You can redirect this output by running the following command:

`/gobot -domain <insert_your_domain_here> 2> <log_file_name>`

Finally, to send both log output and results to the same file, runt this command:

`/gobot -domain <insert_your_domain_here> 2> <log_file_name>`

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

1. Add benchmarks to the tests to set SLAs and maintain perf with each feature
2. Add aditional tests for edge cases, and improve test coverage
3. Add feature to honor 'robots.txt'
4. Deal with throttling better; currently just log an error

Ideas and thoughts are always welcome

## License

MIT
