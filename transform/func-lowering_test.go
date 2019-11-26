package transform

import (
	"testing"
)

func TestFuncLowering(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/func-lowering", LowerFuncValues)
}
