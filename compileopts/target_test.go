package compileopts

import "testing"

func TestLoadTarget(t *testing.T) {
	_, err := LoadTarget("arduino")
	if err != nil {
		t.Error("LoadTarget test failed:", err)
	}

	_, err = LoadTarget("notexist")
	if err == nil {
		t.Error("LoadTarget should have failed with non existing target")
	}

	if err.Error() != "expected a full LLVM target or a custom target in -target flag" {
		t.Error("LoadTarget failed for wrong reason:", err)
	}
}
