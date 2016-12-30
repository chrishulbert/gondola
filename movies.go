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

	// Get the OMDB metadata.
	omdbMovie, omdbErr := omdbRequest(fileTitle, year)
	if omdbErr != nil {
		log.Println("Failed to find OMDB data for", fileTitle, "error:", omdbErr)
		failedPath := filepath.Join(paths.Failed, file) // Move to 'failed'.
		os.Rename(inPath, failedPath)
		os.RemoveAll(stagingOutputFolder) // Tidy up.
		return omdbErr
	} else {
		// Save the OMDB metadata.
		metadata, _ := json.Marshal(omdbMovie)
		metadataPath := filepath.Join(stagingOutputFolder, metadataFilename)
		ioutil.WriteFile(metadataPath, metadata, os.ModePerm)
	}

	// Get the image.
	log.Println("Downloading an image")
	imageData, imageErr := imageForTitle(omdbMovie.Title)
	if imageErr != nil {
		log.Println("Couldn't download the image", omdbMovie.Title, imageErr)
	} else {
		// Save the image.
		imagePath := filepath.Join(stagingOutputFolder, imageFilename)
		ioutil.WriteFile(imagePath, imageData, os.ModePerm)
	}

	// Convert it.
	outPath := filepath.Join(stagingOutputFolder, hlsFilename)
	convertErr := convertToHLSAppropriately(inPath, outPath, config)

	// Fail! Move it to the failed folder.
	if convertErr != nil {
		log.Println("Failed to convert", file, "; moving to the Failed folder, err:", convertErr)
		failedPath := filepath.Join(paths.Failed, file) // Move it to 'failed'.
		os.Rename(inPath, failedPath)
		os.RemoveAll(stagingOutputFolder) // Tidy up.
		return errors.New("Couldn't convert " + file)
	}

	// Success!
	log.Println("Success! Removing original.")
	goodFolder := filepath.Join(paths.Movies, fileTitle) // TODO use the title from OMDB and filesystem-sanitise it.
	os.Rename(stagingOutputFolder, goodFolder)           // Move the HLS across.
	os.Remove(inPath)                                    // Remove the original file.
	// Assumption is that the user ripped their original from their DVD so doesn't care to lose it.

	return nil
}
