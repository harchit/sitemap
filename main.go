package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	link "github.com/harchit/linkparser"
)

//1 get the webpage
//2 parse all links on page using linkparser
//3 build urls with links
//4 filter out different domain links
//5 find all pages (bfs)
//6 print out xml

type loc struct {
	Value string `xml:"loc"`
}
type urlSet struct {
	Urls  []loc  `xml:"url"`
	Xmlns string `xml:"xmlns,attr"`
}

const xmlns = "http://www.sitemaps.org/schemas/sitemap/0.9"

func main() {
	urlFlag := flag.String("url", "https://gophercises.com", "url for building sitemap")
	maxDepth := flag.Int("depth", 3, "max number of links deep to traverse")
	flag.Parse()

	pages := bfs(*urlFlag, *maxDepth)

	toXml := urlSet{
		Xmlns: xmlns,
	}
	for _, page := range pages {
		toXml.Urls = append(toXml.Urls, loc{page})
	}
	fmt.Print(xml.Header)
	enc := xml.NewEncoder(os.Stdout)
	enc.Indent("", "  ")
	if err := enc.Encode(toXml); err != nil {
		panic(err)
	}

}

func bfs(urlStr string, maxDepth int) []string {
	seen := make(map[string]struct{})

	//queue of domains we need to GET
	var q map[string]struct{}

	//queue of domains seen from GETting each domain in the first queue
	nq := map[string]struct{}{
		urlStr: struct{}{},
	}

	for i := 0; i <= maxDepth; i++ {
		q, nq = nq, make(map[string]struct{})
		for url, _ := range q {
			//check if url is already visited, if not set as seen
			if _, ok := seen[url]; ok {
				continue
			}
			seen[url] = struct{}{}

			//set values for next queue
			for _, link := range get(url) {
				nq[link] = struct{}{}
			}
		}
	}
	var ret = make([]string, 0, len(seen))
	for url, _ := range seen {
		ret = append(ret, url)
	}
	return ret
}

func get(urlStr string) []string {
	res, err := http.Get(urlStr)
	if err != nil {
		log.Fatal(err)
		return []string{}
	}
	defer res.Body.Close()
	reqUrl := res.Request.URL
	baseUrl := &url.URL{
		Scheme: reqUrl.Scheme,
		Host:   reqUrl.Host,
	}
	base := baseUrl.String()
	return filter(base, hrefs(res.Body, base))

}
func hrefs(res io.Reader, base string) []string {
	links, _ := link.Parse(res)

	var hrefs []string
	for _, l := range links {
		switch {
		case strings.HasPrefix(l.Href, "/"):
			hrefs = append(hrefs, base+l.Href)
		case strings.HasPrefix(l.Href, "http"):
			hrefs = append(hrefs, l.Href)
		default:
		}
	}

	return hrefs
}

func filter(base string, links []string) []string {
	var ret []string
	for _, link := range links {
		if strings.HasPrefix(link, base) {
			ret = append(ret, link)
		}
	}

	return ret
}
