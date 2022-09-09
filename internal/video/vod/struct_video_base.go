package vod
// from https://github.com/aliyun/alibaba-cloud-sdk-go

// VideoBase is a nested struct in vod response
type VideoBase struct {
	CreationTime  string                     `json:"CreationTime" xml:"CreationTime"`
	Status        string                     `json:"Status" xml:"Status"`
	TranscodeMode string                     `json:"TranscodeMode" xml:"TranscodeMode"`
	OutputType    string                     `json:"OutputType" xml:"OutputType"`
	VideoID       string                     `json:"VideoId" xml:"VideoId"`
	CoverURL      string                     `json:"CoverURL" xml:"CoverURL"`
	Duration      string                     `json:"Duration" xml:"Duration"`
	Title         string                     `json:"Title" xml:"Title"`
	MediaType     string                     `json:"MediaType" xml:"MediaType"`
	DanMuURL      string                     `json:"DanMuURL" xml:"DanMuURL"`
}