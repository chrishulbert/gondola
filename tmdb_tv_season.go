package main

import (
	"strconv"
)

type TmdbTvSeasonDetails struct {
	Id           int
	AirDate      string `json:"air_date"`
	Name         string
	Overview     string
	PosterPath   string `json:"poster_path"`
	SeasonNumber int    `json:"season_number"`
	Episodes     []TmdbTvSeasonDetailsEpisode
}

type TmdbTvSeasonDetailsEpisode struct {
	EpisodeNumber int `json:"episode_number"`
	SeasonNumber  int `json:"season_number"`
	Name          string
	Overview      string
}

// Finds the tv show details for one season.
func requestTmdbTVSeason(id int, season int) (TmdbTvSeasonDetails, []byte, error) {
	url := tmdbApiRoot + "tv/" + strconv.Itoa(id) +
		"/season/" + strconv.Itoa(season) +
		"?api_key=" + tmdbApiKey
	var results TmdbTvSeasonDetails
	data, err := tmdbDownloadParse(url, &results)
	return results, data, err
}
