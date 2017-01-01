package main

import (
	"encoding/json"
	// "errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// Actually processses a file that's in the new folder.
func processTV(folder string, file string, paths Paths, config Config) error {
	inPath := filepath.Join(folder, file)
	log.Println("Processing", file)

	// Parse the title.
	showTitleFromFile, _, _, err := showSeasonEpisodeFromFile(file)
	// showTitleFromFile, season, episode, err := showSeasonEpisodeFromFile(file)
	if err != nil {
		log.Println("Failed to parse season/episode for", file)
		failedPath := filepath.Join(paths.Failed, file) // Move to 'failed'.
		os.Rename(inPath, failedPath)
		return err
	}

	// Get and save the show data. This has to happen for every episode so we can get the proper title name.
	omdbSeries, omdbErr := omdbRequestTVSeries(showTitleFromFile)
	if omdbErr != nil {
		log.Println("Could not get OMDB metadata for", showTitleFromFile)
		failedPath := filepath.Join(paths.Failed, file) // Move it to 'failed'.
		os.Rename(inPath, failedPath)
		return omdbErr
	}
	showOutputFolder := filepath.Join(paths.TV, sanitiseForFilesystem(omdbSeries.Title))
	os.MkdirAll(showOutputFolder, os.ModePerm)
	seriesMetadata, _ := json.Marshal(omdbSeries)
	seriesMetadataPath := filepath.Join(showOutputFolder, metadataFilename)
	ioutil.WriteFile(seriesMetadataPath, seriesMetadata, os.ModePerm)

	// Get show pic if needed.
	seriesImagePath := filepath.Join(showOutputFolder, imageFilename)
	if _, err := os.Stat(seriesImagePath); os.IsNotExist(err) {
		if image, err := imageForIMDB(omdbSeries.ImdbID); err == nil {
			ioutil.WriteFile(seriesImagePath, image, os.ModePerm)
		}
	}

	// // Make the temporary output folder.
	// stagingOutputFolder := filepath.Join(paths.Staging, file)
	// os.MkdirAll(stagingOutputFolder, os.ModePerm)

	// // Get the OMDB metadata.
	// omdbMovie, omdbErr := omdbRequest(fileTitle, year)
	// if omdbErr != nil {
	// 	log.Println("Failed to find OMDB data for", fileTitle, "error:", omdbErr)
	// 	failedPath := filepath.Join(paths.Failed, file) // Move to 'failed'.
	// 	os.Rename(inPath, failedPath)
	// 	os.RemoveAll(stagingOutputFolder) // Tidy up.
	// 	return omdbErr
	// } else {
	// 	// Save the OMDB metadata.
	// 	metadata, _ := json.Marshal(omdbMovie)
	// 	metadataPath := filepath.Join(stagingOutputFolder, metadataFilename)
	// 	ioutil.WriteFile(metadataPath, metadata, os.ModePerm)
	// }

	// // Get the image.
	// log.Println("Downloading an image")
	// imageData, imageErr := imageForTitle(omdbMovie.Title)
	// if imageErr != nil {
	// 	log.Println("Couldn't download the image", omdbMovie.Title, imageErr)
	// } else {
	// 	// Save the image.
	// 	imagePath := filepath.Join(stagingOutputFolder, imageFilename)
	// 	ioutil.WriteFile(imagePath, imageData, os.ModePerm)
	// }

	// // Convert it.
	// outPath := filepath.Join(stagingOutputFolder, hlsFilename)
	// convertErr := convertToHLSAppropriately(inPath, outPath, config)

	// // Fail! Move it to the failed folder.
	// if convertErr != nil {
	// 	log.Println("Failed to convert", file, "; moving to the Failed folder, err:", convertErr)
	// 	failedPath := filepath.Join(paths.Failed, file) // Move it to 'failed'.
	// 	os.Rename(inPath, failedPath)
	// 	os.RemoveAll(stagingOutputFolder) // Tidy up.
	// 	return errors.New("Couldn't convert " + file)
	// }

	// // Success!
	// log.Println("Success! Removing original.")
	// TODO use the title from OMDB and filesystem-sanitise it!!! So that we can do the same with TV
	// goodTitle := fileTitle + " " + omdbMovie.Year
	// goodFolder := filepath.Join(paths.Movies, goodTitle)
	// os.Rename(stagingOutputFolder, goodFolder) // Move the HLS across.
	os.Remove(inPath) // Remove the original file.
	// Assumption is that the user ripped their original from their DVD so doesn't care to lose it.

	return nil
}
