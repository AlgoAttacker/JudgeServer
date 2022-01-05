package main

import (
	"crypto/rand"
	"math/big"
	"os"
)

func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	return false
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func randomString(length int) string {
	result := make([]rune, length)
	for i := range result {
		randomInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(letterRunes))))

		if err != nil {
			continue
		}

		result[i] = letterRunes[int(randomInt.Int64())]
	}
	return string(result)
}
