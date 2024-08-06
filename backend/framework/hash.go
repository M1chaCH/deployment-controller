package framework

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"github.com/cespare/xxhash/v2"
)

func QuickHash(value string) uint64 {
	data := []byte(value)
	return xxhash.Sum64(data)
}

func SecureHash(value string, salt []byte) string {
	combined := append(salt, []byte(value)...)
	hash := sha256.New()
	hash.Write(combined)
	hashSum := hash.Sum(nil)
	return hex.EncodeToString(hashSum)
}

func SecureHashWithSalt(value string) (string, []byte, error) {
	salt, err := GenerateSalt()
	if err != nil {
		return "", nil, err
	}

	return SecureHash(value, salt), salt, nil
}

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}
