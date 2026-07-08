package auth

import (
	"crypto/rand"
	"encoding/hex"
)
var Sessions = map[string]string{}
func GenerateSessionID() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
