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
)

const baseURL = "https://www.thetvdb.com"
const fanArtBaseURL = "https://www.thetvdb.com/banners/fanart/original/"
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
	u, _ := url.Parse(baseURL) // Assume it'll work.
	values := url.Values{}
	values.Add("string", name)
	values.Add("tab", "listseries")
	values.Add("function", "Search")
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

/// Returns <=0 if can't find.
func tvdbSearchForSeries(name string) int {
	resp := get(searchUrl(name))
	results := chop(resp, "<h1>TV Shows", "</table>")
	result := chop(results, "<tr><td class=\"odd\"><a href", "</td></tr>")
	idStr := chopLast(result, ">")
	id, _ := strconv.Atoi(idStr)
	return id
}

func seriesUrl(id int) string {
	u, _ := url.Parse(baseURL) // Assume it'll work.
	values := url.Values{}
	values.Add("id", strconv.Itoa(id))
	values.Add("tab", "series")
	u.RawQuery = values.Encode()
	return u.String()
}

func unescapeTrim(str string) string {
	return strings.TrimSpace(html.UnescapeString(str))
}

type TVDBSeries struct {
	TVDBID       int // Eg 55271 - thetvdb id
	Name         string
	Overview     string
	Art          string // The 16:9 background full url.
	Poster       string // The portrait dvd cover full url.
	FirstAirDate string
	Seasons      []TVDBSeason
}

type TVDBSeason struct {
	TVDBID int    // Eg 55271 - thetvdb id
	Season int    // Eg 1,2,3, or 0 for specials
	Name   string // Eg 'Specials' or 'Season 1'
	/// Following are only filled in when requesting season details.
	Episodes []TVDBEpisode
	Image    string // Full url.
}

/// Finds all the seasons in an html for a tv series details page.
func seasonsForResponse(resp string) []TVDBSeason {
	section := chop(resp, "<h1>Seasons</h1>", "</div>")
	regex := regexp.MustCompile(`seasonid=([0-9]+).*?class="seasonlink">(.+?)<`)
	seasons := make([]TVDBSeason, 0)
	matches := regex.FindAllStringSubmatch(section, -1)
	for _, match := range matches {
		if len(match) >= 3 {
			id, _ := strconv.Atoi(match[1])
			number, _ := strconv.Atoi(match[2])
			var name string
			if number == 0 {
				name = match[2]
			} else {
				name = "Season " + match[2]
			}

			seasons = append(seasons, TVDBSeason{
				TVDBID: id,
				Season: number,
				Name:   name,
			})
		}
	}
	return seasons
}

func tvdbSeriesDetails(id int) (TVDBSeries, error) {
	resp := get(seriesUrl(id))
	content := chop(resp, "<div id=\"content\">", "</div>")
	// Name and details.
	nameRaw := chop(content, "<h1>", "</h1>")
	name := unescapeTrim(nameRaw)
	detailsRaw := chopLast(content, "</h1>")
	details := unescapeTrim(detailsRaw)
	// The 16:9 background.
	art := chop(resp, "<h1>Fan Art</h1>", "</div>")
	artA := chop(art, "<a href=\"javascript:;\"", "</a>")
	artFile := chop(artA, "original/", "\"")
	artURL := fanArtBaseURL + artFile
	// The portrait dvd cover.
	posters := chop(resp, "<h1>Posters</h1>", "</div>")
	posterFile := chop(posters, "<a href=\"banners/posters/", "\" target=\"_blank\">View Full Size")
	posterURL := posterBaseURL + posterFile
	firstAirDate := chop(resp, `<input type="text" name="FirstAired" value="`, `"`)
	// The seasons.
	seasons := seasonsForResponse(resp)
	if name == "" || details == "" || len(seasons) == 0 {
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
	Episode      int
	Name         string
	AirDate      string
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
