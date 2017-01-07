package main

import (
	"strconv"
)

type TmdbTvEpisodeDetails struct {
	Id             int
	Name           string
	Overview       string
	SeasonNumber   int    `json:"season_number"`
	EpisodeNumber  int    `json:"episode_number"`
	AirDate        string `json:"air_date"`
	ProductionCode string `json:"production_code"`
	StillPath      string `json:"still_path"`
}

// Finds the details for one episode.
func requestTmdbTVEpisode(id int, season int, episode int) (TmdbTvEpisodeDetails, []byte, error) {
	url := tmdbApiRoot + "tv/" + strconv.Itoa(id) +
		"/season/" + strconv.Itoa(season) +
		"/episode/" + strconv.Itoa(episode) +
		"?api_key=" + tmdbApiKey
	var results TmdbTvEpisodeDetails
	data, err := tmdbDownloadParse(url, &results)
	return results, data, err
}
