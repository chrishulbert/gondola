package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// Actually processses a file that's in the new folder.
func processMovie(folder string, file string, paths Paths, config Config) error {
	log.Println("Processing", file)

	// Parse the title.
	fileTitle, year := titleAndYearFromFilename(file)

	// Make the temporary output folder.
	inPath := filepath.Join(folder, file)
	stagingOutputFolder := filepath.Join(paths.Staging, fileTitle)
	os.MkdirAll(stagingOutputFolder, os.ModePerm)

	// Get the metadata.
	tmdbMovie, tmdbErr := requestTmdbMovieSearch(fileTitle, year)
	if tmdbErr != nil {
		log.Println("Failed to find TMDB data for", fileTitle, "error:", tmdbErr)
		failedPath := filepath.Join(paths.Failed, file) // Move to 'failed'.
		os.Rename(inPath, failedPath)
		os.RemoveAll(stagingOutputFolder) // Tidy up.
		return tmdbErr
	} else {
		// Save the metadata.
		metadata, _ := json.Marshal(tmdbMovie)
		metadataPath := filepath.Join(stagingOutputFolder, metadataFilename)
		ioutil.WriteFile(metadataPath, metadata, os.ModePerm)
	}

	// Get the image.
	getImageIfNeeded(tmdbMovie.PosterPath, "w780", stagingOutputFolder, imageFilename)
	getImageIfNeeded(tmdbMovie.BackdropPath, "w1280", stagingOutputFolder, imageBackdropFilename)

	// Convert it.
	outPath := filepath.Join(stagingOutputFolder, hlsFilename)
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
		os.RemoveAll(stagingOutputFolder) // Tidy up.
		return errors.New("Couldn't convert " + file)
	}

	// Success!
	log.Println("Success! Removing original.")
	goodTitle := sanitiseForFilesystem(tmdbMovie.Title) + " " + tmdbMovie.ReleaseDate[:4]
	goodFolder := filepath.Join(paths.Movies, goodTitle)
	os.Rename(stagingOutputFolder, goodFolder) // Move the HLS across.
	os.Remove(inPath)                          // Remove the original file.
	// Assumption is that the user made a backup of their original from their DVD so doesn't care to lose it.

	generateMetadata(paths)

	return nil
}
