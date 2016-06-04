// Exercise in Tour of Go
// Implements a parallel web crawler
// Usage: crawler <url> <max-depth>
//
// Author: Ziyi Yan <cxfyzy@gmail.com>
// Date: 2016-06-03

package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

type Crawler struct {
	visited map[string]bool
	muxLock sync.Mutex
}

func (c *Crawler) Fetch(url string) (body string, urls []string, err error) {
	// Mark the `url` visited.
	// Protect this operation with lock.
	c.muxLock.Lock()
	if c.visited[url] {
		c.muxLock.Unlock()
		return
	} else {
		c.visited[url] = true
	}
	c.muxLock.Unlock()

	// Fetch the web body
	fmt.Printf("Fetching url: %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("http.Get(): ", err)
		return
	}
	//defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("http.Get(): ", err)
		return
	}
	body = string(bytes)

	// Match urls in body
	urlRegex := regexp.MustCompile(`(http|ftp|https):\/\/([\w\-_]+(?:(?:\.[\w\-_]+)+))([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?`)
	urls = urlRegex.FindAllString(body, -1)
	return
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher) {
	// TODO: Fetch URLs in parallel.
	// TODO: Don't fetch the same URL twice.
	// This implementation doesn't do either:
	if depth <= 0 {
		return
	}
	body, urls, err := fetcher.Fetch(url)
	fmt.Println("urls: ", urls)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("found(depth %v): %s %q\n", depth, url, body)
	// Crawl `urls` parallel
	var wg sync.WaitGroup
	wg.Add(len(urls))
	for _, u := range urls {
		go func(url string) {
			Crawl(url, depth-1, fetcher)
			wg.Done()
		}(u)
	}
	wg.Wait()
	return
}

func main() {
	fetcher := Crawler{
		visited: make(map[string]bool),
	}
	if len(os.Args) < 3 {
		fmt.Printf("%s: Too few arguments.\n", os.Args[0])
		return
	}

	depth, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("%s: need an integer for sencond argument.\n", os.Args[0])
	}
	Crawl(os.Args[1], depth, &fetcher)
}
