package fail_test

import "testing"

func TestFail(t *testing.T) {
	t.Error("fail")
}
