package transform

import (
	"testing"
)

func TestAllocs(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/allocs", OptimizeAllocs)
}
