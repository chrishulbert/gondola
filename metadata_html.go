package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
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
				<a href="LINK">
					<p><img src="IMAGE" /></p>
					<p>NAME</p>
				</a>
			</td>`
	htmlEnd = `
	</table>

	<h3>
		Gondola made with &hearts; by <a href="http://www.splinter.com.au">Chris Hulbert</a>
		<br />
		<a href="http://gondolamedia.com">gondolamedia.com</a>
		<br />
		<small><a href="javascript:alert('CAPACITY')">Usage</a></small>
		<br />
		<br />
		<small>This product uses the <a href="https://www.themoviedb.org">TMDb</a> API but is not endorsed or certified by TMDb.</small>
		<br />
		<img src="https://www.themoviedb.org/assets/9b3f9c24d9fd5f297ae433eb33d93514/images/v4/logos/408x161-powered-by-rectangle-green.png" 
			onload="this.width/=2;this.onload=null;" 
			alt="TMDb" />
	</h3>

</body>
</html>`
)

func generateMetadata(paths Paths) error {

	log.Println("Generating metadata html...")

	html := htmlStart
	isLeft := true
	trOpen := false

	// Movies.
	movieFiles, _ := ioutil.ReadDir(paths.Movies)
	for _, fileInfo := range movieFiles {
		if fileInfo.IsDir() {
			if isLeft {
				html += "<tr>"
				trOpen = true
			}

			linkPath := paths.MoviesRelativeToRoot + "/" + fileInfo.Name() + "/" + hlsFilename
			imagePath := paths.MoviesRelativeToRoot + "/" + fileInfo.Name() + "/" + imageFilename

			h := htmlTd
			h = strings.Replace(h, "LINK", linkPath, -1)
			h = strings.Replace(h, "IMAGE", imagePath, -1)
			h = strings.Replace(h, "NAME", fileInfo.Name(), -1)
			html += h

			if !isLeft {
				html += "</tr>"
				trOpen = false
			}

			isLeft = !isLeft
		}
	}

	// TV Shows.
	tvFiles, _ := ioutil.ReadDir(paths.TV)
	for _, fileInfo := range tvFiles {
		if fileInfo.IsDir() {
			if isLeft {
				html += "<tr>"
				trOpen = true
			}

			linkPath := paths.TVRelativeToRoot + "/" + fileInfo.Name()
			imagePath := paths.TVRelativeToRoot + "/" + fileInfo.Name() + "/" + imageFilename

			h := htmlTd
			h = strings.Replace(h, "LINK", linkPath, -1)
			h = strings.Replace(h, "IMAGE", imagePath, -1)
			h = strings.Replace(h, "NAME", fileInfo.Name(), -1)
			html += h

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

	// Add the html trailer.
	capacity := capacity(paths)
	end := strings.Replace(htmlEnd, "CAPACITY", capacity, -1)
	html += end

	// Save.
	outPath := filepath.Join(paths.Root, indexHtml)
	ioutil.WriteFile(outPath, []byte(html), os.ModePerm)

	log.Println("Successfully generated metadata html")

	generateRootMetadata(paths, capacity)

	return nil
}
