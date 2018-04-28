package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Sanitises to make a filesystem-safe name.
func sanitiseForFilesystem(s string) string {
	s = strings.Replace(s, "/", "-", -1) // For linux.
	s = strings.Replace(s, "<", "-", -1) // For windows or FAT disks on linux.
	s = strings.Replace(s, ">", "-", -1)
	s = strings.Replace(s, ":", "-", -1)
	s = strings.Replace(s, "\"", "-", -1)
	s = strings.Replace(s, "\\", "-", -1)
	s = strings.Replace(s, "|", "-", -1)
	s = strings.Replace(s, "?", "-", -1)
	s = strings.Replace(s, "*", "-", -1)
	return s
}

func getMovieImageIfNeeded(image string, size string, folder string, filename string) {
	path := filepath.Join(folder, filename)
	if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
		image, imageErr := tmdbDownloadImage(image, size)
		if imageErr == nil {
			ioutil.WriteFile(path, image, os.ModePerm)
		} else {
			log.Println("Couldn't download image:", imageErr)
		}
	}
}

func getTVImageIfNeeded(image string, folder string, filename string) {
	path := filepath.Join(folder, filename)
	if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
		image, imageErr := vanillaDownload(image)
		if imageErr == nil || len(image) < 10000 {
			ioutil.WriteFile(path, image, os.ModePerm)
		} else {
			log.Println("Couldn't download image:", imageErr)
		}
	}
}

// Downloads contents of a url.
func vanillaDownload(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
