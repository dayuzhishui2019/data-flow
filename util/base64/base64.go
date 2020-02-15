package base64

import (
	"encoding/base64"
)

func Decode(raw string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(raw)
}

func Encode(raw []byte) string {
	return base64.StdEncoding.EncodeToString(raw)
}
