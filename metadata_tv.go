package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const (
	episodeListFilename = "episodes.json"
)

type EpisodeMetadata struct {
	Media    string
	Image    string
	Season   int
	Episode  int
	Metadata *OmdbTVEpisode // This can be 'null' if OMDB didn't return info for this ep.
}

/// Regenerates the list of episodes for the given show path.
func generateEpisodeList(showPath string, paths Paths) {

	log.Println("Generating episode list:", showPath)

	var episodes []EpisodeMetadata

	files, _ := ioutil.ReadDir(showPath) // Assume this works.
	for _, fileInfo := range files {
		if fileInfo.IsDir() {
			// Parse the 'SxEy'
			_, season, episode, _ := showSeasonEpisodeFromFile(fileInfo.Name())

			// Load the metadata.
			var metadata *OmdbTVEpisode = nil
			metadataPath := filepath.Join(filepath.Join(showPath, fileInfo.Name()), metadataFilename)
			metadataData, metadataErr := ioutil.ReadFile(metadataPath)
			if metadataErr == nil {
				var m OmdbTVEpisode
				if err := json.Unmarshal(metadataData, &m); err == nil {
					metadata = &m
				}
			}

			epPath := filepath.Join(showPath, fileInfo.Name())
			mediaPath, _ := filepath.Rel(paths.Root, filepath.Join(epPath, hlsFilename))
			imagePath, _ := filepath.Rel(paths.Root, filepath.Join(epPath, imageFilename))

			ep := EpisodeMetadata{
				Media:    mediaPath,
				Image:    imagePath,
				Season:   season,
				Episode:  episode,
				Metadata: metadata,
			}
			episodes = append(episodes, ep)
		}
	}

	// Save.
	data, _ := json.Marshal(episodes)
	outPath := filepath.Join(showPath, episodeListFilename)
	ioutil.WriteFile(outPath, data, os.ModePerm)

	log.Println("Successfully generated episode list")
}
