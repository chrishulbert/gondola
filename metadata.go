package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	indexHtml = "index.html"
	htmlStart = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<title>Gondola</title>
	<style>
	* {
		margin:0;
		padding:0;
		border:0;
		outline:0;
		font-family: "Helvetica Neue";
		border-collapse: collapse;
	}
	h1, h3 {
		padding: 0.3em;
		text-align: center;
	}
	table {
		width: 100%;
	}
	td {
		vertical-align: top;
		width: 50%;
		padding: 0.5em;
	}
	p {
		text-align: center;
		width: 100%;
	}
	img {
		width: 100%;
	}
	a {
		text-decoration: none;
	}
	</style>
</head>

<body>

	<h1>Gondola</h1>

	<table>`
	htmlTd = `<td>
				<a href="FOLDER/hls.m3u8">
					<p><img src="FOLDER/image.jpg" /></p>
					<p>NAME</p>
				</a>
			</td>`
	htmlEnd = `
	</table>

	<h3>
		Gondola made with &hearts; by <a href="http://www.splinter.com.au">Chris Hulbert</a>
		<br />
		<a href="http://gondolamedia.com">gondolamedia.com</a>
	</h3>

</body>
</html>`
)

func generateMetadata(paths Paths) error {

	log.Println("Generating metadata html...")

	html := htmlStart
	isLeft := true
	trOpen := false

	files, err := ioutil.ReadDir(paths.Movies)
	if err != nil {
		log.Println("Couldn't scan the folder to create the metadata")
		return err
	}

	for _, fileInfo := range files {
		if fileInfo.IsDir() {
			if isLeft {
				html += "<tr>"
				trOpen = true
			}
			relativePath := paths.MoviesRelativeToRoot + "/" + fileInfo.Name()
			html += strings.Replace(strings.Replace(htmlTd, "FOLDER", relativePath, -1), "NAME", fileInfo.Name(), -1)
			if !isLeft {
				html += "</tr>"
				trOpen = false
			}

			isLeft = !isLeft
		}
	}
	if trOpen {
		html += "</tr>"
	}

	html += htmlEnd

	// TODO tv.

	// Save.
	outPath := filepath.Join(paths.Root, indexHtml)
	ioutil.WriteFile(outPath, []byte(html), os.ModePerm)

	log.Println("Successfully generated metadata html")

	generateRootMetadata(paths)

	return nil
}

type RootMetadata struct {
	TVShows []interface{}
	Movies  []interface{}
}

// Generates the root metadata.json
func generateRootMetadata(paths Paths) {
	log.Println("Generating root metadata json...")

	var metadata RootMetadata

	// Load tv metadata.
	if data, err := ioutil.ReadFile(filepath.Join(paths.TV, metadataFilename)); err == nil {
		var tv []interface{}
		if err := json.Unmarshal(data, &tv); err == nil {
			metadata.TVShows = tv
		}
	}

	// Load movies metadata.
	if data, err := ioutil.ReadFile(filepath.Join(paths.Movies, metadataFilename)); err == nil {
		var m []interface{}
		if err := json.Unmarshal(data, &m); err == nil {
			metadata.Movies = m
		}
	}

	// Save.
	data, _ := json.MarshalIndent(metadata, "", "    ")
	outPath := filepath.Join(paths.Root, metadataFilename)
	ioutil.WriteFile(outPath, data, os.ModePerm)

	log.Println("Successfully generated root metadata json")
}
