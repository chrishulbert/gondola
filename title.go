package main

import (
	"errors"
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

func showSeasonEpisodeFromFile(file string) (string, int, int, error) {
	// Try to figure out the year and title.
	regex := regexp.MustCompile(`(?i)(.*)S(\d+)E(\d+)`)
	matches := regex.FindStringSubmatch(file)
	if len(matches) >= 4 {
		title := strings.TrimSpace(strings.Replace(matches[1], ".", " ", -1))
		season, _ := strconv.Atoi(matches[2])
		episode, _ := strconv.Atoi(matches[3])
		return title, season, episode, nil
	} else {
		return "", 0, 0, errors.New("Couldn't find SxxEyy in " + file)
	}
}

// Find `AudioStreamX` in a file and returns X. Or nil if it can't find. 0 is a valid stream number.
func audioStreamFromFile(file string) *int {
	regex := regexp.MustCompile(`AudioStream(\d+)`)
	matches := regex.FindStringSubmatch(file)
	if len(matches) >= 2 {
		index, _ := strconv.Atoi(matches[1])
		return &index
	} else {
		return nil
	}
}
