package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

const (
	tmdbApiKey  = "3388d8b49bc28a62c0f373900ee4cbee"
	tmdbApiRoot = "https://api.themoviedb.org/3/"
)

func tmdbDownloadParse(url string, v interface{}) error {
	data, err := tmdbDownload(url)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &v)
}

// Downloads contents of a url.
func tmdbDownload(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
