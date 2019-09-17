package main

import (
	"errors"
	"net/url"
	"os"
	"path"
	"regexp"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

type urlChecker struct {
	timeout         time.Duration
	documentRoot    string
	excludedPattern *regexp.Regexp
	semaphore       semaphore
}

func newURLChecker(t time.Duration, d string, r *regexp.Regexp, s semaphore) urlChecker {
	return urlChecker{t, d, r, s}
}

func (c urlChecker) Check(u string, f string) error {
	u, local, err := c.resolveURL(u, f)
	if err != nil {
		return err
	}

	if c.excludedPattern != nil && c.excludedPattern.MatchString(u) {
		return nil
	}

	if local {
		_, err := os.Stat(u)
		return err
	}

	c.semaphore.Request()
	defer c.semaphore.Release()

	if c.timeout == 0 {
		_, _, err := fasthttp.Get(nil, u)
		return err
	}

	_, _, err = fasthttp.GetTimeout(nil, u, c.timeout)
	return err
}

func (c urlChecker) CheckMany(us []string, f string, rc chan<- urlResult) {
	wg := sync.WaitGroup{}

	for _, s := range us {
		wg.Add(1)

		go func(s string) {
			rc <- urlResult{s, c.Check(s, f)}
			wg.Done()
		}(s)
	}

	wg.Wait()
	close(rc)
}

func (c urlChecker) resolveURL(u string, f string) (string, bool, error) {
	uu, err := url.Parse(u)

	if err != nil {
		return "", false, err
	}

	if uu.Scheme != "" {
		return u, false, nil
	}

	if !path.IsAbs(uu.Path) {
		return path.Join(path.Dir(f), uu.Path), true, nil
	}

	if c.documentRoot == "" {
		return "", false, errors.New("document root directory is not specified")
	}

	return path.Join(c.documentRoot, uu.Path), true, nil
}
