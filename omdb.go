package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

type OmdbMovie struct {
	Title    string
	Year     string
	Rated    string
	Runtime  string
	Genre    string
	Director string
	Actors   string
	Plot     string
}

type OmdbResponse struct {
	Response string
}

// Year can be nil.
func omdbRequest(title string, year *int) (OmdbMovie, error) {
	// Hit the api.
	url := "http://www.omdbapi.com/?t=" + url.QueryEscape(title)
	if year != nil {
		url += "&y=" + strconv.Itoa(*year)
	}
	data, err := omdbDownload(url)
	if err != nil {
		return OmdbMovie{}, err
	}

	// Check it has the 'response="true"' field.
	var r OmdbResponse
	if err := json.Unmarshal(data, &r); err != nil {
		return OmdbMovie{}, err
	}
	if r.Response != "True" {
		return OmdbMovie{}, errors.New("No matching movie found in OMDB, you should either rename it or drop this movie file into 'other' instead of 'movies'")
	}

	// Parse it.
	var m OmdbMovie
	if err := json.Unmarshal(data, &m); err != nil {
		return OmdbMovie{}, err
	}

	return m, nil
}

// Downloads contents of a url.
func omdbDownload(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
