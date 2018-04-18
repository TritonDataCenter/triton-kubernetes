package test_pkg

import (
	"fmt"
	"strings"
	"testing"
)

type T struct {
	t *testing.T
}

func NewT(t *testing.T) T {
	return T{t: t}
}

func (t *T) Fatal(context string, expected, actual interface{}) {
	divider := strings.Repeat("=", len(context))

	output := fmt.Sprintf(
		"\n\n%v\n%v\n\nexpected:\n\n\t%v\n\ngot:\n\n\t%v\n",
		context,
		divider,
		expected,
		actual,
	)

	t.t.Fatal(output)
}

func (t *T) Logf(fmt string, args ...interface{}) {
	t.t.Logf(fmt, args...)
}
