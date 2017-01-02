package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
)

type MovieMetadata struct {
	Media    string
	Image    string
	Metadata OmdbMovie
}

type ByTitle []MovieMetadata

func (a ByTitle) Len() int           { return len(a) }
func (a ByTitle) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTitle) Less(i, j int) bool { return a[i].Metadata.Title < a[j].Metadata.Title }

/// Regenerates the list of all movies.
func generateMovieList(paths Paths) {
	var movies []MovieMetadata

	files, _ := ioutil.ReadDir(paths.Movies) // Assume this works.
	for _, fileInfo := range files {
		if fileInfo.IsDir() {
			moviePath := filepath.Join(paths.Movies, fileInfo.Name())

			var metadata OmdbMovie
			{
				// Load the metadata.
				data, err := ioutil.ReadFile(filepath.Join(moviePath, metadataFilename))
				if err != nil {
					continue
				}
				if err := json.Unmarshal(data, &metadata); err != nil {
					continue
				}
			}

			mediaPath, _ := filepath.Rel(paths.Root, filepath.Join(moviePath, hlsFilename))
			imagePath, _ := filepath.Rel(paths.Root, filepath.Join(moviePath, imageFilename))
			m := MovieMetadata{
				Media:    mediaPath,
				Image:    imagePath,
				Metadata: metadata,
			}
			movies = append(movies, m)
		}
	}

	sort.Sort(ByTitle(movies))

	// Save.
	data, _ := json.MarshalIndent(movies, "", "    ")
	outPath := filepath.Join(paths.Movies, metadataFilename)
	ioutil.WriteFile(outPath, data, os.ModePerm)
}
