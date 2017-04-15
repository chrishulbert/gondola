package main

import (
	"bufio"
	"bytes"
	"log"
	"os/exec"
	"strings"
)

func scanCarriageReturns(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\r'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

/// Executes, logging lines as they come in, returning all stdin/err output.
func execLog(command string, args []string) (string, error) {
	cmd := exec.Command(command, args...)
	output := ""

	// Stdout.
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("StdoutPipe error:", err)
		return "", err
	}

	stdoutScanner := bufio.NewScanner(stdout)
	stdoutScanner.Split(scanCarriageReturns)
	go func() {
		for stdoutScanner.Scan() {
			text := stdoutScanner.Text()
			output = output + text
			log.Println(strings.TrimSpace(text))
		}
	}()

	// Stderr.
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Println("StderrPipe error:", err)
		return "", err
	}

	stderrScanner := bufio.NewScanner(stderr)
	stderrScanner.Split(scanCarriageReturns)
	go func() {
		for stderrScanner.Scan() {
			text := stderrScanner.Text()
			output = output + text
			log.Println(strings.TrimSpace(text))
		}
	}()

	// Run.
	err = cmd.Start()
	if err != nil {
		log.Println("Start error:", err)
		return "", err
	}

	err = cmd.Wait()
	return output, err
}
