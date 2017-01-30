package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// Actually processes a file that's in the new folder.
func processTV(folder string, file string, paths Paths, config Config) error {
	inPath := filepath.Join(folder, file)
	log.Println("Processing", file)

	// Parse the title.
	showTitleFromFile, season, episode, err := showSeasonEpisodeFromFile(file)
	if err != nil {

		// Try to guess the season/ep if it's eg `Some TV Show - Episode Name.vob` format.
		guessErr := tvEpisodeGuess(folder, file, paths, config)
		if guessErr == nil {
			// Succeeded in making a guess! Now skip this file because it's been renamed and the user must confirm.
			return nil
		} else {
			log.Println("Couldn't guess the episode", guessErr)
			log.Println("Failed to parse season/episode for", file)
			failedPath := filepath.Join(paths.Failed, file) // Move to 'failed'.
			os.Rename(inPath, failedPath)
			return err
		}
	}

	// Get and save the show data. This has to happen for every episode so we can get the proper title name.

	// Search for the id.
	tmdbId, tmdbIdErr := requestTmdbTVSearch(showTitleFromFile)
	if tmdbIdErr != nil {
		log.Println("Could not find TV show for", showTitleFromFile)
		failedPath := filepath.Join(paths.Failed, file) // Move it to 'failed'.
		os.Rename(inPath, failedPath)
		return tmdbIdErr
	}

	// Get show details.
	tmdbSeries, tmdbSeriesData, tmdbSeriesErr := requestTmdbTVShowDetails(tmdbId)
	if tmdbSeriesErr != nil {
		log.Println("Could not get TV show metadata for", showTitleFromFile)
		failedPath := filepath.Join(paths.Failed, file) // Move it to 'failed'.
		os.Rename(inPath, failedPath)
		return tmdbSeriesErr
	}

	// Get season details.
	tmdbSeason, tmdbSeasonData, tmdbSeasonErr := requestTmdbTVSeason(tmdbId, season)
	if tmdbSeasonErr != nil {
		log.Println("Could not get season metadata for", showTitleFromFile)
		failedPath := filepath.Join(paths.Failed, file) // Move it to 'failed'.
		os.Rename(inPath, failedPath)
		return tmdbSeasonErr
	}

	// Get episode details.
	tmdbEpisode, tmdbEpisodeData, tmdbEpisodeErr := requestTmdbTVEpisode(tmdbId, season, episode)
	if tmdbEpisodeErr != nil {
		log.Println("Could not get episode metadata for", showTitleFromFile)
		failedPath := filepath.Join(paths.Failed, file) // Move it to 'failed'.
		os.Rename(inPath, failedPath)
		return tmdbEpisodeErr
	}

	// Write the details.
	showFolder := filepath.Join(paths.TV, sanitiseForFilesystem(tmdbSeries.Name))
	seasonFolder := filepath.Join(showFolder, tvSeasonFolderNameFor(season))
	episodeFolder := filepath.Join(seasonFolder, tvFolderNameFor(season, episode, tmdbEpisode.Name))
	os.MkdirAll(episodeFolder, os.ModePerm)
	ioutil.WriteFile(filepath.Join(showFolder, metadataFilename), tmdbSeriesData, os.ModePerm)
	ioutil.WriteFile(filepath.Join(seasonFolder, metadataFilename), tmdbSeasonData, os.ModePerm)
	ioutil.WriteFile(filepath.Join(episodeFolder, metadataFilename), tmdbEpisodeData, os.ModePerm)

	// Get pics if needed.
	getImageIfNeeded(tmdbSeries.PosterPath, "w780", showFolder, imageFilename)
	getImageIfNeeded(tmdbSeries.BackdropPath, "w1280", showFolder, imageBackdropFilename)
	getImageIfNeeded(tmdbSeason.PosterPath, "w780", seasonFolder, imageFilename)
	getImageIfNeeded(tmdbEpisode.StillPath, "w300", episodeFolder, imageFilename)

	// Convert it.
	outPath := filepath.Join(episodeFolder, hlsFilename)
	convertErr := convertToHLSAppropriately(inPath, outPath, config)

	// Fail! Move it to the failed folder.
	if convertErr != nil {
		log.Println("Failed to convert", file, "; moving to the Failed folder, err:", convertErr)
		failedPath := filepath.Join(paths.Failed, file) // Move it to 'failed'.
		os.Rename(inPath, failedPath)
		os.RemoveAll(episodeFolder) // Tidy up.
		return errors.New("Couldn't convert " + file)
	}

	// Success!
	// Assumption is that the user ripped their original from their DVD so doesn't care to lose it and would prefer to save the space.
	log.Println("Success! Removing original.")
	os.Remove(inPath)

	// Generate metadata.
	generateMetadata(paths)

	return nil
}

func tvSeasonFolderNameFor(season int) string {
	return fmt.Sprintf("Season %d", season)
}

// Makes the folder name for the given show.
func tvFolderNameFor(season int, episode int, title string) string {
	if title == "" {
		return fmt.Sprintf("S%02dE%02d", season, episode)
	} else {
		return fmt.Sprintf("S%02dE%02d %s", season, episode, sanitiseForFilesystem(title))
	}
}
