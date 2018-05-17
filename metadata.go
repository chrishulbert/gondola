package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type Metadata struct {
	TVShows  []TVShowMetadata
	Movies   []MovieMetadata
	Capacity string
}

type MovieMetadata struct {
	TMDBId      int
	Name        string
	Overview    string
	Image       string
	Backdrop    string
	ReleaseDate string
	Vote        float32
	Media       string
}

type TVShowMetadata struct {
	TVDBId       string
	Name         string
	Overview     string
	Image        string
	Backdrop     string
	FirstAirDate string
	Seasons      []TVSeasonMetadata
}

type TVSeasonMetadata struct {
	TVDBId   int
	Season   int
	Name     string
	Image    string
	Episodes []TVEpisodeMetadata
}

type TVEpisodeMetadata struct {
	TVDBId   int
	Episode  int
	Name     string
	Overview string
	Image    string
	Media    string
	AirDate  string
}

// Generates metadata for everything.
func generateMetadata(paths Paths) {
	log.Println("Generating metadata")

	shows := make([]TVShowMetadata, 0)
	for _, showFolder := range directoriesIn(paths.TV) {

		// Load the metadata.
		var showDetails *TVDBSeries
		if err := readAndUnmarshal(showFolder, metadataFilename, &showDetails); err != nil {
			continue
		}

		// Create the model.
		image, _ := filepath.Rel(paths.Root, filepath.Join(showFolder, imageFilename))
		backdrop, _ := filepath.Rel(paths.Root, filepath.Join(showFolder, imageBackdropFilename))
		show := TVShowMetadata{
			TVDBId:       showDetails.TVDBID,
			Name:         showDetails.Name,
			Overview:     showDetails.Overview,
			Image:        image,
			Backdrop:     backdrop,
			FirstAirDate: showDetails.FirstAirDate,
		}

		// Find the seasons.
		for _, seasonFolder := range directoriesIn(showFolder) {
			var seasonDetails *TVDBSeason
			if err := readAndUnmarshal(seasonFolder, metadataFilename, &seasonDetails); err != nil {
				continue
			}

			image, _ := filepath.Rel(paths.Root, filepath.Join(seasonFolder, imageFilename))
			season := TVSeasonMetadata{
				TVDBId: seasonDetails.TVDBID,
				Season: seasonDetails.Season,
				Name:   seasonDetails.Name,
				Image:  image,
			}

			// Find the episodes.
			for _, epFolder := range directoriesIn(seasonFolder) {
				var epDetails *TVDBEpisode
				if err := readAndUnmarshal(epFolder, metadataFilename, &epDetails); err != nil {
					continue
				}

				image, _ := filepath.Rel(paths.Root, filepath.Join(epFolder, imageFilename))
				media, _ := filepath.Rel(paths.Root, filepath.Join(epFolder, hlsFilename))
				episode := TVEpisodeMetadata{
					TVDBId:   epDetails.TVDBID,
					Episode:  epDetails.Episode,
					Name:     epDetails.Name,
					Overview: epDetails.Overview,
					Image:    image,
					Media:    media,
					AirDate:  epDetails.AirDate,
				}
				season.Episodes = append(season.Episodes, episode)
			}
			sort.Sort(ByEpisode(season.Episodes))

			show.Seasons = append(show.Seasons, season)
			generateSeasonHTML(seasonFolder, season, paths)
		}
		sort.Sort(BySeason(show.Seasons))

		shows = append(shows, show)
		generateShowHTML(showFolder, show, paths)
	}

	// Find the movies.
	movies := make([]MovieMetadata, 0)
	for _, folder := range directoriesIn(paths.Movies) {
		// Load the metadata.
		var details *TmdbMovieSearchResult
		if err := readAndUnmarshal(folder, metadataFilename, &details); err != nil {
			continue
		}

		// Create the model.
		image, _ := filepath.Rel(paths.Root, filepath.Join(folder, imageFilename))
		backdrop, _ := filepath.Rel(paths.Root, filepath.Join(folder, imageBackdropFilename))
		media, _ := filepath.Rel(paths.Root, filepath.Join(folder, hlsFilename))
		movie := MovieMetadata{
			TMDBId:      details.Id,
			Name:        details.Title,
			Overview:    details.Overview,
			Image:       image,
			Backdrop:    backdrop,
			ReleaseDate: details.ReleaseDate,
			Vote:        details.VoteAverage,
			Media:       media,
		}
		movies = append(movies, movie)
	}

	// Make the root metadata.
	capacity := capacity(paths)
	metadata := Metadata{
		TVShows:  shows,
		Movies:   movies,
		Capacity: capacity,
	}

	// Save.
	data, _ := json.MarshalIndent(metadata, "", "    ")
	ioutil.WriteFile(filepath.Join(paths.Root, metadataFilename), data, os.ModePerm)

	generateRootHTML(capacity, paths)
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
	return directories
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
		name := fmt.Sprintf("S%02d E%02d %s", season.Season, episode.Episode, episode.Name)

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

/// Generates the root index.html
func generateRootHTML(capacity string, paths Paths) {
	html := htmlStart
	isLeft := true
	trOpen := false

	// Movies.
	movieFiles, _ := ioutil.ReadDir(paths.Movies)
	for _, fileInfo := range movieFiles {
		if fileInfo.IsDir() {
			if isLeft {
				html += "<tr>"
				trOpen = true
			}

			linkPath := paths.MoviesRelativeToRoot + "/" + fileInfo.Name() + "/" + hlsFilename
			imagePath := paths.MoviesRelativeToRoot + "/" + fileInfo.Name() + "/" + imageFilename

			h := htmlTd
			h = strings.Replace(h, "LINK", linkPath, -1)
			h = strings.Replace(h, "IMAGE", imagePath, -1)
			h = strings.Replace(h, "NAME", fileInfo.Name(), -1)
			html += h

			if !isLeft {
				html += "</tr>"
				trOpen = false
			}

			isLeft = !isLeft
		}
	}

	// TV Shows.
	tvFiles, _ := ioutil.ReadDir(paths.TV)
	for _, fileInfo := range tvFiles {
		if fileInfo.IsDir() {
			if isLeft {
				html += "<tr>"
				trOpen = true
			}

			linkPath := paths.TVRelativeToRoot + "/" + fileInfo.Name()
			imagePath := paths.TVRelativeToRoot + "/" + fileInfo.Name() + "/" + imageFilename

			h := htmlTd
			h = strings.Replace(h, "LINK", linkPath, -1)
			h = strings.Replace(h, "IMAGE", imagePath, -1)
			h = strings.Replace(h, "NAME", fileInfo.Name(), -1)
			html += h

			if !isLeft {
				html += "</tr>"
				trOpen = false
			}

			isLeft = !isLeft
		}
	}

	if trOpen {
		html += "</tr>"
	}

	// Add the html trailer.
	end := strings.Replace(htmlEnd, "CAPACITY", capacity, -1)
	html += end

	// Save.
	outPath := filepath.Join(paths.Root, indexHtml)
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
	outPath := filepath.Join(showPath, indexHtml)
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
