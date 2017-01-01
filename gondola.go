package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	validExtensions  = "mp4 mkv vob avi mpg m4v"
	imageFilename    = "image.jpg"
	metadataFilename = "metadata.json"
	hlsFilename      = "hls.m3u8"
)

// Returns true if it's an extension we're interested in.
func isValidExtension(extension string) bool {
	lowerExtension := strings.ToLower(extension)
	extensions := strings.Split(validExtensions, " ")
	for _, e := range extensions {
		if "."+e == lowerExtension {
			return true
		}
	}
	return false
}

// Scans the new paths, looking for any media files we're interested in.
func scanNewPaths(paths Paths, config Config) {
	scanNewPath(paths.NewMovies, true, paths, config)
	scanNewPath(paths.NewTV, false, paths, config)
}

func scanNewPath(whichPath string, isMovies bool, paths Paths, config Config) {
	files, err := ioutil.ReadDir(whichPath)
	if err != nil {
		log.Println("Couldn't scan path, error: ", err)
		return
	}

	for _, file := range files {
		if !strings.HasPrefix(file.Name(), ".") { // Ignore hidden files.
			if !file.IsDir() {
				ext := path.Ext(file.Name())
				if isValidExtension(ext) {
					log.Println("Found file", file.Name())
					tryProcess(whichPath, file.Name(), isMovies, paths, config)
				} else {
					log.Println("Ignoring file with unexpected extension", file.Name())
				}
			} else {
				log.Println("Unexpected, found a directory", file.Name())
			}
		}
	}
}

// Tries processing a file. Doesn't worry if it can't, eg if the file is half-copied, as the completion of the copy will trigger another scan.
func tryProcess(folder string, file string, isMovies bool, paths Paths, config Config) {
	source := filepath.Join(folder, file)
	if canGetExclusiveAccessToFile(source) {
		if isMovies {
			processMovie(folder, file, paths, config)
		} else {
			processTV(folder, file, paths, config)
		}

		// Re-generate.
		generateMetadata(paths)
	} else {
		log.Println("Couldn't get exclusive access to", file, "might be still copying")
	}
}

// Keeps track of where all the paths are.
type Paths struct {
	Root                 string // Config.Root (expanded path, no tilde)
	NewBase              string // Root/New
	NewMovies            string // Root/New/Movies
	NewTV                string // Root/New/TV
	Staging              string // Root/Staging
	MoviesRelativeToRoot string // Movies
	Movies               string // Root/Movies
	TVRelativeToRoot     string // TV
	TV                   string // Root/TV
	Failed               string // Root/Failed
}

func main() {
	config, configErr := loadConfig()
	if configErr != nil {
		log.Fatal(configErr)
	}
	log.Println("Config loaded:")
	log.Printf("%+v\n", config)

	// Figure out all the folders.
	var paths Paths
	paths.Root = expandTilde(config.Root)
	paths.NewBase = filepath.Join(paths.Root, "New")
	paths.NewMovies = filepath.Join(paths.NewBase, "Movies")
	paths.NewTV = filepath.Join(paths.NewBase, "TV")
	paths.Staging = filepath.Join(paths.Root, "Staging")
	paths.MoviesRelativeToRoot = "Movies"
	paths.Movies = filepath.Join(paths.Root, paths.MoviesRelativeToRoot)
	paths.TVRelativeToRoot = "TV"
	paths.TV = filepath.Join(paths.Root, paths.TVRelativeToRoot)
	paths.Failed = filepath.Join(paths.Root, "Failed")
	os.MkdirAll(paths.Root, os.ModePerm) // This will cause permission issues on a non-FAT mount eg local drive.
	os.MkdirAll(paths.NewMovies, os.ModePerm)
	os.MkdirAll(paths.NewTV, os.ModePerm)
	os.RemoveAll(paths.Staging) // Clear the staging folder on startup. Warning - this'll remove any in-progress source files. The idea is that those files vanish when complete anyway so this should be fine.
	os.MkdirAll(paths.Staging, os.ModePerm)
	os.MkdirAll(paths.Movies, os.ModePerm)
	os.MkdirAll(paths.TV, os.ModePerm)
	os.MkdirAll(paths.Failed, os.ModePerm)

	// When starting, re-gen metadata and scan for new files.
	generateMetadata(paths)
	scanNewPaths(paths, config)

	// Listen for changes on the folder.
	folders := []string{paths.NewMovies, paths.NewTV}
	changes := watch(folders)
	log.Println("Watching for changes in " + paths.NewBase)
	for {
		<-changes
		scanNewPaths(paths, config)
	}
}
