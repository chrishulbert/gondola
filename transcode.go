package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
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
func convertToHLSAppropriately(inPath string, outFolder string, config Config) error {
	if config.DebugSkipHLS {
		// Skip conversion, this is good for debugging.
		log.Println("Not converting to HLS due to DebugSkipHLS flag")
		return nil
	}

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
	if len(videoStreams) == 0 {
		return errors.New("No video stream")
	}

	// Figure out which audio stream.
	var audioStream ProbeStream
	if len(audioStreams) == 0 {
		return errors.New("No audio stream")
	} else if len(audioStreams) == 1 {
		// Easy case, just one to choose from.
		audioStream = audioStreams[0]
	} else {
		// More than one audio. Either need to make the user choose, or take their choice.
		indexFromFilename := audioStreamFromFile(inPath)
		if indexFromFilename == nil {
			// User hasn't made a selection.
			log.Printf("Too many audio streams, splitting them out and forcing the user to choose one.")
			for _, stream := range audioStreams {
				args := []string{
					// "-ss", "60", // Start from 60s
					"-t", "180", // Only grab Xs
					"-i", inPath,
					"-map", fmt.Sprintf("0:%d", stream.Index),
					"-ac", "1", // Make it mono for speed and size.
					"-b:a", "64k", // CBR so it previews nicely on osx.
					inPath + fmt.Sprintf(".AudioStream%d preview.mp3", stream.Index),
				}
				ffmpeg(args) // TODO handle errors one day. This *should* work if probing succeeded earlier however.
			}
			// Rename it.
			ext := filepath.Ext(inPath) // Eg '.vob'
			nameSansExt := strings.TrimSuffix(inPath, ext)
			newName := nameSansExt + ".AudioStreamX" + ext + ".please insert correct audio stream number then remove this"
			os.Rename(inPath, newName)
			return &convertRenamedError{text: "Too many audio streams"}
		} else {
			for _, stream := range audioStreams {
				if stream.Index == *indexFromFilename {
					audioStream = stream
				}
			}

			// Did it find it?
			if audioStream.Index != *indexFromFilename {
				return errors.New("Couldn't find the stream with the index as per the filename")
			}
		}
	}

	// Figure out what to do with the audio.
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
	duration, _ := strconv.ParseFloat(probeResult.Format.Duration, 64)
	deinterlace := strings.Contains(inPath, "deinterlace")
	scaleAndCrop := strings.Contains(inPath, "scalecrop1080")
	crop1920_940Ratio := strings.Contains(inPath, "crop1920_940Ratio") // For 1920xshort (eg 800) inputs.
	scaleCrop1920_940 := strings.Contains(inPath, "scalecrop1920_940") // For eg 4k inputs.
	scaleCrop239Letterbox := strings.Contains(inPath, "scalecrop239letterbox1080")
	scaleCrop239Letterbox1920_940 := strings.Contains(inPath, "scalecrop239letterbox1920_940")
	crop240LetterboxThenUnivisium := strings.Contains(inPath, "crop240LetterboxThenUnivisium")
	crop235LetterboxThenUnivisium := strings.Contains(inPath, "crop235LetterboxThenUnivisium")
	crop240LetterboxThen169 := strings.Contains(inPath, "crop240LetterboxThen169")
	crop235LetterboxThen169 := strings.Contains(inPath, "crop235LetterboxThen169")
	isIncompatible := isIncompatiblePixelFormat(videoStream.Pix_fmt)
	var videoArgs []string
	if videoStream.Codec_name == "h264" && videoStream.Codec_tag_string != "avc1" && !isIncompatible && !deinterlace && !scaleAndCrop {
		// Can only direct copy if not avc1, or it won't be a seekable video.
		log.Println("Eligible for video not being transcoded, so no quality loss :)")
		videoArgs = []string{"-vcodec", "copy"}
	} else {
		log.Println("Video not eligible for muxing without transcoding.")
		if isIncompatible {
			log.Println("Video needs pixel format conversion")
			videoArgs = append(videoArgs, "-pix_fmt", "yuv420p")
		}
		if deinterlace {
			log.Println("Video is to be deinterlaced")
			videoArgs = append(videoArgs, "-vf", "yadif")
		}
		if scaleAndCrop {
			log.Println("Scale+crop to 1080p")
			videoArgs = append(videoArgs, "-vf", "scale=-1:1080,crop=1920:1080")
		}
		if crop1920_940Ratio {
			log.Println("Crop to 1920x940 ratio.")
			videoArgs = append(videoArgs, "-vf", "crop=ih/940*1920:ih")
		}
		if scaleCrop1920_940 {
			log.Println("Scale+crop to 1920x940")
			videoArgs = append(videoArgs, "-vf", "scale=-1:940,crop=1920:940")
		}
		if strings.Contains(inPath, "cropScaleDown4kWideToUnivisium") {
			// For if the input ratio is >= 1:2.1, crops down to 1:2 to fill the tv nicely, then scales to 1920w.
			videoArgs = append(videoArgs, "-vf", "crop=ih*2:ih,scale=1920:-1")
		}
		if strings.Contains(inPath, "scaleInside1920_1080MaintainingRatio") {
			// Both dimensions will be <= 1920x1080, while maintaining the aspect ratio.
			videoArgs = append(videoArgs, "-vf", "scale=1920:1080:force_original_aspect_ratio=decrease")
		}
		if scaleCrop239Letterbox {
			log.Println("Scale+crop+remove 1:2.39 letterbox bars to 1080p")
			videoArgs = append(videoArgs, "-vf", "crop=in_w:in_w/2.39,scale=-1:1080,crop=1920:1080")
		}
		if scaleCrop239Letterbox1920_940 {
			log.Println("Scale+crop+remove 1:2.39 letterbox bars to 1920x940")
			videoArgs = append(videoArgs, "-vf", "crop=iw:iw/2.39,scale=-1:940,crop=1920:940")
		}
		if crop240LetterboxThenUnivisium {
			log.Println("Crop out baked-in 1:2.40 letterbox bars, then crop again to univisium 1:2")
			videoArgs = append(videoArgs, "-vf", "crop=iw:iw/2.4,crop=ih*2:ih")
		}
		if crop235LetterboxThenUnivisium {
			log.Println("Crop out baked-in 1:2.35 letterbox bars, then crop again to univisium 1:2")
			videoArgs = append(videoArgs, "-vf", "crop=iw:iw/2.35,crop=ih*2:ih")
		}
		if strings.Contains(inPath, "crop235LetterboxThenUnivisiumThen1920") { // 1920/816 = 2.35
			log.Println("Crop out baked-in 1:2.35 letterbox bars, then crop that to univisium 1:2, then scale to 1920 (for 4k inputs)")
			videoArgs = append(videoArgs, "-vf", "crop=iw:iw/2.35,crop=ih*2:ih,scale=1920:-1")
		}
		if strings.Contains(inPath, "crop4k240LetterboxThenUnivisiumThen1920") {
			log.Println("Crop out baked-in 1:2.4 letterbox bars, then crop that to univisium 1:2, then scale to 1920 (for 4k inputs)")
			videoArgs = append(videoArgs, "-vf", "crop=iw:iw/2.4,crop=ih*2:ih,scale=1920:-1")
		}
		if crop240LetterboxThen169 {
			log.Println("Crop out baked-in 1:2.40 letterbox bars, then crop again to 16:9")
			videoArgs = append(videoArgs, "-vf", "crop=iw:iw/2.4,crop=ih*16/9:ih")
		}
		if crop235LetterboxThen169 {
			log.Println("Crop out baked-in 1:2.35 letterbox bars, then crop again to 16:9")
			videoArgs = append(videoArgs, "-vf", "crop=iw:iw/2.35,crop=ih*16/9:ih")
		}
		if strings.Contains(inPath, "crop240LetterboxDVDThenUnivisium") {
			log.Println("Cropping out baked-in 2:40 letterbox bars from a dvd with non-square pixels, then to univisium.")
			// For 2.4:1 DVDs with a non-square SAR eg 720x576 stretched to 16:9 [SAR 64:45 DAR 16:9], with letterbox bars baked in.
			// 720 is visually 1024 (720*64/45), so it's visually 1024x576 inclusive of black bars.
			// We want 1024/2.4 = 426 vertical pixels of that.
			// Then double the 426 to get the univisium displayed width of 852.
			// Then divide that by the SAR (852/64*45) to get 600.
			videoArgs = append(videoArgs, "-vf", "crop=600:426")
		}
	}

	return runConvertToHLS(
		inPath,
		outFolder,
		audioStream.Index,
		videoStream.Index,
		audioCommand,
		videoArgs,
		videoStream.Avg_frame_rate,
		duration,
		probeResult.hasSubtitles())
}

func isIncompatiblePixelFormat(pf string) bool {
	return strings.HasSuffix(pf, "9le") || strings.HasSuffix(pf, "9be") ||
		strings.HasSuffix(pf, "10le") || strings.HasSuffix(pf, "10be") ||
		strings.HasSuffix(pf, "12le") || strings.HasSuffix(pf, "12be") ||
		strings.HasSuffix(pf, "14le") || strings.HasSuffix(pf, "14be")
}

// Converts to HLS. If it gets back an error about h264_mp4toannexb, it retries with the appropriate command.
// frameRate is as per the probe eg "24000/1001"
func runConvertToHLS(inPath string, outFolder string, audioStreamIndex int, videoStreamIndex int, audioArgs []string, videoArgs []string, frameRateString string, duration float64, hasSubtitles bool) error {
	log.Printf("Converting to HLS with ffmpeg, audio: %+v; video: %+v\n", audioArgs, videoArgs)
	hlsHeaderPath := filepath.Join(outFolder, hlsFilename)
	var frameRate float64 = 60
	if strings.Contains(frameRateString, "/") {
		parts := strings.Split(frameRateString, "/")
		a, _ := strconv.ParseFloat(parts[0], 64)
		b, _ := strconv.ParseFloat(parts[1], 64)
		frameRate = a / b
	} else {
		frameRate, _ = strconv.ParseFloat(frameRateString, 64)
	}
	xStreamInfSuffix := ""
	headerSubsLine := ""
	if hasSubtitles {
		xStreamInfSuffix = ",SUBTITLES=\"subs\""
		headerSubsLine = "#EXT-X-MEDIA:TYPE=SUBTITLES,GROUP-ID=\"subs\",NAME=\"English\",DEFAULT=NO,AUTOSELECT=YES,FORCED=NO,LANGUAGE=\"en\",CHARACTERISTICS=\"public.accessibility.transcribes-spoken-dialog\",URI=\"subtitles.m3u8\"\n"
	}
	hlsHeaderContent := fmt.Sprintf("#EXTM3U\n%v#EXT-X-STREAM-INF:BANDWIDTH=1000000,FRAME-RATE=%f%v\n%s\n#EXT-X-ENDLIST", headerSubsLine, frameRate, xStreamInfSuffix, hlsSegmentsFilename)
	hlsHeaderErr := os.WriteFile(hlsHeaderPath, []byte(hlsHeaderContent), os.ModePerm)
	if hlsHeaderErr != nil {
		log.Println("Error writing hls header:", hlsHeaderErr)
		return hlsHeaderErr
	}

	// Write the subs m3u8.
	if hasSubtitles {
		durationInt := int(duration)
		subsContent := fmt.Sprintf("#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-TARGETDURATION:%d\n#EXT-X-MEDIA-SEQUENCE:0\n#EXTINF:%d,\nsubtitles.vtt\n#EXT-X-ENDLIST", durationInt, durationInt)
		subsPath := filepath.Join(outFolder, "subtitles.m3u8")
		os.WriteFile(subsPath, []byte(subsContent), os.ModePerm)

		// Extract the VTT.
		vttPath := filepath.Join(outFolder, "subtitles.vtt")
		args := []string{
			"-i", inPath, // Select the input file.
			"-map", "0:s:0", // Select first subtitles track only.
			vttPath,
		}
		result, err := ffmpeg(args)
		if err != nil {
			log.Println("Extracting subs failed, output was as follows, however I'll continue anyway:")
			log.Println(string(result))
		}
	}

	firstArgs := []string{
		"-i", inPath, // Select the input file.
		"-map", fmt.Sprintf("0:%d", videoStreamIndex), // Select the video stream. '0:v' would copy all video channels, but that's out of scope for this simple project.
		"-map", fmt.Sprintf("0:%d", audioStreamIndex), // 0:a would copy all audio channels, but iOS won't let you select channels from the stock media player.
	}
	hlsSegmentsPath := filepath.Join(outFolder, hlsSegmentsFilename)
	lastArgs := []string{"-hls_list_size", "0", hlsSegmentsPath}
	allArgs := append(append(append(firstArgs, audioArgs...), videoArgs...), lastArgs...)
	result, err := ffmpeg(allArgs)

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
		result2, err2 := ffmpeg(allArgs)

		// Print result if its an error.
		if err2 != nil {
			log.Println("Second ffmpeg attempt failed, the output was as follows:")
			log.Println(string(result2))
		}

		return err2
	}
	return err
}

// Runs FFMPEG, nicely, returning the stdout/stderr and any error.
func ffmpeg(args []string) (string, error) {
	nice := []string{"-n", "20", "ffmpeg"}
	allArgs := append(nice, args...)
	return execLog("nice", allArgs)
}
