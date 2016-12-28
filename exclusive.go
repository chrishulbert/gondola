package main

import (
	"bytes"
	"log"
	"os/exec"
	"strings"
)

// Can the given file be accessed exclusively eg no other process is still writing to it?
// The os.OpenFile trick didn't work IME when someone's SCP'ing a file across, so we're going nuclear with lsof.
// lsof needs to be made accessible via sudo sans password for the current user by running:
// On ubuntu: sudo visudo -f /etc/sudoers.d/lsof
// On osx:    sudo mkdir /private/etc/sudoers.d, sudo visudo -f /private/etc/sudoers.d/lsof
// And adding the following line (replace 'chris' with the user you'll run as)
// ubuntu: chris ALL = (root) NOPASSWD: /usr/bin/lsof
// osx:    chris ALL = (root) NOPASSWD: /usr/sbin/lsof
func canGetExclusiveAccessToFile(path string) bool {
	cmd := exec.Command("sudo", "lsof", "-Fal", path)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()

	outString := strings.TrimSpace(out.String())
	outLines := strings.Split(outString, "\n")

	for _, line := range outLines {
		if line == "aw" || line == "au" || line == "lw" || line == "lW" || line == "lu" {
			log.Println("Another process has a lock on this file - it's likely still being copied")
			return false // Someone has access/lock for writing/updating.
		}
	}

	log.Println("No other process has dibs on this file, likely the transfer has completed")
	return true
}
