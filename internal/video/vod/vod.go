package vod

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	pc "github.com/nicoxiang/geektime-downloader/internal/pkg/crypto"
)

var (
	// PlayAuthSign1 ...
	PlayAuthSign1 = []int{52, 58, 53, 121, 116, 102}
	// PlayAuthSign2 ...
	PlayAuthSign2 = []int{90, 91}
)

type PlayAuthData struct {
	SecurityToken   string `json:"SecurityToken"`
	AuthInfo        string `json:"AuthInfo"`
	AccessKeyID     string `json:"AccessKeyId"`
	AccessKeySecret string `json:"AccessKeySecret"`
}

// BuildVodGetPlayInfoURL ...
func BuildVodGetPlayInfoURL(playAuth, videoID, clientRand string) (string, error) {
	decodedPlayAuth := decodePlayAuth(playAuth)
	var playAuthData PlayAuthData
	err := json.Unmarshal([]byte(decodedPlayAuth), &playAuthData)
	if err != nil {
		return "", err
	}

	encryptedClientRand, err := pc.RSAEncrypt([]byte(clientRand))
	if err != nil {
		return "", err
	}

	publicParams := map[string]string{}
	publicParams["AccessKeyId"] = playAuthData.AccessKeyID
	publicParams["SignatureMethod"] = "HMAC-SHA1"
	publicParams["SignatureVersion"] = "1.0"
	publicParams["SignatureNonce"] = uuid.NewString()
	publicParams["Format"] = "JSON"
	publicParams["Channel"] = "HTML5"
	publicParams["StreamType"] = "video"
	publicParams["Rand"] = encryptedClientRand
	publicParams["Formats"] = ""
	publicParams["Version"] = "2017-03-21"

	privateParams := map[string]string{}
	privateParams["Action"] = "GetPlayInfo"
	privateParams["AuthInfo"] = playAuthData.AuthInfo
	privateParams["AuthTimeout"] = "7200"
	privateParams["PlayConfig"] = "{}"
	privateParams["PlayerVersion"] = "2.8.2"
	privateParams["ReAuthInfo"] = "{}"
	privateParams["SecurityToken"] = playAuthData.SecurityToken
	privateParams["VideoId"] = videoID
	allParams := getAllParams(publicParams, privateParams)
	cqs := getCQS(allParams)
	stringToSign := "GET" + "&" + percentEncode("/") + "&" + percentEncode(cqs)
	accessKeySecret := playAuthData.AccessKeySecret
	signature := pc.HmacSHA1Signature(accessKeySecret, stringToSign)
	queryString := cqs + "&Signature=" + percentEncode(signature)
	return "https://vod.cn-shanghai.aliyuncs.com/?" + queryString, nil
}

func decodePlayAuth(playAuth string) string {
	if isSignedPlayAuth(playAuth) {
		playAuth = decodeSignedPlayAuth2B64(playAuth)
	}
	data, err := base64.StdEncoding.DecodeString(playAuth)
	if err != nil {
		return ""
	}
	return string(data)
}

func isSignedPlayAuth(playAuth string) bool {
	signPos1 := time.Now().Year() / 100
	signPos2 := len(playAuth) - 2
	sign1 := getSignStr(PlayAuthSign1)
	sign2 := getSignStr(PlayAuthSign2)
	r1 := playAuth[signPos1 : signPos1+len(sign1)]
	r2 := playAuth[signPos2:]
	return sign1 == r1 && r2 == sign2
}

func decodeSignedPlayAuth2B64(playAuth string) string {
	sign1 := getSignStr(PlayAuthSign1)
	sign2 := getSignStr(PlayAuthSign2)
	playAuth = strings.Replace(playAuth, sign1, "", 1)
	playAuth = playAuth[:len(playAuth)-len(sign2)]
	factor := time.Now().Year() / 100
	newCharCodeList := []byte(playAuth)
	for i, code := range newCharCodeList {
		r := int(code) / factor
		z := factor / 10
		if r == z {
			newCharCodeList[i] = code
		} else {
			newCharCodeList[i] = code - 1
		}
	}
	return string(newCharCodeList)
}

func getSignStr(sign []int) string {
	s := strings.Builder{}
	for i, b := range sign {
		s.WriteByte(byte(b - i))
	}
	return s.String()
}

func percentEncode(s string) string {
	return url.QueryEscape(s)
}

func getCQS(allParams []string) string {
	sort.Strings(allParams)
	return strings.Join(allParams, "&")
}

func getAllParams(publicParams, privateParams map[string]string) (allParams []string) {
	for key, value := range publicParams {
		allParams = append(allParams, percentEncode(key)+"="+percentEncode(value))
	}
	for key, value := range privateParams {
		allParams = append(allParams, percentEncode(key)+"="+percentEncode(value))
	}
	return allParams
}
