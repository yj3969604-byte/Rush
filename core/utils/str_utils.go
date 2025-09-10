package utils

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandomString(length int) string {
	sb := strings.Builder{}
	for i := 0; i < length; i++ {
		index := rand.IntN(len(charset))
		sb.WriteRune(rune(charset[index]))
	}
	return sb.String()
}

func RandomPhone() string {
	index1 := rand.IntN(9) + 1
	index2 := rand.IntN(10)
	index3 := rand.IntN(10)
	index4 := rand.IntN(10)
	return fmt.Sprintf("%d***%d%d%d", index1, index2, index3, index4)
}
