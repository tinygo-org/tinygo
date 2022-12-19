//go:build windows || darwin || (linux && !baremetal)

package os_test

import (
	"os"
	"testing"
)

func TestGetpagesize(t *testing.T) {
	pagesize := os.Getpagesize()
	if pagesize == 0x1000 || pagesize == 0x4000 || pagesize == 0x10000 {
		return
	}
	t.Errorf("os.Getpagesize() returns strange value %d", pagesize)
}
