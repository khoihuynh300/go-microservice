package utils

import (
	"crypto/sha256"
	"fmt"
)

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}
