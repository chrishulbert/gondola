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
	Response string
}

func omdbRequest(year int, title string) (OmdbMovie, error) {
	// Hit the api.
	url := "http://www.omdbapi.com/?y=" + strconv.Itoa(year) + "&r=json&t=" + url.QueryEscape(title)
	data, err := omdbDownload(url)
	if err != nil {
		return OmdbMovie{}, err
	}

	// Parse it.
	var m OmdbMovie
	if err := json.Unmarshal(data, &m); err != nil {
		return OmdbMovie{}, err
	}
	if m.Response != "True" {
		return OmdbMovie{}, errors.New("No matching movie found in OMDB, you should either rename it or drop this movie file into 'other' instead of 'movies'")
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
