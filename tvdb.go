package main

import (
	"errors"
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/xrash/smetrics"
)

const baseURL = "https://www.thetvdb.com"
const posterBaseURL = "https://www.thetvdb.com/banners/posters/"

func chop(str string, start string, end string) string {
	indexA := strings.Index(str, start)
	if indexA < 0 {
		return ""
	}
	remainder := str[indexA+len(start):]
	indexB := strings.Index(remainder, end)
	if indexB < 0 {
		return ""
	}
	return remainder[:indexB]
}

/// Finds the last occurrence of 'start' and return the text after it.
func chopLast(str string, start string) string {
	index := strings.LastIndex(str, start)
	if index < 0 {
		return ""
	}
	return str[index+len(start):]
}

/// Finds the first occurrence of 'start' and return the text after it.
func chopFirst(str string, start string) string {
	index := strings.Index(str, start)
	if index < 0 {
		return ""
	}
	return str[index+len(start):]
}

func searchUrl(name string) string {
	u, _ := url.Parse(baseURL + "/search") // Assume it'll work.
	values := url.Values{}
	values.Add("q", name)
	u.RawQuery = values.Encode()
	return u.String()
}

func get(url string) string {
	resp, respErr := http.Get(url)
	if respErr != nil {
		return ""
	}

	defer resp.Body.Close()
	body, bodyErr := ioutil.ReadAll(resp.Body)
	if bodyErr != nil {
		return ""
	}

	return string(body)
}

/// Returns the show url slug / id. "" if can't find. eg 'i-dream-of-jeannie'.
func tvdbSearchForSeries(name string) string {
	resp := get(searchUrl(name))
	results := chop(resp, "<h2>Search Results</h2>", "</table>")
	regex := regexp.MustCompile(`<a href="/series/(.*?)">(.*?)</a>`)
	matches := regex.FindAllStringSubmatch(results, -1)
	closestDistance := 99999
	var closestSlug string
	for _, match := range matches {
		if len(match) >= 3 {
			slug := match[1]
			name := match[2]

			thisDistance := smetrics.WagnerFischer(
				strings.ToLower(name),
				strings.ToLower(name),
				1, 3, 2)

			if thisDistance < closestDistance {
				closestDistance = thisDistance
				closestSlug = slug
			}
		}
	}
	return closestSlug
}

func unescapeTrim(str string) string {
	return strings.TrimSpace(html.UnescapeString(str))
}

// Some of these fields below have json names for compatability with old TMDB metadata.

type TVDBSeries struct {
	TVDBID       int // Eg 55271 - thetvdb id
	Name         string
	Overview     string
	Art          string // The 16:9 background full url.
	Poster       string // The portrait dvd cover full url.
	FirstAirDate string `json:"first_air_date"`
	Seasons      []TVDBSeason
}

type TVDBSeason struct {
	TVDBID int    // Eg 1 - the season number (old tvdb used to make this different from season number)
	Season int    `json:"season_number"` // Eg 1,2,3, or 0 for specials
	Name   string // Eg 'Specials' or 'Season 1'
	/// Following are only filled in when requesting season details.
	Episodes []TVDBEpisode
	Image    string // Full url.
}

/// Finds all the seasons in an html for a tv series details page.
func seasonsForResponse(resp string) []TVDBSeason {
	section := chop(resp, "<h2>Seasons</h2>", "</div>")
	regex := regexp.MustCompile(`(?s)seasons/([0-9]+)">(.*?)</a>`) // s flag means . can span newlines.
	seasons := make([]TVDBSeason, 0)
	matches := regex.FindAllStringSubmatch(section, -1)
	for _, match := range matches {
		if len(match) >= 3 {
			number, _ := strconv.Atoi(match[1])
			name := unescapeTrim(match[2])

			seasons = append(seasons, TVDBSeason{
				TVDBID: number,
				Season: number,
				Name:   name,
			})
		}
	}
	return seasons
}

func tvdbSeriesDetails(id string) (TVDBSeries, error) {
	seriesUrl := baseURL + "/series/" + id
	resp := get(seriesUrl)
	// Name and details.
	nameRaw := chop(resp, "<title>", " @ TheTVDB</title>")
	name := unescapeTrim(nameRaw)
	detailsArea := chop(resp, `<div class="row">`, "</div>")
	detailsRaw := chop(detailsArea, `<p>`, `</p>`)
	details := unescapeTrim(detailsRaw)
	// The 16:9 background.
	artArea := chop(resp, "<h2>Backgrounds (Fan Art)</h2>", "</div>")
	artURL := chop(artArea, `src="`, `"`)
	// The portrait dvd cover.
	postersArea := chop(resp, "<h2>Posters</h2>", "</div>")
	posterURL := chop(postersArea, `src="`, `"`)
	firstAiredArea := chop(resp, `First Aired`, `</li>`)
	firstAirDate := chop(firstAiredArea, `<span>`, `</span>`)
	// The seasons.
	seasons := seasonsForResponse(resp)
	if name == "" || len(seasons) == 0 {
		return TVDBSeries{}, errors.New("Could not find series")
	}
	series := TVDBSeries{
		TVDBID:       id,
		Name:         name,
		Overview:     details,
		Art:          artURL,
		Poster:       posterURL,
		FirstAirDate: firstAirDate,
		Seasons:      seasons,
	}
	return series, nil
}

type TVDBEpisode struct {
	TVDBID       int
	SeasonNumber int
	Episode      int `json:"episode_number"`
	Name         string
	AirDate      string `json:"air_date"`
	// Following are empty until you ask for episode details.
	Overview string
	Image    string // full url.
}

func seasonUrl(seriesid int, seasonid int) string {
	u, _ := url.Parse(baseURL) // Assume it'll work.
	values := url.Values{}
	values.Add("seriesid", strconv.Itoa(seriesid))
	values.Add("seasonid", strconv.Itoa(seasonid))
	values.Add("tab", "season")
	u.RawQuery = values.Encode()
	return u.String()
}

/// Extracts the episodes from the season screen.
func episodesForSeasonDetails(resp string, seasonNumber int) []TVDBEpisode {
	section := chop(resp, `<td class="head">Episode Number</td>`, `</table>`)
	regex := regexp.MustCompile(`&id=([0-9]+).+?>([0-9]+)<.+?;lid=7">(.*)</a></td><td class=".+?">(.*?)</td>`)
	// regex := regexp.MustCompile(`&id=([0-9]+).+?>([0-9]+)<.+?;lid=7">(.+)</a></td><td class=".+?">(.+?)</td>`)
	episodes := make([]TVDBEpisode, 0)
	matches := regex.FindAllStringSubmatch(section, -1)
	for _, match := range matches {
		if len(match) >= 5 {
			id, _ := strconv.Atoi(match[1])
			number, _ := strconv.Atoi(match[2])
			name := unescapeTrim(match[3])
			date := unescapeTrim(match[4])

			if name != "" && number > 0 && id > 0 {
				episodes = append(episodes, TVDBEpisode{
					TVDBID:       id,
					SeasonNumber: seasonNumber,
					Episode:      number,
					Name:         name,
					AirDate:      date,
				})
			}
		}
	}
	return episodes
}

/// Get the episodes in a season.
func tvdbSeasonDetails(seriesid int, seasonId int, seasonNumber int) (TVDBSeason, error) {
	resp := get(seasonUrl(seriesid, seasonId))
	titleSection := chop(resp, `<div class="titlesection">`, `</table>`)
	titleRaw := chop(titleSection, `<h2>`, `<`)
	title := unescapeTrim(titleRaw) // Eg 'Season 2'
	bannersSection := chop(resp, `<h1>Season Banners</h1>`, `</div>`)
	banner := chop(bannersSection, `<td align=right><a href="`, `"`) // eg banners/seasons/90601-2-2.jpg
	episodes := episodesForSeasonDetails(resp, seasonNumber)
	if title == "" || len(episodes) == 0 {
		return TVDBSeason{}, errors.New("Could not find season")
	}
	season := TVDBSeason{
		TVDBID:   seasonId,
		Season:   seasonNumber,
		Name:     title,
		Image:    baseURL + `/` + banner,
		Episodes: episodes,
	}
	return season, nil
}

func episodeUrl(seriesid int, seasonid int, episodeid int) string {
	u, _ := url.Parse(baseURL) // Assume it'll work.
	values := url.Values{}
	values.Add("seriesid", strconv.Itoa(seriesid))
	values.Add("seasonid", strconv.Itoa(seasonid))
	values.Add("id", strconv.Itoa(episodeid))
	values.Add("tab", "episode")
	u.RawQuery = values.Encode()
	return u.String()
}

func tvdbEpisodeDetails(seriesid int, seasonId int, seasonNumber int, episodeid int) (TVDBEpisode, error) {
	resp := get(episodeUrl(seriesid, seasonId, episodeid))
	episodeNumberRaw := chop(resp, `<input type="text" name="EpisodeNumber" value="`, `"`)
	episodeNumber, _ := strconv.Atoi(episodeNumberRaw)
	titleSection := chop(resp, `<div class="titlesection">`, `</div>`)
	titleH3 := chop(titleSection, `<h3>`, `</h3>`)
	title := unescapeTrim(chopFirst(titleH3, `: `))
	airDate := chop(resp, `<input type="text" name="FirstAired" value="`, `"`)
	overview := unescapeTrim(chop(resp, `name="Overview_7" style="display: inline">`, `</textarea>`))
	images := chop(resp, `<h1>Episode Image</h1>`, `</div>`)
	image := chop(images, `<a href="`, `">`) // eg 'banners/episodes/90601/3038381.jpg'

	if title == "" || episodeNumber <= 0 {
		return TVDBEpisode{}, errors.New("Could not find episode")
	}

	episode := TVDBEpisode{
		TVDBID:       episodeid,
		SeasonNumber: seasonNumber,
		Episode:      episodeNumber,
		Name:         title,
		AirDate:      airDate,
		// Following are empty until you ask for episode details.
		Overview: overview,
		Image:    baseURL + `/` + image,
	}
	return episode, nil
}
