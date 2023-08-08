package main

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

type TmdbMovieSearchResults struct {
	Results []TmdbMovieSearchResult
}

type TmdbMovieSearchResult struct {
	Id           int
	Title        string
	Overview     string
	BackdropPath string `json:"backdrop_path"`
	PosterPath   string `json:"poster_path"`
	ReleaseDate  string `json:"release_date"` // yyyy-mm-dd
	Homepage     string
	VoteAverage  float32 `json:"vote_average"`
}

// Finds the movie id.
// If year is given, it includes that in the criteria.
func requestTmdbMovieSearch(title string, year *int) (TmdbMovieSearchResult, error) {	
	if strings.HasPrefix(title, "!") { // Don't use TMDB.
		newTitleA := strings.TrimSpace(title[1:])
		newTitle := strings.Split(newTitleA, ".")[0] // Eg if it is '!Foo.deinterlace.audiostream2' this only returns the 'Foo'.
		actualYear := 2000
		if year != nil {
			actualYear = *year
		}
		yearStr := strconv.Itoa(actualYear)
		return TmdbMovieSearchResult{
			Title: newTitle,
			ReleaseDate: yearStr + "-01-01",
		}, nil
	}

	url := tmdbApiRoot + "search/movie?api_key=" + tmdbApiKey + "&query=" + url.QueryEscape(title)
	if year != nil {
		url += "&year=" + strconv.Itoa(*year)
	}

	var results TmdbMovieSearchResults
	if _, err := tmdbDownloadParse(url, &results); err != nil {
		return TmdbMovieSearchResult{}, err
	}

	// Return the first result.
	for _, result := range results.Results {
		return result, nil
	}

	return TmdbMovieSearchResult{}, errors.New("No results")
}
