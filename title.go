package main

import (
	"regexp"
	"strings"
	"path/filepath"	
)

// Nice-ify the title for a filename. Best case, becomes "Some movie 2016".
func titleFromFilename(file string) string {
	// Try to figure out the year and title.
	regex := regexp.MustCompile("\\d{4}")
	year := regex.FindString(file)
	if year != "" {
		rawTitle := regex.Split(file, 2)[0]
		titleNoDots := strings.Replace(rawTitle, ".", " ", -1)
		title := strings.TrimSpace(titleNoDots)

		return title + " " + year
	} else {
		// Just deal with it as-is.
		extension := filepath.Ext(file)
		nameSansExtension := strings.TrimSuffix(file, extension)
		titleNoDots := strings.Replace(nameSansExtension, ".", " ", -1)
		title := strings.TrimSpace(titleNoDots)

		return title
	}
}
