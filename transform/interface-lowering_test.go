package transform

import (
	"testing"
)

func TestInterfaceLowering(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/interface", LowerInterfaces)
}
