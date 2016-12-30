package main

import (
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
	title := titleFromFilename(file)

	// Make the temporary output folder.
	stagingOutputFolder := filepath.Join(paths.Staging, title)
	os.MkdirAll(stagingOutputFolder, os.ModePerm)

	// Get the image.
	log.Println("Downloading an image")
	imageData, imageErr := imageForTitle(title)
	if imageErr != nil {
		log.Println("Couldn't download the image", title, imageErr)
	} else {
		// Save the image.
		imagePath := filepath.Join(stagingOutputFolder, imageFilename)
		ioutil.WriteFile(imagePath, imageData, os.ModePerm)
	}

	// Convert it.
	inPath := filepath.Join(folder, file)
	outPath := filepath.Join(stagingOutputFolder, hlsFilename)
	convertErr := convertToHLSAppropriately(inPath, outPath, config)

	// Fail! Move it to the failed folder.
	if convertErr != nil {
		log.Println("Failed to convert", file, "; moving to the Failed folder, err:", convertErr)
		failedPath := filepath.Join(paths.Failed, file)
		os.Rename(inPath, failedPath)
		os.RemoveAll(stagingOutputFolder) // Tidy up.
		return errors.New("Couldn't convert " + file)
	}

	// Success!
	log.Println("Success! Removing original.")
	goodFolder := filepath.Join(paths.Movies, title)
	os.Rename(stagingOutputFolder, goodFolder) // Move the HLS across.
	os.Remove(inPath)                          // Remove the original file.
	// Assumption is that the user ripped their original from their DVD so doesn't care to lose it.

	return nil
}
