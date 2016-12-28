package main

import (
	"gopkg.in/fsnotify.v1"
	"log"
	"strings"
	"time"
)

// Returns the channel that'll send you file changes.
// Folder cannot use ~
// Won't include subfolders - linux doesn't allow for this.
func watch(folders []string) chan string {
	changes := make(chan string)

	// This won't recursively watch a folder; this can't be done by any library on linux anyway.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Error starting watcher: ", err)
	}

	// Watch all the folders.
	for _, folder := range folders {
		err = watcher.Add(folder)
		if err != nil {
			log.Fatal("Error adding watcher: ", err)
		}
	}

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				// Ignore changes to hidden/system/ds_store files.
				if !strings.HasPrefix(event.Name, ".") {
					log.Println("Event: ", event)
					time.Sleep(time.Second) // Wait a bit before scanning the file, to give the transfer's write-lock a chance to take hold.
					changes <- event.Name
				} else {
					log.Println("Ignoring event for hidden/system file: ", event.Name)
				}
			case err := <-watcher.Errors:
				log.Println("Error: ", err)
			}
		}
	}()

	return changes
}
