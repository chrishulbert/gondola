package main

import (
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
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
	Poster       string
	TotalSeasons string
}

// Gets an overall series info.
func omdbRequestTVSeries(title string) (OmdbTVSeries, error) {
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

type OmdbTVEpisode struct {
	Title    string
	Year     string
	Rated    string
	Released string
	Season   string
	Episode  string
	Runtime  string
	Genre    string
	Director string
	Writer   string
	Actors   string
	Plot     string
	Country  string
	Poster   string // If this ends with '...foo@._V1_SX300.jpg', replace with '...foo@._V1_.jpg' for high res.
	ImdbID   string
}

// Gets an overall series info.
func omdbRequestTVEpisode(title string, season int, episode int) (OmdbTVEpisode, error) {
	// Hit the api.
	url := "http://www.omdbapi.com/?type=episode&t=" + url.QueryEscape(title) +
		"&season=" + strconv.Itoa(season) +
		"&episode=" + strconv.Itoa(episode)
	data, err := omdbDownload(url)
	if err != nil {
		return OmdbTVEpisode{}, err
	}

	// Check it has the 'response="true"' field.
	var r OmdbResponse
	if err := json.Unmarshal(data, &r); err != nil {
		return OmdbTVEpisode{}, err
	}
	if r.Response != "True" {
		return OmdbTVEpisode{}, errors.New("No match found in OMDB, you should either rename it or drop this file into 'other'.")
	}

	// Parse it.
	var t OmdbTVEpisode
	if err := json.Unmarshal(data, &t); err != nil {
		return OmdbTVEpisode{}, err
	}

	return t, nil
}
