package helpers

import (
	"crypto/rand"
	"fmt"
)

func RandBytesHex(n int) string {
	return fmt.Sprintf("%x", RandBytes(n))
}

func RandBytes(n int) []byte {
	res := make([]byte, n)
	rand.Read(res)
	return res
}