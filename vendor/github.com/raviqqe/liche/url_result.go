package main

import (
	"strings"

	"github.com/fatih/color"
)

type urlResult struct {
	url string
	err error
}

func (r urlResult) String() string {
	if r.err == nil {
		return color.GreenString("OK") + "\t" + r.url
	}

	s := r.err.Error()

	return color.RedString("ERROR") + "\t" + r.url + "\n\t" +
		color.YellowString(strings.ToUpper(s[:1])+s[1:])
}
