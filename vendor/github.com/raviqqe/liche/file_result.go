package main

import (
	"sort"
	"strings"

	"github.com/fatih/color"
)

type fileResult struct {
	filename   string
	urlResults []urlResult
	err        error
}

func (r fileResult) String(verbose bool) string {
	ss := make([]string, 0, len(r.urlResults))

	if r.err != nil {
		ss = append(ss, indent(color.RedString(capitalizeFirst(r.err.Error()))))
	}

	os := make([]string, 0, len(r.urlResults))
	xs := make([]string, 0, len(r.urlResults))

	for _, r := range r.urlResults {
		s := indent(r.String())

		if r.err != nil {
			xs = append(xs, s)
		} else if verbose {
			os = append(os, s)
		}
	}

	sort.Strings(os)
	sort.Strings(xs)

	return strings.Join(append([]string{r.filename}, append(ss, append(os, xs...)...)...), "\n")
}

func (r fileResult) Ok() bool {
	if r.err != nil {
		return false
	}

	for _, r := range r.urlResults {
		if r.err != nil {
			return false
		}
	}

	return true
}
