package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Actually processes a file that's in the new folder.
func processTV(folder string, file string, paths Paths, config Config) error {
	inPath := filepath.Join(folder, file)
	log.Println("Processing", file)

	var series TVDBSeries
	var season TVDBSeason
	var seasonNumber int
	var episode TVDBEpisode
	var episodeNumber int

	// Is this the kind of file we fetch metadata online for?
	if strings.HasPrefix(file, "!") { // It's a non-TMDB file.
		// Should be like "!Tacfit - S1 Lite - E1 Instructions.mp4"
		extension := filepath.Ext(file)
		nameSansExtension := strings.TrimSuffix(file, extension)
		regex := regexp.MustCompile(`(?i)!(.*?) - S(\d+)(.*?) - E(\d+)(.*)`)
		matches := regex.FindStringSubmatch(nameSansExtension)
		if len(matches) >= 6 {
			seriesName := strings.TrimSpace(matches[1])
			seasonNumber, _ = strconv.Atoi(matches[2])
			seasonName := strings.TrimSpace(matches[3])
			episodeNumber, _ = strconv.Atoi(matches[4])
			episodeName := strings.TrimSpace(matches[5])
			series = TVDBSeries{
				Name: seriesName,
			}
			season = TVDBSeason{
				Season: seasonNumber,
				Name:   seasonName,
			}
			episode = TVDBEpisode{
				SeasonNumber: seasonNumber,
				Episode:      episodeNumber,
				Name:         episodeName,
			}
		} else {
			return errors.New("Could not parse filename, expecting something like '!MySeries - S10 Season X - E01 MyEpisode.mp4'")
		}
	} else { // TMDB file.
		// Parse the title.
		showTitleFromFile, seasonNumber, episodeNumber, err := showSeasonEpisodeFromFile(file)
		if err != nil {

			// Try to guess the season/ep if it's eg `Some TV Show - Episode Name.vob` format.
			guessErr := tvEpisodeGuess(folder, file, paths, config)
			if guessErr == nil {
				// Succeeded in making a guess! Now skip this file because it's been renamed and the user must confirm.
				return nil
			} else {
				log.Println("Couldn't guess the episode, error:", guessErr)
				log.Println("Failed to parse season/episode for", file)
				failedPath := filepath.Join(paths.Failed, file) // Move to 'failed'.
				os.Rename(inPath, failedPath)
				return err
			}
		}

		// Get and save the show data. This has to happen for every episode so we can get the proper title name.

		// Search for the id.
		seriesId := tvdbSearchForSeries(showTitleFromFile)
		if seriesId == "" {
			log.Println("Could not find TV show for", showTitleFromFile)
			failedPath := filepath.Join(paths.Failed, file) // Move it to 'failed'.
			os.Rename(inPath, failedPath)
			return errors.New("Could not find TV show")
		}

		// Get show details.
		series, err = tvdbSeriesDetails(seriesId)
		if err != nil {
			log.Println("Could not get TV show metadata for", showTitleFromFile)
			failedPath := filepath.Join(paths.Failed, file) // Move it to 'failed'.
			os.Rename(inPath, failedPath)
			return err
		}

		// Find the season id.
		seasonId := 0
		for _, season := range series.Seasons {
			if season.Season == seasonNumber {
				seasonId = season.TVDBID
			}
		}
		if seasonId <= 0 {
			log.Println("Could not find season number", seasonNumber)
			failedPath := filepath.Join(paths.Failed, file) // Move it to 'failed'.
			os.Rename(inPath, failedPath)
			return errors.New("Season number")
		}

		// Get season details.
		log.Println("Getting the season details...")
		season, err = tvdbSeasonDetails(seriesId, seasonId, seasonNumber)
		if err != nil {
			log.Println("Could not get season metadata for", showTitleFromFile, "; seriesId", seriesId, "seasonId", seasonId, "seasonNumber", seasonNumber)
			failedPath := filepath.Join(paths.Failed, file) // Move it to 'failed'.
			os.Rename(inPath, failedPath)
			return err
		}

		// Find the episode ID.
		episodeId := 0
		for _, episode := range season.Episodes {
			if episode.Episode == episodeNumber {
				episodeId = episode.TVDBID
			}
		}
		if episodeId <= 0 {
			log.Println("Could not find episode id for ", showTitleFromFile)
			failedPath := filepath.Join(paths.Failed, file) // Move it to 'failed'.
			os.Rename(inPath, failedPath)
			return errors.New("Episode number")
		}

		// Get episode details.
		episode, err = tvdbEpisodeDetails(seriesId, seasonId, seasonNumber, episodeId)
		if err != nil {
			log.Println("Could not get episode metadata for", showTitleFromFile)
			failedPath := filepath.Join(paths.Failed, file) // Move it to 'failed'.
			os.Rename(inPath, failedPath)
			return err
		}
	}

	// Write the details.
	showFolder := filepath.Join(paths.TV, sanitiseForFilesystem(series.Name))
	seasonFolder := filepath.Join(showFolder, tvSeasonFolderNameFor(seasonNumber))
	episodeFolder := filepath.Join(seasonFolder, tvFolderNameFor(seasonNumber, episodeNumber, episode.Name))
	os.MkdirAll(episodeFolder, os.ModePerm)
	seriesData, _ := json.Marshal(series)
	seasonData, _ := json.Marshal(season)
	episodeData, _ := json.Marshal(episode)
	ioutil.WriteFile(filepath.Join(showFolder, metadataFilename), seriesData, os.ModePerm)
	ioutil.WriteFile(filepath.Join(seasonFolder, metadataFilename), seasonData, os.ModePerm)
	ioutil.WriteFile(filepath.Join(episodeFolder, metadataFilename), episodeData, os.ModePerm)

	// Get pics if needed.
	getTVImageIfNeeded(series.Poster, showFolder, imageFilename)
	getTVImageIfNeeded(series.Art, showFolder, imageBackdropFilename)
	getTVImageIfNeeded(season.Image, seasonFolder, imageFilename)
	getTVImageIfNeeded(episode.Image, episodeFolder, imageFilename)

	// Convert it.
	outPath := filepath.Join(episodeFolder, hlsFilename)
	convertErr := convertToHLSAppropriately(inPath, outPath, config)

	// Fail! Move it to the failed folder.
	if convertErr != nil {
		switch err := convertErr.(type) {
		case *convertRenamedError:
			log.Println("Failed to convert", file, "; file renamed for user intervention, err:", err)
		default:
			log.Println("Failed to convert", file, "; moving to the Failed folder, err:", err)
			failedPath := filepath.Join(paths.Failed, file) // Move it to 'failed'.
			os.Rename(inPath, failedPath)
		}
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
