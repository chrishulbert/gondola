package main

import (
	"bytes"
	"os"
	"os/exec"
	"log"
	"runtime"
	"fmt"
)

// Can the given file be accessed exclusively eg no other process is still writing to it?
// The os.OpenFile trick didn't work IME when someone's SCP'ing a file across, so we're going nuclear with lsof.
// lsof needs to be made accessible via sudo sans password for the current user by running:
// On ubuntu: sudo visudo -f /etc/sudoers.d/lsof
// On osx:    sudo mkdir /private/etc/sudoers.d, sudo visudo -f /private/etc/sudoers.d/lsof
// And adding the following line (replace 'chris' with the user you'll run as)
// ubuntu: chris ALL = (root) NOPASSWD: /usr/bin/lsof
// osx:    chris ALL = (root) NOPASSWD: /usr/sbin/lsof
// lsof returns the following:
// Missing file: error 1, lots of text
// File there, other is using it: status 0, 1+ line(s)
// File there, nobody using it: status 1, 0 lines
// On OSX, it might think the current process is touching it, so check for that.
func canGetExclusiveAccessToFile(path string) bool {
	// On OSX, it thinks the current process is touching the file as well, so for now i simply
	// return true and recommend you only install on linux.
	if runtime.GOOS == "darwin" {
		log.Println("On OSX, can't determine file exclusivity, assuming it is exclusive. You really should install this only on linux until this is fixed.")
		return true
	}

	cmd := exec.Command("sudo", "lsof", "-Fp", path)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()
	outString := out.String()

	// This doesn't quite work just yet.
	// For OSX, it may return "pX" where x is the current PID.
	os.Getpid()
	osxThisPidOnly := fmt.Sprintf("p%d", os.Getpid())
	// fmt.Println("outString >", outString, "<")
	// fmt.Println("osxThisPidOnly >", osxThisPidOnly, "<")

	return outString == "" || outString == osxThisPidOnly
}
