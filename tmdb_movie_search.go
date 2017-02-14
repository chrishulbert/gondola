package main

import (
	"errors"
	"net/url"
	"strconv"
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
