package utils

import (
	"crypto/sha1"
	"encoding/base64"
)

func ShortenURL(url string) string {
	h := sha1.New()
	h.Write([]byte(url))
	bs := h.Sum(nil)
	encoded := base64.URLEncoding.EncodeToString(bs)
	// Используем первые 8 символов для короткого URL
	short := encoded[:8]
	return short
}
