package util

import (
	"fmt"
	"math/rand"
	"testing"
)

// GetAlphaNum tests
//   Verify random seeding occurs naturally, resulting in unique strings
func TestGetAlphaNum(t *testing.T) {
	s1 := GetAlphaNum(6)
	s2 := GetAlphaNum(6)
	if s1 == s2 {
		t.Errorf("Random alpha strings should NOT match: %s == %s", s1, s2)
	}
}

//   TestGetAlphaNum_Fixed verifies a forced seed provides a repeatable string
func TestGetAlphaNum_Fixed(t *testing.T) {
	rand.Seed(1)
	s1 := GetAlphaNum(6)
	rand.Seed(1)
	s2 := GetAlphaNum(6)
	if s1 != s2 {
		t.Errorf("Random alpha strings SHOULD match: %s != %s", s1, s2)
	}
}

// IsUnique tests
var uniqueTestCases = []struct {
	Search []string
	Find   string
	Result bool
}{
	{[]string{"abcdef", "123456"}, "abcdef", false},
	{[]string{"abcdef", "123456"}, "abc123", true},
}

//  Validate unique testing for test cases above
func TestIsUnique(t *testing.T) {
	for _, tc := range uniqueTestCases {
		if IsUnique(tc.Search, tc.Find) != tc.Result {
			msg := fmt.Sprintf("\nSearch: %q\n", tc.Search)
			msg += fmt.Sprintf("Find:   %q\n", tc.Find)
			msg += fmt.Sprintf("Expected: %v\n", tc.Result)
			t.Error(msg)
		}
	}
}
