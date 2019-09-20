package main

import (
	"bytes"
	"io/ioutil"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
	"gopkg.in/russross/blackfriday.v2"
)

type fileChecker struct {
	urlChecker urlChecker
	semaphore  semaphore
}

func newFileChecker(timeout time.Duration, d string, r *regexp.Regexp, s semaphore) fileChecker {
	return fileChecker{newURLChecker(timeout, d, r, s), s}
}

func (c fileChecker) Check(f string) ([]urlResult, error) {
	n, err := c.parseFile(f)

	if err != nil {
		return nil, err
	}

	us, err := c.extractURLs(n)

	if err != nil {
		return nil, err
	}

	rc := make(chan urlResult, len(us))
	rs := make([]urlResult, 0, len(us))

	go c.urlChecker.CheckMany(us, f, rc)

	for r := range rc {
		rs = append(rs, r)
	}

	return rs, nil
}

func (c fileChecker) CheckMany(fc <-chan string, rc chan<- fileResult) {
	wg := sync.WaitGroup{}

	for f := range fc {
		wg.Add(1)

		go func(f string) {
			if rs, err := c.Check(f); err == nil {
				rc <- fileResult{filename: f, urlResults: rs}
			} else {
				rc <- fileResult{filename: f, err: err}
			}

			wg.Done()
		}(f)
	}

	wg.Wait()
	close(rc)
}

func (c fileChecker) parseFile(f string) (*html.Node, error) {
	c.semaphore.Request()
	bs, err := ioutil.ReadFile(f)
	c.semaphore.Release()

	if err != nil {
		return nil, err
	}

	if !isHTMLFile(f) {
		bs = blackfriday.Run(bs)
	}

	n, err := html.Parse(bytes.NewReader(bs))

	if err != nil {
		return nil, err
	}

	return n, nil
}

func (c fileChecker) extractURLs(n *html.Node) ([]string, error) {
	us := make(map[string]bool)
	ns := []*html.Node{n}

	for len(ns) > 0 {
		i := len(ns) - 1
		n := ns[i]
		ns = ns[:i]

		if n.Type == html.ElementNode {
			switch n.Data {
			case "a":
				for _, a := range n.Attr {
					if a.Key == "href" && isURL(a.Val) {
						us[a.Val] = true
						break
					}
				}
			case "img":
				for _, a := range n.Attr {
					if a.Key == "src" && isURL(a.Val) {
						us[a.Val] = true
						break
					}
				}
			}
		}

		for n := n.FirstChild; n != nil; n = n.NextSibling {
			ns = append(ns, n)
		}
	}

	return stringSetToSlice(us), nil
}

func isURL(s string) bool {
	if strings.HasPrefix(s, "#") {
		return false
	}

	u, err := url.Parse(s)
	return err == nil && (u.Scheme == "" || u.Scheme == "http" || u.Scheme == "https")
}
