package create

import (
	"fmt"
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
	{[]string{"test-1", "test-2"}, "bar", 1, []string{"bar"}},
	{[]string{"test"}, "test", 1, []string{"test-1"}},
	// node count > 1
	{[]string{"foo", "bar"}, "test", 3, []string{"test-1", "test-2", "test-3"}},
	{[]string{"test"}, "test", 3, []string{"test-1", "test-2", "test-3"}},
	{[]string{"test-1", "test-2", "bar-3", "bar-4"}, "test", 3, []string{"test-3", "test-4", "test-5"}},
}

func TestGetNewHostnames(t *testing.T) {
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
