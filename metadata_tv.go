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

type Metadata struct {
	TVShows  []interface{}
	Movies   []interface{}
	Capacity string
}

type TVShowMetadata struct {
	TMDBId       int
	Name         string
	Overview     string
	Image        string
	Backdrop     string
	FirstAirDate string
	LastAirDate  string
	Seasons      []TVSeasonMetadata
}

type TVSeasonMetadata struct {
	TMDBId   int
	Season   int
	Name     string
	Overview string
	Image    string
	AirDate  string
	Episodes []TVEpisodeMetadata
}

type TVEpisodeMetadata struct {
	TMDBId         int
	Episode        int
	Name           string
	Overview       string
	Image          string
	Media          string
	AirDate        string
	ProductionCode string
	Vote           float32
}

// Generates metadata for everything.
func generateMetadata(paths Paths) {
	log.Println("Generating TV metadata")

	shows := make([]TVShowMetadata, 0)
	for _, showFolder := range directoriesIn(paths.TV) {
		// Load the metadata.
		var showDetails *TmdbTvShowDetails
		if err := readAndUnmarshal(showFolder, metadataFilename, &showDetails); err != nil {
			continue
		}

		// Create the model.
		image, _ := filepath.Rel(paths.Root, filepath.Join(showFolder, imageFilename))
		backdrop, _ := filepath.Rel(paths.Root, filepath.Join(showFolder, imageBackdropFilename))
		show := TVShowMetadata{
			TMDBId:       showDetails.Id,
			Name:         showDetails.Name,
			Overview:     showDetails.Overview,
			Image:        image,
			Backdrop:     backdrop,
			FirstAirDate: showDetails.FirstAirDate,
			LastAirDate:  showDetails.LastAirDate,
		}

		// Find the seasons.
		for _, seasonFolder := range directoriesIn(showFolder) {
			var seasonDetails *TmdbTvSeasonDetails
			if err := readAndUnmarshal(seasonFolder, metadataFilename, &seasonDetails); err != nil {
				continue
			}

			image, _ := filepath.Rel(paths.Root, filepath.Join(seasonFolder, imageFilename))
			season := TVSeasonMetadata{
				TMDBId:   seasonDetails.Id,
				Season:   seasonDetails.SeasonNumber,
				Name:     seasonDetails.Name,
				Overview: seasonDetails.Overview,
				Image:    image,
				AirDate:  seasonDetails.AirDate,
			}

			// Find the episodes.
			for _, epFolder := range directoriesIn(seasonFolder) {
				var epDetails *TmdbTvEpisodeDetails
				if err := readAndUnmarshal(epFolder, metadataFilename, &epDetails); err != nil {
					continue
				}

				image, _ := filepath.Rel(paths.Root, filepath.Join(epFolder, imageFilename))
				media, _ := filepath.Rel(paths.Root, filepath.Join(epFolder, hlsFilename))
				episode := TVEpisodeMetadata{
					TMDBId:         epDetails.Id,
					Episode:        epDetails.EpisodeNumber,
					Name:           epDetails.Name,
					Overview:       epDetails.Overview,
					Image:          image,
					Media:          media,
					AirDate:        epDetails.AirDate,
					ProductionCode: epDetails.ProductionCode,
					Vote:           epDetails.VoteAverage,
				}
				season.Episodes = append(season.Episodes, episode)
			}
			sort.Sort(ByEpisode(season.Episodes))

			show.Seasons = append(show.Seasons, season)
			generateSeasonHTML(seasonFolder, season, paths)
		}
		sort.Sort(BySeason(show.Seasons))
		generateShowHTML(showFolder, show, paths) {
	}

	// Make the root metadata.
	capacity := capacity(paths)
	metadata := Metadata{
		TVShows:  shows,
		Movies:   nil,
		Capacity: capacity,
	}

	// Save.
	data, _ := json.MarshalIndent(metadata, "", "    ")
	ioutil.WriteFile(filepath.Join(paths.Root, metadataFilename), data, os.ModePerm)
}

// List of full directories (eg path + name).
func directoriesIn(path string) []string {
	directories := make([]string, 0)
	infos, _ := ioutil.ReadDir(path)
	for _, info := range infos {
		if info.IsDir() {
			directory := filepath.Join(path, info.Name())
			directories = append(directories, directory)
		}
	}
}

func readAndUnmarshal(folder string, file string, v interface{}) error {
	path := filepath.Join(folder, file)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// Sorting.

type ByEpisode []TVEpisodeMetadata

func (a ByEpisode) Len() int      { return len(a) }
func (a ByEpisode) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByEpisode) Less(i, j int) bool {
	return a[i].Episode < a[j].Episode
}

type BySeason []TVSeasonMetadata

func (a BySeason) Len() int      { return len(a) }
func (a BySeason) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a BySeason) Less(i, j int) bool {
	return a[i].Season < a[j].Season
}

/// Generate a season's html (which is the list of episodes in it).
func generateSeasonHTML(seasonPath string, season TVSeasonMetadata, paths Paths) {
	html := htmlStart
	isLeft := true
	trOpen := false

	for _, episode := range season.Episodes {
		if isLeft {
			html += "<tr>"
			trOpen = true
		}

		linkPath, _ := filepath.Rel(seasonPath, filepath.Join(paths.Root, episode.Media))
		imagePath, _ := filepath.Rel(seasonPath, filepath.Join(paths.Root, episode.Image))
		name := fmt.Sprintf("S%02d E%02d", season.Season, episode.Episode, episode.Name)

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
	outPath := filepath.Join(seasonPath, indexHtml)
	ioutil.WriteFile(outPath, []byte(html), os.ModePerm)
}

/// Generate a show's html (which is the list of seasons in it).
func generateShowHTML(showPath string, show TVShowMetadata, paths Paths) {
	html := htmlStart
	isLeft := true
	trOpen := false

	for _, season := range show.Seasons {
		if isLeft {
			html += "<tr>"
			trOpen = true
		}

		linkPath := tvSeasonFolderNameFor(season.Season)
		imagePath := linkPath + "/" + imageFilename
		name := season.Name

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
	outPath := filepath.Join(seasonPath, indexHtml)
	ioutil.WriteFile(outPath, []byte(html), os.ModePerm)
}

/// Figure out how much disk space is left.
func capacity(paths Paths) string {
	cmd := exec.Command("df", "-h", paths.Root)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()
	// TODO split and transpose
	return strings.TrimSpace(out.String())
}
