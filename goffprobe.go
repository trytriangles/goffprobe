package goffprobe

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Disposition struct {
	Default int `json:"default"`
	Dub int `json:"dub"`
	Original int `json:"original"`
	Comment int `json:"comment"`
	Lyrics int `json:"lyrics"`
	Karaoke int `json:"karaoke"`
	Forced int `json:"forced"`
	HearingImpaired int `json:"hearing_impaired"`
	VisualImpaired int `json:"visual_impaired"`
	CleanEffects int `json:"clean_effects"`
	AttachedPic int `json:"attached_pic"`
	TimedThumbnails int `json:"timed_thumbnails"`
}

type StreamTags struct {
	Language string `json:"language"`
	HandlerName string `json:"handler_name"`
}

type Stream struct {
	Index int `json:"index"`
	CodecName string `json:"codec_name"`
	CodecLongName string `json:"codec_long_name"`
	Profile string `json:"profile"`
	CodecType string `json:"codec_type"`
	CodecTimeBase string `json:"codec_time_base"`
	CodecTagString string `json:"codec_tag_string"`
	CodecTag string `json:"codec_tag"`
	Width int `json:"width,omitempty"`
	Height int `json:"height,omitempty"`
	// CodedWidth and CodedHeight explanation:
	// Some encoders require frame dimensions to be multiples of a certain
	// number e.g. x264 (16). The encoder will pad the frame to a suitable
	// number, if needed, and store the cropping values for the decoder. The
	// coded size is the size before cropping.
	CodedWidth int `json:"coded_width,omitempty"`
	CodedHeight int `json:"coded_height,omitempty"`
	ClosedCaptions int `json:"closed_captions,omitempty"`
	HasBFrames int `json:"has_b_frames,omitempty"`
	SampleAspectRatio string `json:"sample_aspect_ratio,omitempty"`
	DisplayAspectRatio string `json:"display_aspect_ratio,omitempty"`
	PixFmt string `json:"pix_fmt,omitempty"`
	Level int `json:"level,omitempty"`
	ChromaLocation string `json:"chroma_location,omitempty"`
	Refs int `json:"refs,omitempty"`
	IsAvc string `json:"is_avc,omitempty"`
	NalLengthSize string `json:"nal_length_size,omitempty"`
	RFrameRate string `json:"r_frame_rate"`
	AvgFrameRate string `json:"avg_frame_rate"`
	TimeBase string `json:"time_base"`
	StartPts int `json:"start_pts"`
	StartTime string `json:"start_time"`
	DurationTs int `json:"duration_ts"`
	Duration string `json:"duration"`
	BitRate string `json:"bit_rate"`
	BitsPerRawSample string `json:"bits_per_raw_sample,omitempty"`
	NbFrames string `json:"nb_frames"`
	Disposition Disposition `json:"disposition"`
	Tags StreamTags `json:"tags"`
	SampleFmt string `json:"sample_fmt,omitempty"`
	SampleRate string `json:"sample_rate,omitempty"`
	Channels int `json:"channels,omitempty"`
	ChannelLayout string `json:"channel_layout,omitempty"`
	BitsPerSample int `json:"bits_per_sample,omitempty"`
	MaxBitRate string `json:"max_bit_rate,omitempty"`
}

type FormatTags struct {
	MajorBrand string `json:"major_brand"`
	MinorVersion string `json:"minor_version"`
	CompatibleBrands string `json:"compatible_brands"`
	Encoder string `json:"encoder"`
}

type Format struct {
	Filename string `json:"filename"`
	NbStreams int `json:"nb_streams"`
	NbPrograms int `json:"nb_programs"`
	FormatName string `json:"format_name"`
	FormatLongName string `json:"format_long_name"`
	StartTime string `json:"start_time"`
	Duration string `json:"duration"`
	Size string `json:"size"`
	BitRate string `json:"bit_rate"`
	ProbeScore int `json:"probe_score"`
	Tags FormatTags `json:"tags"`
}

type ProbeResult struct {
	Streams []Stream `json:"streams"`
	Format Format `json:"format"`
}

type VideoInfo struct {
	ProbeResult ProbeResult `json:"probe_result"`
	VideoFormats []string `json:"video_formats"`
	VideoBitrates []uint64 `json:"video_bitrates"`
	AudioFormats []string `json:"audio_formats"`
	AudioBitrates []uint64 `json:"audio_bitrates"`
	Bitrate uint64 `json:"bitrate"`
	Duration int `json:"duration"`
	Filename string `json:"filename"`
	// If multiple video tracks present, pixel count of the highest resolution
	Pixels int `json:"pixels"` 
	Basename string `json:"basename"`
	AtTime int64 `json:"at_time"` // Epoch time of the probe
	HasMultipleAudio bool `json:"has_multiple_audio"`
	HasMultipleVideo bool `json:"has_multiple_video"`
	// SimpleDescription:
	// A string in the form 
	// "videoFormat+videoFormat@totalVideoBitrate/audioFormat+audioFormat@totalAudioBitrate"
	// e.g. "h264@4000000/aac@128000" or "h264+av1@7800000/aac+opus@64000"
	SimpleDescription string `json:"simple_description"`
	// If multiple video or audio formats are present. they will be joined
	// with '+'
	VideoFormat string `json:"video_format"`
	AudioFormat string `json:"audio_format"`
	// Combined total bitrates
	VideoBitrate uint64 `json:"video_bitrate"`
	AudioBitrate uint64 `json:"audio_bitrate"`
}

func (v *VideoInfo) calculateSimpleDescription() {
	v.VideoFormat = strings.Join(v.VideoFormats, "+")
	v.AudioFormat = strings.Join(v.AudioFormats, "+")
	v.AudioBitrate = sum(v.AudioBitrates...)
	v.VideoBitrate = sum(v.VideoBitrates...)
	v.SimpleDescription = fmt.Sprintf("%s@%d/%s@%d", v.VideoFormat, v.VideoBitrate, v.AudioFormat, v.AudioBitrate)
}

func (v *VideoInfo) calculatePixels() {
	for _, s := range v.ProbeResult.Streams {
		if s.CodecType != "video" {
			continue
		}
		pixels := s.Width * s.Height
		if pixels > v.Pixels {
			v.Pixels = pixels
		}
	}
}

func (v *VideoInfo) calculateDuration() {
	f, err := strconv.ParseFloat(v.ProbeResult.Format.Duration, 64)
	if err != nil {
		panic(err)
	}
	v.Duration = int(math.Ceil(f))
}

func (v *VideoInfo) calculateStreamInfo() {
	for _, s := range v.ProbeResult.Streams {
		if s.CodecType == "video" {
			bitrate, err := strconv.ParseUint(s.BitRate, 10, 64)
			if err != nil {
				panic(err)
			}
			v.Bitrate += bitrate 
			v.VideoBitrates = append(v.VideoBitrates, bitrate)
			v.VideoFormats = append(v.VideoFormats, s.CodecName)
		} else if s.CodecType == "audio" {
			bitrate, err := strconv.ParseUint(s.BitRate, 10, 64)
			if err != nil {
				panic(err)
			}
			v.Bitrate += bitrate
			v.AudioBitrates = append(v.AudioBitrates, bitrate)
			v.AudioFormats = append(v.AudioFormats, s.CodecName)
		}
	}
	v.HasMultipleAudio = len(v.AudioBitrates) > 1
	v.HasMultipleVideo = len(v.VideoBitrates) > 1
}

func (v *VideoInfo) WriteJSONFile(filename string, fileMode os.FileMode) error {
	jsonOutput, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, jsonOutput, fileMode)
}

func NewVideoInfo(filename string) *VideoInfo {
	var v VideoInfo
	v.Filename = filename
	v.Basename = filepath.Base(filename)
	v.ProbeResult = callffprobe(filename)
	v.AtTime = time.Now().Unix()
	v.calculateDuration()
	v.calculateStreamInfo()
	v.calculatePixels()
	v.calculateSimpleDescription()
	return &v
}

func sum(uints ...uint64) uint64 {
	var result uint64
	for _, n := range uints {
		if (math.MaxInt - n) < result {
			panic("Overflow while summing")
		}
		result += n
	}
	return result
}

func callffprobe(filename string) ProbeResult {
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", filename)
	output, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	var data ProbeResult
	err = json.Unmarshal(output, &data)
	if err != nil {
		panic(err)
	}
	return data
}
