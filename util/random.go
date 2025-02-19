package util

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GenerateRandomString(n uint) (string, error) {
	var sb strings.Builder
	k := len(letters)

	for i := uint(0); i < n; i++ {
		// Generate a random index within the range of letters
		index, err := rand.Int(rand.Reader, big.NewInt(int64(k)))
		if err != nil {
			return "", err
		}
		sb.WriteByte(letters[index.Int64()])
	}

	return sb.String(), nil
}
