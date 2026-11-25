package os_test

import (
	"os"
	"testing"
)

func TestNextRandom(t *testing.T) {
	for range 100 {
		s := os.NextRandom()
		if len(s) == 0 || len(s) > 10 || (len(s) == 10 && s > "4294967295") {
			t.Fatalf("nextRandom() = %q, want a decimal uint32", s)
		}
		if len(s) > 1 && s[0] == '0' {
			t.Fatalf("nextRandom() = %q, want no leading zero", s)
		}
		for _, c := range s {
			if c < '0' || c > '9' {
				t.Fatalf("nextRandom() = %q, want decimal digits", s)
			}
		}
	}
}
