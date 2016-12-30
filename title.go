package main

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Expects eg 'Big.Buck.Bunny.2008.blahblah.vob'
// Nice-ify the title for a filename. Best case, becomes "Some movie", 2016.
// If it cannot find year, returns a nil year pointer.
func titleAndYearFromFilename(file string) (string, *int) {
	// Try to figure out the year and title.
	regex := regexp.MustCompile("\\d{4}")
	yearString := regex.FindString(file)
	if yearString != "" {
		rawTitle := regex.Split(file, 2)[0]
		titleNoDots := strings.Replace(rawTitle, ".", " ", -1)
		title := strings.TrimSpace(titleNoDots)
		yearInt, _ := strconv.Atoi(yearString)
		return title, &yearInt
	} else {
		// Just deal with it as-is.
		extension := filepath.Ext(file)
		nameSansExtension := strings.TrimSuffix(file, extension)
		titleNoDots := strings.Replace(nameSansExtension, ".", " ", -1)
		title := strings.TrimSpace(titleNoDots)
		return title, nil
	}
}
