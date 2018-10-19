package main

import (
	"errors"
	"html"
	"io/ioutil"
	"log"
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
	log.Println("tvdbSearchForSeries got # results:", len(matches))
	for _, match := range matches {
		if len(match) >= 3 {
			slug := match[1]
			thisName := match[2]

			thisDistance := smetrics.WagnerFischer(
				strings.ToLower(thisName),
				strings.ToLower(name),
				1, 3, 2)

			log.Println("Possibility, id:", slug, ", name:", thisName, "; score:", thisDistance)

			if thisDistance < closestDistance {
				closestDistance = thisDistance
				closestSlug = slug
			}
		}
	}
	log.Println("Best match", closestSlug)
	return closestSlug
}

func unescapeTrim(str string) string {
	return strings.TrimSpace(html.UnescapeString(str))
}

// Some of these fields below have json names for compatability with old TMDB metadata.

type TVDBSeries struct {
	TVDBID       string // Eg 'i-dream-of-jeannie' - thetvdb url slug / id
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
	Image        string // full url.
	// Following are empty until you ask for episode details.
	Overview string
}

/// Extracts the episodes from the season screen.
func episodesForSeasonDetails(resp string, seasonNumber int) []TVDBEpisode {
	section := chop(resp, `<div class="col-xs-12 col-sm-8 episodes">`, `</table>`)
	regex := regexp.MustCompile(`(?s)<tr>.*?/episodes/([0-9]+)">.*?([0-9]+).*?<span.*?>(.*?)</span>.*?<td>.*?([0-9][0-9])/([0-9][0-9])/([0-9][0-9][0-9][0-9]).*?</td>.*?<td>.*?data-featherlight="(.*?)".*?</td>.*?</tr>`)
	episodes := make([]TVDBEpisode, 0)
	matches := regex.FindAllStringSubmatch(section, -1)
	for _, match := range matches {
		if len(match) >= 8 {
			id, _ := strconv.Atoi(match[1])
			number, _ := strconv.Atoi(match[2])
			name := unescapeTrim(match[3])
			mm := match[4]   // mm
			dd := match[5]   // dd
			yyyy := match[6] // yyyy
			date := yyyy + "-" + mm + "-" + dd
			image := match[7]

			if name != "" && number > 0 && id > 0 {
				episodes = append(episodes, TVDBEpisode{
					TVDBID:       id,
					SeasonNumber: seasonNumber,
					Episode:      number,
					Name:         name,
					AirDate:      date,
					Image:        image,
				})
			}
		}
	}
	return episodes
}

/// Get the episodes in a season.
func tvdbSeasonDetails(seriesid string, seasonId int, seasonNumber int) (TVDBSeason, error) {
	seasonUrl := baseURL + "/series/" + seriesid + "/seasons/" + strconv.Itoa(seasonId)
	resp := get(seasonUrl)

	titleSection := chop(resp, `<title>`, `</title>`)
	titleRaw := chop(titleSection, ` - `, ` @ TheTVDB`)
	title := unescapeTrim(titleRaw) // Eg 'Season 2'

	postersSection := chop(resp, `<h2>Posters</h2>`, `</div>`)
	posterURL := chop(postersSection, `src="`, `"`)

	episodes := episodesForSeasonDetails(resp, seasonNumber)

	if title == "" {
		return TVDBSeason{}, errors.New("Could not find season")
	}

	season := TVDBSeason{
		TVDBID:   seasonId,
		Season:   seasonNumber,
		Name:     title,
		Image:    posterURL,
		Episodes: episodes,
	}
	return season, nil
}

func tvdbEpisodeDetails(seriesId string, seasonId int, seasonNumber int, episodeid int) (TVDBEpisode, error) {
	episodeUrl := baseURL + "/series/" + seriesId + "/episodes/" + strconv.Itoa(episodeid)
	resp := get(episodeUrl)

	episodeNumberArea := chop(resp, `<strong>Episode Number</strong>`, `</li>`)
	episodeNumberRaw := chop(episodeNumberArea, `<span>`, `</span>`)
	episodeNumber, _ := strconv.Atoi(episodeNumberRaw)

	titleArea := chop(resp, `<title>`, `</title>`)
	titleRaw := chop(titleArea, ` - `, ` @ TheTVDB`)
	title := unescapeTrim(titleRaw)

	airArea := chop(resp, `<strong>Originally Aired</strong>`, `</li>`)
	airDate := chop(airArea, `<span>`, `</span>`) // Eg Saturday, September 18, 1965

	overviewArea := chop(resp, `<div class="block" id="translations">`, `</div>`)
	overview := unescapeTrim(chop(overviewArea, `<p>`, `</p>`))

	imageArea := chop(resp, `<div class="screenshot">`, `</div>`)
	image := chop(imageArea, `src="`, `"`)

	if title == "" || episodeNumber <= 0 {
		return TVDBEpisode{}, errors.New("Could not find episode")
	}

	episode := TVDBEpisode{
		TVDBID:       episodeid,
		SeasonNumber: seasonNumber,
		Episode:      episodeNumber,
		Name:         title,
		AirDate:      airDate,
		Overview:     overview,
		Image:        image,
	}
	return episode, nil
}
