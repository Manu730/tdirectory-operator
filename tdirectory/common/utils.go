package common

import (
	"fmt"
	"math/rand"
)

const (
	alphaBytes    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	alphaNumeric  = "1234567890" + alphaBytes
	letterIdxBits = 6
	letterIdxMask = 1<<letterIdxBits - 1
)

func SecureRandomAlphaNumericString(length int) string {
	return SecureRandomString(alphaNumeric, length)
}

func SecureRandomAlphaString(length int) string {
	return SecureRandomString(alphaBytes, length)
}

func SecureRandomString(charset string, length int) string {
	result := make([]byte, length)
	bufferSize := int(float64(length) * 1.3)
	for i, j, randomBytes := 0, 0, []byte{}; i < length; j++ {
		if j%bufferSize == 0 {
			randomBytes = secureRandomBytes(bufferSize)
		}

		if idx := int(randomBytes[j%length] & letterIdxMask); idx < len(charset) {
			result[i] = charset[idx]
			i++
		}
	}
	return string(result)
}

func secureRandomBytes(length int) []byte {
	var randomBytes = make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		fmt.Println("Unable to generate random bytes")
	}
	return randomBytes
}
