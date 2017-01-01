package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

type BySeasonThenEpisode []EpisodeMetadata

func (a BySeasonThenEpisode) Len() int      { return len(a) }
func (a BySeasonThenEpisode) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a BySeasonThenEpisode) Less(i, j int) bool {
	return a[i].Season*1000+a[i].Episode < a[j].Season*1000+a[j].Episode
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

	// Sort
	sort.Sort(BySeasonThenEpisode(episodes))

	// Save.
	data, _ := json.Marshal(episodes)
	outPath := filepath.Join(showPath, episodeListFilename)
	ioutil.WriteFile(outPath, data, os.ModePerm)

	generateEpisodeListHTML(showPath, episodes, paths)

	log.Println("Successfully generated episode list")
}

func generateEpisodeListHTML(showPath string, episodes []EpisodeMetadata, paths Paths) {
	html := htmlStart
	isLeft := true
	trOpen := false

	for _, episode := range episodes {
		if isLeft {
			html += "<tr>"
			trOpen = true
		}

		linkPath, _ := filepath.Rel(showPath, filepath.Join(paths.Root, episode.Media))
		imagePath, _ := filepath.Rel(showPath, filepath.Join(paths.Root, episode.Image))
		name := fmt.Sprintf("S%02d E%02d", episode.Season, episode.Episode)
		if episode.Metadata != nil {
			name = name + " " + episode.Metadata.Title
		}

		h := htmlTd
		h = strings.Replace(h, "LINK", linkPath, -1)
		h = strings.Replace(h, "IMAGE", imagePath, -1)
		h = strings.Replace(h, "NAME", name, -1)
		html += h

		if !isLeft {
			html += "</tr>"
			trOpen = false
		}

		isLeft = !isLeft
	}

	if trOpen {
		html += "</tr>"
	}
	html += htmlEnd

	// Save.
	outPath := filepath.Join(showPath, indexHtml)
	ioutil.WriteFile(outPath, []byte(html), os.ModePerm)
}

type ShowMetadata struct {
	Image    string
	Metadata interface{}
	Episodes []interface{}
}

type ByName []ShowMetadata

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].Image < a[j].Image }

/// Regenerates the list of all tv shows.
func generateShowList(paths Paths) {

	log.Println("Generating show list")

	var shows []ShowMetadata

	files, _ := ioutil.ReadDir(paths.TV) // Assume this works.
	for _, fileInfo := range files {
		if fileInfo.IsDir() {

			showPath := filepath.Join(paths.TV, fileInfo.Name())

			// Load the metadata.
			metadataData, metadataErr := ioutil.ReadFile(filepath.Join(showPath, metadataFilename))
			if metadataErr != nil {
				continue
			}
			var metadata interface{}
			if err := json.Unmarshal(metadataData, &metadata); err != nil {
				continue
			}

			// Load the episodes.
			episodesData, episodesErr := ioutil.ReadFile(filepath.Join(showPath, episodeListFilename))
			if episodesErr != nil {
				continue
			}
			var episodes []interface{}
			if err := json.Unmarshal(episodesData, &episodes); err != nil {
				continue
			}

			imagePath, _ := filepath.Rel(paths.Root, filepath.Join(showPath, imageFilename))
			s := ShowMetadata{
				Image:    imagePath,
				Metadata: metadata,
				Episodes: episodes,
			}
			shows = append(shows, s)
		}
	}

	sort.Sort(ByName(shows))

	// Save.
	data, _ := json.MarshalIndent(shows, "", "    ")
	outPath := filepath.Join(paths.TV, metadataFilename)
	ioutil.WriteFile(outPath, data, os.ModePerm)
}
