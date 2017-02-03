package main

import (
	"encoding/json"
	"os/exec"
)

// For JSON to parse, _ needs to be as-is, and they need to start caps in the structs.

type ProbeResult struct {
	Streams []ProbeStream
	Format  ProbeFormat
}

type ProbeStream struct {
	Avg_frame_rate  string //": "25/1",
	Bit_rate        string //": "192000",
	Bits_per_sample int    //": 0,
	Channel_layout  string // ": "stereo", "5.1(side)",
	// 5.1            FL+FR+FC+LFE+BL+BR
	// 5.1(side)      FL+FR+FC+LFE+SL+SR
	Channels        int    //": 2,
	Chroma_location string // ": "left",

	// video codec stuff:
	// Video: h264 (High)
	//      name: h264
	// long_name: H.264 / AVC / MPEG-4 AVC / MPEG-4 part 10
	//       is_avc=1
	//  profile=High

	// Video: mpeg2video (Main)
	// codec_name=mpeg2video
	// codec_long_name=MPEG-2 video
	// profile=Main

	// Video: h264 (High)
	// codec_name=h264
	// codec_long_name=H.264 / AVC / MPEG-4 AVC / MPEG-4 part 10
	// profile=High
	// is_avc=1

	// Video: h264 (High) (avc1 / 0x31637661),
	// codec_name=h264
	// codec_long_name=H.264 / AVC / MPEG-4 AVC / MPEG-4 part 10
	// profile=High
	// is_avc=1
	// codec_tag_string=avc1  <- !!! needs transcode
	// codec_tag=0x31637661  <- 1 c v a

	Codec_long_name      string //": "ATSC A/52A (AC-3)" "DVD subtitles" "MPEG-2 video",
	Codec_name           string // "ac3" "dvdsub" "mpeg2video",
	Codec_tag            string // "0x0000",
	Codec_tag_string     string // "[0][0][0][0]",
	Codec_time_base      string // "0/1" "1/48000" "1/50",
	Codec_type           string // "audio" "subtitle" "video",
	Coded_height         int    // 0,
	Coded_width          int    // 0,
	Color_range          string // "tv",
	Display_aspect_ratio string // "16:9",
	Dmix_mode            string // ": "-1",
	Duration             string // "2.033911",
	Duration_ts          int    // 183052,
	Has_b_frames         int    // 1,
	Height               int    // 576,
	Id                   string // "0x1e0",
	Index                int    // 0,
	Level                int    // 8,
	Loro_cmixlev         string // "-1.000000",
	Loro_surmixlev       string // "-1.000000",
	Ltrt_cmixlev         string // "-1.000000",
	Ltrt_surmixlev       string // "-1.000000",
	Max_bit_rate         string // "9800000",
	Pix_fmt              string // "yuv420p",
	Profile              string // "Main",
	R_frame_rate         string // "25/1",
	Refs                 int    // 1,
	Sample_aspect_ratio  string // "64:45",
	Sample_fmt           string // fltp",
	Sample_rate          string // "48000",
	Start_pts            int    // 25854,
	Start_time           string // "0.287267",
	Time_base            string // "1/90000",
	Timecode             string // "00:59:58:00",
	Width                int    // 720,
}

type ProbeFormat struct {
	Filename         string
	Nb_streams       int
	Nb_programs      int
	Format_name      string // ": "mpeg",
	Format_long_name string //": "MPEG-PS (MPEG-2 Program Stream)",
	Start_time       string // ": "0.287267",
	Duration         string //": "2.057911",
	Size             string //": "954906624",
	Probe_score      int    //": 52
}

// Probes a media file
func probe(path string) (*ProbeResult, error) {
	// Run
	out, runErr := exec.Command("ffprobe", "-v", "quiet", "-of", "json", "-show_format", "-show_streams", path).Output()
	if runErr != nil {
		return nil, runErr
	}

	// Parse
	var result ProbeResult
	parseErr := json.Unmarshal(out, &result)
	return &result, parseErr
}

func (r *ProbeResult) audioStreams() []ProbeStream {
	streams := make([]ProbeStream, 0)
	for _, stream := range r.Streams {
		if stream.Codec_type == "audio" {
			streams = append(streams, stream)
		}
	}
	return streams
}

func (r *ProbeResult) videoStreams() []ProbeStream {
	streams := make([]ProbeStream, 0)
	for _, stream := range r.Streams {
		if stream.Codec_type == "video" {
			streams = append(streams, stream)
		}
	}
	return streams
}
