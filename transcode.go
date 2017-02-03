package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Custom error for 'couldn't transcode, but i renamed it, so don't move it to failed'.
type convertRenamedError struct {
	text string
}

func (e *convertRenamedError) Error() string {
	return "Something was wrong with this file that needs user intervention: " + e.text + ". Renamed the file so the user is forced to choose."
}

// Tries to convert the given video to hls.
func convertToHLSAppropriately(inPath string, outPath string, config Config) error {
	// Probe it to find out what needs doing.
	log.Println("Probing, this sometimes takes a while...")
	probeResult, probeErr := probe(inPath)
	if probeErr != nil {
		return errors.New("Couldn't probe " + inPath)
	}
	log.Printf("Probed, found %v streams", len(probeResult.Streams))
	log.Printf("Probe result: %+v", probeResult)

	// Find the streams
	audioStreams := probeResult.audioStreams()
	videoStreams := probeResult.videoStreams()
	if len(audioStreams) == 0 {
		return errors.New("No audio stream")
	}
	if len(videoStreams) == 0 {
		return errors.New("No video stream")
	}

	// Too many audio streams?
	if len(audioStreams) > 1 {
		TODO check if AudioStreamX exists in the filename?
		log.Printf("Too many audio streams, splitting them out and forcing the user to choose one.")
		for _, stream := range audioStreams {
			args := []string{
				"-ss", "60", // Start from 60s
				"-t", "60", // Only grab 60s
				"-i", inPath,
				"-map", fmt.Sprintf("0:%d", stream.Index),
				"-b:a", "128k", // CBR so it previews nicely on osx.
				inPath + fmt.Sprintf(".AudioStream%d preview.mp3", stream.Index),
			}
			exec.Command("ffmpeg", args...).CombinedOutput() // TODO handle errors one day. This *should* work if probing succeeded earlier however.
		}
		// Rename it.
		ext := filepath.Ext(inPath) // Eg '.vob'
		nameSansExt := strings.TrimSuffix(inPath, ext)
		newName := nameSansExt + ".AudioStreamX" + ext + ".please insert correct audio stream number then remove this"
		os.Rename(inPath, newName)
		return &convertRenamedError{text: "Too many audio streams"}
	}

	// Figure out what to do with the audio.
	audioStream := audioStreams[TODO choose the one as per the filename or the first]
	var audioCommand []string
	if audioStream.Channel_layout == "stereo" && audioStream.Codec_name == "aac" {
		audioCommand = []string{"-acodec", "copy"} // Best case, can leave as-is.
	} else if audioStream.Channel_layout == "stereo" {
		audioCommand = []string{"-strict", "experimental", "-b:a", "192k"} // Transcode, same channels.
	} else if audioStream.Channel_layout == "5.1" { // FL+FR+FC+LFE+BL+BR
		// Tweak the 5.1 conversion, as by default it is quiet and drops the subwoofer.
		log.Println("Using custom downmix from 5.1 to stereo that preserves bass and speech")
		audioCommand = []string{"-strict", "experimental", "-b:a", "192k", "-af", "pan=stereo|FL<FL+BL+FC+LFE|FR<FR+BR+FC+LFE"}
	} else if audioStream.Channel_layout == "5.1(side)" { // FL+FR+FC+LFE+SL+SR
		log.Println("Using custom downmix from 5.1 to stereo that preserves bass and speech")
		audioCommand = []string{"-strict", "experimental", "-b:a", "192k", "-af", "pan=stereo|FL<FL+SL+FC+LFE|FR<FR+SR+FC+LFE"}
	} else {
		log.Println("Using `-ac 2` due to unexpected channel layout:", audioStream.Channel_layout)
		audioCommand = []string{"-strict", "experimental", "-b:a", "192k", "-ac", "2"} // Lousy cover-all.
	}

	// Figure out what to do with the video.
	videoStream := videoStreams[0]
	var videoArgs []string
	if videoStream.Codec_name == "h264" && videoStream.Codec_tag_string != "avc1" {
		// Can only direct copy if not avc1, or it won't be a seekable video.
		log.Println("Eligible for video not being transcoded, so no quality loss :)")
		videoArgs = []string{"-vcodec", "copy"}
	} else if videoStream.Codec_name == "h264" && videoStream.Codec_tag_string == "avc1" {
		// Have to transcode avc1 or they can't seek when watching.
		log.Println("Video needs transcoding to be seekable HLS, because it's AVC1")
		videoArgs = nil
	} else {
		// Any other codec needs transcoding.
		log.Println("Video needs transcoding, original codec doesn't suit")
		videoArgs = nil
	}

	if config.DebugSkipHLS {
		// Skip conversion, this is good for debugging.
		log.Println("Not converting to HLS due to DebugSkipHLS flag")
		return nil
	}

	return runConvertToHLS(inPath, outPath, audioStream.Index, videoStream.Index, audioCommand, videoArgs)
}

/// Converts to HLS. If it gets back an error about h264_mp4toannexb, it retries with the appropriate command.
func runConvertToHLS(inPath string, outPath string, audioStreamIndex int, videoStreamIndex int, audioArgs []string, videoArgs []string) error {
	log.Printf("Converting to HLS with ffmpeg, audio: %+v; video: %+v\n", audioArgs, videoArgs)
	firstArgs := []string{
		"-i", inPath, // Select the input file.
		"-map", "0:v", // This copies all video channels, even though there's hopefully only one.
		"-map", "0:a", // This copies all audio channels, so if there's a commentary channel it won't copy only the commentary.
		// Can't copy subs, as ffmpeg segfaults with an error "Exactly one WebVTT stream is needed".
		// "-map", "0:s", // Copies all subs
		// "-c:s", "copy", // Copy subs, no transcode option for subs. It's a bit silly this needs to be specified.
	}
	lastArgs := []string{"-hls_list_size", "0", outPath}
	allArgs := append(append(append(firstArgs, audioArgs...), videoArgs...), lastArgs...)
	result, err := exec.Command("ffmpeg", allArgs...).CombinedOutput()

	// Print result if its an error.
	if err != nil {
		log.Println("Initial ffmpeg attempt failed, the output was as follows:")
		log.Println(string(result))
	}

	// Did it fail with the annex b issue? If so, retry.
	// You can't simply *always* have h264_mp4toannexb enabled, it fails if not needed.
	if err != nil && strings.Contains(string(result), "h264_mp4toannexb") {
		log.Println("Attempting to convert to HLS using h264_mp4toannexb option")
		allArgs := append(append(append(append(firstArgs, audioArgs...), videoArgs...), "-bsf:v", "h264_mp4toannexb"), lastArgs...)
		result2, err2 := exec.Command("ffmpeg", allArgs...).CombinedOutput()

		// Print result if its an error.
		if err2 != nil {
			log.Println("Second ffmpeg attempt failed, the output was as follows:")
			log.Println(string(result2))
		}

		return err2
	}
	return err
}
