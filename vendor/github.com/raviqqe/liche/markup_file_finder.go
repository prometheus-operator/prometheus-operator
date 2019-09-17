package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

type markupFileFinder struct {
	filenames chan string
	errors    chan error
}

func newMarkupFileFinder() markupFileFinder {
	return markupFileFinder{
		make(chan string, maxOpenFiles),
		make(chan error, 64),
	}
}

func (m markupFileFinder) Filenames() chan string {
	return m.filenames
}

func (m markupFileFinder) Errors() chan error {
	return m.errors
}

func (m markupFileFinder) Find(fs []string, recursive bool) {
	for _, f := range fs {
		i, err := os.Stat(f)

		if err != nil {
			m.errors <- err
			continue
		}

		if i.IsDir() && recursive {
			m.listDirectory(f)
		} else if i.IsDir() {
			m.errors <- fmt.Errorf("%v is not a file", f)
		} else {
			m.filenames <- f
		}
	}

	close(m.filenames)
	close(m.errors)
}

func (m markupFileFinder) listDirectory(d string) {
	err := filepath.Walk(d, func(f string, i os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		b, err := regexp.MatchString("(^\\.)|(/\\.)", f)

		if err != nil {
			return err
		}

		if !i.IsDir() && !b && isMarkupFile(f) {
			m.filenames <- f
		}

		return nil
	})

	if err != nil {
		m.errors <- err
	}
}
