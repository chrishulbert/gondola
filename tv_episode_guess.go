package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/xrash/smetrics"
)

type GuessEpisode struct {
	Season  int
	Episode int
	Name    string
}

/// From a tv episode eg 'Some TV Show - Episode Name.vob' it looks up tmdb, finds the closest episode name, and
/// renames to eg 'Seinfeld S09E03 The Serenity Now.Seinfeld - Serenity.vob.remove if correct'
/// Returns an error if it can't figure anything out.
func tvEpisodeGuess(folder string, file string, paths Paths, config Config) error {
	s := strings.Split(file, "-")

	if len(s) < 2 {
		return errors.New("Unrecognised file naming, expected eg 'Some show - Episode name.vob'")
	}

	showTitleFromFile := strings.TrimSpace(s[0])
	episodeTitleFromFile := strings.TrimSpace(s[1])

	if showTitleFromFile == "" {
		return errors.New("Missing show name before the dash")
	}

	if episodeTitleFromFile == "" {
		return errors.New("Missing episode name after the dash")
	}

	// Search for the id.
	log.Println("Fetching all episodes metadata for show", showTitleFromFile)
	seriesId := tvdbSearchForSeries(showTitleFromFile)
	if seriesId == "" {
		log.Println("Could not find TV show for", showTitleFromFile)
		return errors.New("Series search")
	}

	// Get show details.
	series, seriesErr := tvdbSeriesDetails(seriesId)
	if seriesErr != nil {
		log.Println("Could not get TV show metadata for", showTitleFromFile)
		return seriesErr
	}

	allEpisodes := make([]GuessEpisode, 0)

	// Find all seasons.
	for _, sparseSeason := range series.Seasons {

		fatSeason, seasonErr := tvdbSeasonDetails(seriesId, sparseSeason.TVDBID, sparseSeason.Season)
		if seasonErr != nil {
			log.Println("Could not get season metadata for", showTitleFromFile)
			return seasonErr
		}

		for _, episode := range fatSeason.Episodes {
			guess := GuessEpisode{
				Season:  episode.SeasonNumber,
				Episode: episode.Episode,
				Name:    episode.Name,
			}
			allEpisodes = append(allEpisodes, guess)
		}

	}

	// Any episodes?
	if len(allEpisodes) == 0 {
		return errors.New("No episodes found")
	}

	// Find the closest.
	closestDistance := 99999
	var closestGuess *GuessEpisode
	for _, ep := range allEpisodes {
		thisDistance := smetrics.WagnerFischer(
			strings.ToLower(ep.Name),
			strings.ToLower(episodeTitleFromFile),
			1, 3, 2)
		if thisDistance < closestDistance {
			closestDistance = thisDistance
			epCopy := ep // Without this, the memory for 'ep' is overwritten next loop iteration.
			closestGuess = &epCopy
		}
	}

	if closestGuess == nil {
		return errors.New("No guesses found")
	}

	// Rename it and succeed.
	sxex := fmt.Sprintf("S%02dE%02d", closestGuess.Season, closestGuess.Episode)
	newName := series.Name + " " + sxex + " " + closestGuess.Name + "." + file + ".remove if correct"
	os.Rename(filepath.Join(folder, file), filepath.Join(folder, newName))
	log.Println("Guessed a file. You can remove the 'remove if correct' if you're happy with the guess.")
	return nil
}
