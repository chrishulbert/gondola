package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

const (
	tmdbApiKey    = "3388d8b49bc28a62c0f373900ee4cbee"
	tmdbApiRoot   = "https://api.themoviedb.org/3/"
	tmdbImageRoot = "http://image.tmdb.org/t/p/"
)

func tmdbDownloadParse(url string, v interface{}) ([]byte, error) {
	data, err := tmdbDownload(url)
	if err != nil {
		return nil, err
	}

	return data, json.Unmarshal(data, &v)
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

// Size options are:
//   "base_url": "http://image.tmdb.org/t/p/",
//   "secure_base_url": "https://image.tmdb.org/t/p/",
//   "backdrop_sizes": [
//     "w300",
//     "w780",
//     "w1280",
//     "original"
//   ],
//   "logo_sizes": [
//     "w45",
//     "w92",
//     "w154",
//     "w185",
//     "w300",
//     "w500",
//     "original"
//   ],
//   "poster_sizes": [
//     "w92",
//     "w154",
//     "w185",
//     "w342",
//     "w500",
//     "w780",
//     "original"
//   ],
//   "profile_sizes": [
//     "w45",
//     "w185",
//     "h632",
//     "original"
//   ],
//   "still_sizes": [
//     "w92",
//     "w185",
//     "w300",
//     "original"
//   ]
func tmdbDownloadImage(path string, size string) ([]byte, error) {
	return tmdbDownload(tmdbImageRoot + size + path)
}
