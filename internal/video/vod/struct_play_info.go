package vod
// from https://github.com/aliyun/alibaba-cloud-sdk-go

// PlayInfo is a nested struct in vod response
type PlayInfo struct {
	Format           string `json:"Format" xml:"Format"`
	BitDepth         int    `json:"BitDepth" xml:"BitDepth"`
	NarrowBandType   string `json:"NarrowBandType" xml:"NarrowBandType"`
	Fps              string `json:"Fps" xml:"Fps"`
	Encrypt          int64  `json:"Encrypt" xml:"Encrypt"`
	Rand             string `json:"Rand" xml:"Rand"`
	StreamType       string `json:"StreamType" xml:"StreamType"`
	WatermarkID      string `json:"WatermarkId" xml:"WatermarkId"`
	Size             int64  `json:"Size" xml:"Size"`
	Definition       string `json:"Definition" xml:"Definition"`
	Plaintext        string `json:"Plaintext" xml:"Plaintext"`
	JobID            string `json:"JobId" xml:"JobId"`
	EncryptType      string `json:"EncryptType" xml:"EncryptType"`
	PreprocessStatus string `json:"PreprocessStatus" xml:"PreprocessStatus"`
	ModificationTime string `json:"ModificationTime" xml:"ModificationTime"`
	Bitrate          string `json:"Bitrate" xml:"Bitrate"`
	CreationTime     string `json:"CreationTime" xml:"CreationTime"`
	Height           int64  `json:"Height" xml:"Height"`
	Complexity       string `json:"Complexity" xml:"Complexity"`
	Duration         string `json:"Duration" xml:"Duration"`
	HDRType          string `json:"HDRType" xml:"HDRType"`
	Width            int64  `json:"Width" xml:"Width"`
	Status           string `json:"Status" xml:"Status"`
	Specification    string `json:"Specification" xml:"Specification"`
	PlayURL          string `json:"PlayURL" xml:"PlayURL"`
}