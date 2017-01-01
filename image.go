package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
)

func imageForTitle(query string) ([]byte, error) {
	// Download the page.
	url := "http://www.imdb.com/find?ref_=nv_sr_fn&s=tt&q=" + url.QueryEscape(query)
	html, err := download(url)
	if err != nil {
		return nil, err
	}

	// Find the findList table.
	regex := regexp.MustCompile(`(?ims)<table class="findList.+?<\/table>`)
	// Regex modifiers: i=case insensitive, m=multiline, s=let . match \n
	findList := regex.FindString(string(html))
	if findList == "" {
		return nil, errors.New("Couldn't find findList in response")
	}

	// Find the first image url.
	imageRegex := regexp.MustCompile(`http[^"]+@+`)
	imageUrlBase := imageRegex.FindString(findList)
	if imageUrlBase == "" {
		return nil, errors.New("Couldn't find image in response")
	}

	// Download the image.
	return download(imageUrlBase + ".jpg")
}

// Downloads contents of a url.
func download(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// Given a poster link from OMDB, gets the high-res version.
func imageForPosterLink(link string) ([]byte, error) {
	r := regexp.MustCompile(`@.*\.jpg`)
	highResLink := r.ReplaceAllString(link, "@._V1_.jpg")
	return download(highResLink)
}

// Deprecated, use imageForPosterLink.
func deprecatedImageForIMDB(IMDBId string) ([]byte, error) {
	// Download the page.
	url := "http://www.imdb.com/title" + IMDBId
	html, err := download(url)
	if err != nil {
		return nil, err
	}

	// Find the mediaviewer id.
	mediaIdRegex := regexp.MustCompile(`mediaviewer\/(.+)\?`)
	mediaIdMatches := mediaIdRegex.FindStringSubmatch(string(html))
	if len(mediaIdMatches) < 2 {
		return nil, errors.New("Couldn't find mediaviewer id in response")
	}

	// Load the mediaviewer.
	mvUrl := "http://www.imdb.com/title/" + IMDBId + "/mediaviewer/" + mediaIdMatches[1]
	mvHtml, mvErr := download(mvUrl)
	if mvErr != nil {
		return nil, mvErr
	}

	// Get the image
	metaContentRegex := regexp.MustCompile(`<meta itemprop="image" content="(.+)\@`)
	metaContentMatches := metaContentRegex.FindStringSubmatch(string(mvHtml))
	if len(metaContentMatches) < 2 {
		return nil, errors.New("Couldn't find meta content in response")
	}

	// Download the image.
	return download(metaContentMatches[1] + "@._V1_.jpg")
}
