package create

import (
	"fmt"
	"math/rand"
	"testing"
)

var getNewHostnamesTestCases = []struct {
	ExistingNames []string
	NodeName      string
	NodesToAdd    int
	Expected      []string
}{
	// node count <= 0
	{[]string{"test-1", "test-2"}, "test", 0, []string{}},
	{[]string{"test-1", "test-2"}, "test", -10, []string{}},
	// node count == 1
	{[]string{"test-1", "test-2"}, "bar", 1, []string{"bar-0nxrps"}},
	{[]string{"test"}, "test", 1, []string{"test-hu5g3z"}},
	// node count > 1
	{[]string{"foo", "bar"}, "test", 3, []string{"test-ymbuvf", "test-jfja9l", "test-kzt51k"}},
	{[]string{"test"}, "test", 3, []string{"test-zvf07m", "test-n5ns55", "test-tspstz"}},
	{[]string{"test-1", "test-2", "bar-3", "bar-4"}, "test", 3, []string{"test-tssh74", "test-hhcyz0", "test-3s415m"}},
	// existing name (test-0nxrps & test-hhcyz0 are next 2 iterations)
	{[]string{"test-0xmrt1", "test-hhcyz0"}, "test", 1, []string{"test-f3nab8"}},
}

func TestGetNewHostnames(t *testing.T) {
	rand.Seed(1) // Force seed for reproducible results
	for _, tc := range getNewHostnamesTestCases {
		output := getNewHostnames(tc.ExistingNames, tc.NodeName, tc.NodesToAdd)
		if !isEqual(tc.Expected, output) {
			msg := fmt.Sprintf("\nInput:    (%q, %q, %d)\n", tc.ExistingNames, tc.NodeName, tc.NodesToAdd)
			msg += fmt.Sprintf("Output:   %q\n", output)
			msg += fmt.Sprintf("Expected: %q\n", tc.Expected)
			t.Error(msg)
		}
	}
}

func isEqual(expected, actual []string) bool {
	if len(expected) != len(actual) {
		return false
	}
	for index, expectedItem := range expected {
		if actual[index] != expectedItem {
			return false
		}
	}
	return true
}
