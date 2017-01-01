package main

import (
	"net/url"
	"strings"
)

// Sanitises to make a filesystem-safe name.
func sanitiseForFilesystem(filename string) string {
	s := url.QueryEscape(filename)
	return strings.Replace(s, "+", " ", -1)
}
