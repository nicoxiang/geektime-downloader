package vod
// from https://github.com/aliyun/alibaba-cloud-sdk-go

// PlayInfoListInGetPlayInfo is a nested struct in vod response
type PlayInfoListInGetPlayInfo struct {
	PlayInfo []PlayInfo `json:"PlayInfo" xml:"PlayInfo"`
}