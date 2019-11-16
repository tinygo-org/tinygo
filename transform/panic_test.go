package transform

import (
	"testing"
)

func TestReplacePanicsWithTrap(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/panic", ReplacePanicsWithTrap)
}
