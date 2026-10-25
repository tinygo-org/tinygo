//go:build tinygo && qbs_noinput

package dev

// waitGo does not wait. The target cannot receive input from the host.
func waitGo(class string) {}
