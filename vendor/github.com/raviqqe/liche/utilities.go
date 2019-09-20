package main

import (
	"strings"

	"github.com/kr/text"
)

func stringSetToSlice(s2b map[string]bool) []string {
	ss := make([]string, 0, len(s2b))

	for s := range s2b {
		ss = append(ss, s)
	}

	return ss
}

func indent(s string) string {
	return text.Indent(s, "\t")
}

func capitalizeFirst(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}
