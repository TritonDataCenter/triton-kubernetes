package testutils

import (
	"fmt"
	"math/rand"

	"github.com/sean-/seed"
)

func init() {
	seed.Init()
}

// RandString generates a random alphanumeric string of the length specified
func RandString(strlen int) string {
	return RandStringFromCharSet(strlen, CharSetAlphaNum)
}

// RandPrefixString generates a random alphanumeric string of the length specified
// with the given prefix
func RandPrefixString(prefix string, strlen int) string {
	requiredLength := strlen - len(prefix)
	return fmt.Sprintf("%s%s", prefix, RandString(requiredLength))
}

// RandStringFromCharSet generates a random string by selecting characters from
// the charset provided
func RandStringFromCharSet(strlen int, charSet string) string {
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = charSet[rand.Intn(len(charSet))]
	}
	return string(result)
}

const (
	// CharSetAlphaNum is the alphanumeric character set for use with
	// RandStringFromCharSet
	CharSetAlphaNum = "abcdefghijklmnopqrstuvwxyz012346789"

	// CharSetAlpha is the alphabetical character set for use with
	// RandStringFromCharSet
	CharSetAlpha = "abcdefghijklmnopqrstuvwxyz"
)
