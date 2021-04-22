package transform_test

import (
	"testing"

	"github.com/tinygo-org/tinygo/transform"
)

func TestFuncLowering(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/func-lowering", transform.LowerFuncValues)
}
