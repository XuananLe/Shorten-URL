package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"
)

func Hash(originalURL string) string {
	timestamp := time.Now().UnixNano()

	dataToHash := fmt.Sprintf("%s%d", originalURL, timestamp)

	hash := sha256.Sum256([]byte(dataToHash))

	encoded := base64.URLEncoding.EncodeToString(hash[:])

	shortened := encoded[:8]

	return shortened
}
