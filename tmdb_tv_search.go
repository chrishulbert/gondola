package main

import (
	"errors"
	"net/url"
)

type TmdbTvSearchResults struct {
	Results []TmdbTvSearchResult
}

type TmdbTvSearchResult struct {
	Id int
}

// Finds the tv show id.
func requestTmdbTVSearch(title string) (int, error) {
	url := tmdbApiRoot + "search/tv?api_key=" + tmdbApiKey + "&query=" + url.QueryEscape(title)
	var results TmdbTvSearchResults
	if err := tmdbDownloadParse(url, &results); err != nil {
		return 0, err
	}

	// Return the first result.
	for _, result := range results.Results {
		return result.Id, nil
	}

	return 0, errors.New("No results")
}
