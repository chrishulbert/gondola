package main

import (
	"strconv"
)

type TmdbTvShowDetails struct {
	Id           int
	Name         string
	Overview     string
	BackdropPath string `json:"backdrop_path"`
	PosterPath   string `json:"poster_path"`
	FirstAirDate string `json:"first_air_date"` // yyyy-mm-dd
	LastAirDate  string `json:"last_air_date"`  // yyyy-mm-dd
	InProduction bool   `json:"in_production"`
	Homepage     string
	Genres       []TmdbTvShowGenre
	Seasons      []TmdbTvShowSeason
}

type TmdbTvShowGenre struct {
	Id   int
	Name string
}

type TmdbTvShowSeason struct {
	Id           int
	PosterPath   string `json:"poster_path"`
	SeasonNumber int    `json:"season_number"`
}

// Finds the tv show details.
func requestTmdbTVShowDetails(id int) (TmdbTvShowDetails, []byte, error) {
	url := tmdbApiRoot + "tv/" + strconv.Itoa(id) + "?api_key=" + tmdbApiKey
	var results TmdbTvShowDetails
	data, err := tmdbDownloadParse(url, &results)
	return results, data, err
}
