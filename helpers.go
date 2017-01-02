package main

import (
	"strings"
)

// Sanitises to make a filesystem-safe name.
func sanitiseForFilesystem(s string) string {
	return strings.Replace(s, "/", "-", -1)
}
