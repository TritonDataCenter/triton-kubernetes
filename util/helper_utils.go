package util

import (
	"math/rand"
	"strings"
	"time"
)

// alpha-numeric potential runes w/ ambiguous characters removed (i,o,q)
var alpha = []rune("abcdefghjklmnprstuvwxyz0123456789")

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// GetAlphaNum returns a random alpha-numeric string
func GetAlphaNum(length int) string {
	res := make([]rune, length)
	for i := range res {
		res[i] = alpha[rand.Intn(len(alpha))]
	}
	return string(res)
}

// IsUnique checks for a unique value within a string array
func IsUnique(arr []string, find string) bool {
	for i := range arr {
		if strings.EqualFold(find, arr[i]) {
			return false
		}
	}
	return true
}
