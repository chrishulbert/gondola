package main

import (
	"errors"
	"fmt"
	"github.com/xrash/smetrics"
	"log"
	"os"
	"path/filepath"
	"strings"
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
	tmdbId, tmdbIdErr := requestTmdbTVSearch(showTitleFromFile)
	if tmdbIdErr != nil {
		log.Println("Could not find TV show for", showTitleFromFile)
		return tmdbIdErr
	}

	// Get show details.
	tmdbSeries, _, tmdbSeriesErr := requestTmdbTVShowDetails(tmdbId)
	if tmdbSeriesErr != nil {
		log.Println("Could not get TV show metadata for", showTitleFromFile)
		return tmdbSeriesErr
	}

	allEpisodes := make([]GuessEpisode, 0)

	// Find all seasons.
	for _, season := range tmdbSeries.Seasons {

		tmdbSeason, _, tmdbSeasonErr := requestTmdbTVSeason(tmdbId, season.SeasonNumber)
		if tmdbSeasonErr != nil {
			log.Println("Could not get season metadata for", showTitleFromFile)
			return tmdbSeasonErr
		}

		for _, episode := range tmdbSeason.Episodes {
			guess := GuessEpisode{
				Season:  episode.SeasonNumber,
				Episode: episode.EpisodeNumber,
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
			strings.ToLower(episodeTitleFromFile),
			strings.ToLower(ep.Name),
			1, 1, 2)
		log.Println(episodeTitleFromFile, "to", ep.Name, "is", fmt.Sprintf("%d", thisDistance))
		if thisDistance < closestDistance {
			closestDistance = thisDistance
			closestGuess = &ep
		}
	}

	if closestGuess == nil {
		return errors.New("No guesses found")
	}

	// Rename it and succeed.
	sxex := fmt.Sprintf("S%02dE%02d", closestGuess.Season, closestGuess.Episode)
	newName := tmdbSeries.Name + " " + sxex + " " + closestGuess.Name + "." + file + ".remove if correct"
	os.Rename(filepath.Join(folder, file), filepath.Join(folder, newName))
	log.Println("Guessed a file. You can remove the 'remove if correct' if you're happy with the guess.")
	return nil
}
