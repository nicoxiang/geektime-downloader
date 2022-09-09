package crypto

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
)

// HmacSHA1Signature ...
func HmacSHA1Signature(accessKeySecret, stringToSign string) string {
	key := accessKeySecret + "&"
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
