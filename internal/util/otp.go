package util

import (
	"crypto/rand"
	"math/big"
	"strings"
)

var numericVals = []rune("0123456789")

func GenerateOTP(size int) (string, error) {
	output := strings.Builder{}

	for range size {
		numericVal, err := rand.Int(rand.Reader, big.NewInt(int64(len(numericVals))))

		if err != nil {
			return "", err
		}

		_, err2 := output.WriteRune(numericVals[numericVal.Int64()])

		if err2 != nil {
			return "", err
		}
	}

	return output.String(), nil
}
