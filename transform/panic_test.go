package transform_test

import (
	"testing"

	"github.com/tinygo-org/tinygo/transform"
)

func TestReplacePanicsWithTrap(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/panic", transform.ReplacePanicsWithTrap)
}
