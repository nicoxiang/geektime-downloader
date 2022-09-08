package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"fmt"
)

// GetAESDecryptKey get 'aliyun private encyption method' decrypt key
// * * r0 = cmMeyfzJWyZcSwyH //rand
// * * r1 = md5(rand)
// * * r1 = r1.substring(8,24)
// * * iv = base64(r1).getBytes();
// * * key1 = aes.decrypt(rnd,iv,iv)
// * * seed1 = md5(r0+key1)
// * * seed1 = seed1.substring(8,24)
// * * k2 = base64(seed1).getBytes();
// * * key2 = aes.decrypt(plain,k2,iv);
// * * result = hex.encodeHexStr(base64.decode(key2))
// @param cr client random string
// @param sr server response random string
// @param plainText server response plain text
func GetAESDecryptKey(cr, sr, plainText string) string {
	crMD5 := fmt.Sprintf("%x", md5.Sum([]byte(cr)))
	t1 := crMD5[8:24]
	iv := []byte(t1)
	sd, _ := base64.StdEncoding.DecodeString(sr)
	dc1 := AESDecryptCBC(sd, iv, iv)
	r2 := cr + string(dc1)
	r2MD5 := fmt.Sprintf("%x", md5.Sum([]byte(r2)))
	t2 := r2MD5[8:24]
	key2 := []byte(t2)
	pd, _ := base64.StdEncoding.DecodeString(plainText)
	d2c := AESDecryptCBC(pd, key2, iv)
	b, _ := base64.StdEncoding.DecodeString(string(d2c))
	return fmt.Sprintf("%x", b)
}

// AESDecryptCBC ...
func AESDecryptCBC(encrypted, key, iv []byte) (decrypted []byte) {
	block, _ := aes.NewCipher(key)
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, iv[:blockSize])
	decrypted = make([]byte, len(encrypted))
	blockMode.CryptBlocks(decrypted, encrypted)
	decrypted = pkcs5UnPadding(decrypted)
	return decrypted
}

// AESDecryptECB ...
func AESDecryptECB(encrypted, key []byte) (decrypted []byte) {
	block, _ := aes.NewCipher(key)
	decrypted = make([]byte, len(encrypted))
	size := 16
	for bs, be := 0, size; bs < len(encrypted); bs, be = bs+size, be+size {
		block.Decrypt(decrypted[bs:be], encrypted[bs:be])
	}
	return decrypted
}

func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
