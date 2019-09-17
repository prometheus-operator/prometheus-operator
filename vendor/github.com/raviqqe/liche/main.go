package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/fatih/color"
)

func main() {
	args, err := getArguments(nil)

	if err != nil {
		fail(err)
	}

	m := newMarkupFileFinder()
	wg := sync.WaitGroup{}

	go m.Find(args.filenames, args.recursive)

	wg.Add(1)
	go func() {
		for e := range m.Errors() {
			fail(e)
		}

		wg.Done()
	}()

	rc := make(chan fileResult, maxOpenFiles)
	c := newFileChecker(
		args.timeout,
		args.documentRoot,
		args.excludedPattern,
		newSemaphore(args.concurrency))

	go c.CheckMany(m.Filenames(), rc)

	ok := true

	for r := range rc {
		if !r.Ok() {
			ok = false
			printToStderr(r.String(args.verbose))
		} else if args.verbose {
			printToStderr(r.String(true))
		}
	}

	wg.Wait()

	if !ok {
		os.Exit(1)
	}
}

func printToStderr(xs ...interface{}) {
	if _, err := fmt.Fprintln(os.Stderr, xs...); err != nil {
		panic(err)
	}
}

func fail(err error) {
	printToStderr(color.RedString(capitalizeFirst(err.Error())))
	os.Exit(1)
}
