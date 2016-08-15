package main

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/rbeaudoin/gobot/crawler"
	"github.com/urfave/cli"
)

func main() {
	app := getApp()
	app.Run(os.Args)
}

func getApp() *cli.App {
	app := cli.NewApp()
	app.Name = "gobot"
	app.Usage = "a web crawler written in Go"
	app.Version = "0.0.1"

	app.Commands = []cli.Command{
		{
			Name:    "crawl",
			Aliases: []string{"c"},
			Usage:   "crawl a domain",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "domain, d",
					Usage: "`DOMAIN` to crawl",
				},
			},
			Action: func(c *cli.Context) error {
				dmn := c.String("domain")
				url, err := url.Parse(fmt.Sprintf("http://%s/", dmn))
				if err != nil {
					log.Fatal(err)
				}
				sm, err := crawler.Crawl(*url)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("%s\n", sm)
				return nil
			},
		},
	}
	return app
}
