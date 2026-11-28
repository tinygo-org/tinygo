package goenv

import "testing"

// TestGetCgoEnabled checks that Get("CGO_ENABLED") honours the caller's value
// and keeps the stock default "1" when it is unset or empty.
func TestGetCgoEnabled(t *testing.T) {
	for _, tc := range []struct{ env, want string }{{"", "1"}, {"0", "0"}, {"1", "1"}} {
		t.Setenv("CGO_ENABLED", tc.env)
		if got := Get("CGO_ENABLED"); got != tc.want {
			t.Errorf("CGO_ENABLED=%q: Get = %q, want %q", tc.env, got, tc.want)
		}
	}
}
