package transform_test

import (
	"testing"

	"github.com/tinygo-org/tinygo/transform"
)

func TestAllocs(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/allocs", transform.OptimizeAllocs)
}
