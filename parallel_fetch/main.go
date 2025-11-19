package main

import (
	"fmt"
	"sync"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, ch chan string) {
	// depth limit
	if depth <= 0 {
		return
	}
	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	// send result
	ch <- fmt.Sprintf("found: %s %q\n", url, body)
	// spawn children: increment WaitGroup before starting each goroutine
	for _, u := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			Crawl(u, depth-1, fetcher, ch)
		}(u)
	}
}

var wg *sync.WaitGroup

func main() {
	ch := make(chan string)
	wg = &sync.WaitGroup{}
	// start initial crawl
	wg.Add(1)
	go func() {
		defer wg.Done()
		Crawl("https://golang.org/", 4, fetcher, ch)
	}()

	// close channel once all crawls complete
	go func() {
		wg.Wait()
		close(ch)
	}()

	// collect results after channel is closed
	var results []string
	for s := range ch {
		results = append(results, s)
	}

	// results now contains all messages sent on ch
	for _, r := range results {
		fmt.Print(r)
	}
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}
