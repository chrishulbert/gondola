package main

import (
	"strconv"
)

type TmdbTvShowDetails struct {
	Id           int
	Overview     string
	BackdropPath string `json:"backdrop_path"`
	PosterPath   string `json:"poster_path"`
	FirstAirDate string `json:"first_air_date"` // yyyy-mm-dd
	LastAirDate  string `json:"last_air_date"`  // yyyy-mm-dd
	InProduction bool   `json:"in_production"`
	Homepage     string
	Genres       []TmdbTvShowGenre
}

type TmdbTvShowGenre struct {
	Id   int
	Name string
}

// Finds the tv show details.
func requestTmdbTVDetails(id int) (TmdbTvShowDetails, error) {
	url := tmdbApiRoot + "tv/" + strconv.Itoa(id) + "?api_key=" + tmdbApiKey
	var results TmdbTvShowDetails
	err := tmdbDownloadParse(url, &results)
	return results, err
}
