package main

import (
	"os"
	"log"
	"strings"
	"io/ioutil"
	"path/filepath"
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

	log.Println("Generating metadata files...")

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

	log.Println("Successfully generated metadata files")

	return nil
}
