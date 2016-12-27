package main

import (
	"net/http"
	"net/url"
	"io/ioutil"
	"regexp"
	"errors"
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
