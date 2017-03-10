package main

import (
	"io/ioutil"
	"log"
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

func getImageIfNeeded(image string, size string, folder string, filename string) {
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
