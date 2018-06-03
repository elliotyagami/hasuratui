package main

import (
	"fmt"
	"net/http"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func scrapeIndexHasura() map[string]string {
	// request and parse the front page
	resp, err := http.Get("https://docs.hasura.io/0.15/manual/index.html")
	if err != nil {
		panic(err)
	}
	root, err := html.Parse(resp.Body)
	if err != nil {
		panic(err)
	}

	// define a matcher
	matcher := func(n *html.Node) bool {
		// must check for nil values
		if n.DataAtom == atom.A && n.Parent != nil && n.Parent.Parent != nil {
			return scrape.Attr(n, "class") == "reference internal"
		}
		return false
	}
	// grab all articles and print them
	articles := scrape.FindAll(root, matcher)
	// var indexEntries map[string]string
	indexEntries := make(map[string]string)

	for i, article := range articles {
		// fmt.Printf("%2d %s (%s)\n", i, scrape.Text(article), scrape.Attr(article, "href"))
		key := fmt.Sprintf("%d %s", i, scrape.Text(article))
		indexEntries[key] = scrape.Attr(article, "href")
	}
	return indexEntries
}
