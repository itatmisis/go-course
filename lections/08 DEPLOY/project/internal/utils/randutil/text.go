package randutil

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const alphanumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func Alphanumeric(length int) (string, error) {
	alphabetSize := big.NewInt(int64(len(alphanumeric)))
	result := make([]byte, length)

	for i := range result {
		index, err := rand.Int(rand.Reader, alphabetSize)
		if err != nil {
			return "", fmt.Errorf("generate random index: %w", err)
		}

		result[i] = alphanumeric[index.Int64()]
	}

	return string(result), nil
}
