package main

import (
	"encoding/json"
	"errors"
	"net/url"
)

type OmdbTVSeries struct {
	Title        string
	Year         string
	Rated        string
	Runtime      string
	Genre        string
	Director     string
	Writer       string
	Actors       string
	Plot         string
	Country      string
	ImdbID       string
	TotalSeasons string
}

// Gets an overall series info.
func omdbRequestTVSeries(title string, season int, episode int) (OmdbTVSeries, error) {
	// Hit the api.
	url := "http://www.omdbapi.com/?type=series&t=" + url.QueryEscape(title)
	data, err := omdbDownload(url)
	if err != nil {
		return OmdbTVSeries{}, err
	}

	// Check it has the 'response="true"' field.
	var r OmdbResponse
	if err := json.Unmarshal(data, &r); err != nil {
		return OmdbTVSeries{}, err
	}
	if r.Response != "True" {
		return OmdbTVSeries{}, errors.New("No match found in OMDB, you should either rename it or drop this file into 'other'.")
	}

	// Parse it.
	var t OmdbTVSeries
	if err := json.Unmarshal(data, &t); err != nil {
		return OmdbTVSeries{}, err
	}

	return t, nil
}
