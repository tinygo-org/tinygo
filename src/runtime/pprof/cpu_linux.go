//go:build linux && scheduler.threads && (amd64 || arm64)

package pprof

import (
	"os"
	"strconv"
	"strings"
)

// fileMappings returns the executable mappings of the process.
func fileMappings() []fileMapping {
	data, err := os.ReadFile("/proc/self/maps")
	if err != nil {
		return nil
	}
	var mappings []fileMapping
	for _, line := range strings.Split(string(data), "\n") {
		// Each line is address range, permissions, offset, device, inode and
		// path, which may contain spaces. See proc_pid_maps(5).
		var fields [6]string
		rest := line
		for i := 0; i < 5; i++ {
			fields[i], rest, _ = strings.Cut(strings.TrimLeft(rest, " "), " ")
		}
		fields[5] = strings.TrimLeft(rest, " ")
		if fields[5] == "" || !strings.Contains(fields[1], "x") {
			continue
		}
		start, end, ok := strings.Cut(fields[0], "-")
		if !ok {
			continue
		}
		s, err1 := strconv.ParseUint(start, 16, 64)
		e, err2 := strconv.ParseUint(end, 16, 64)
		off, err3 := strconv.ParseUint(fields[2], 16, 64)
		if err1 != nil || err2 != nil || err3 != nil {
			continue
		}
		mappings = append(mappings, fileMapping{start: uintptr(s), end: uintptr(e), offset: uintptr(off), file: fields[5]})
	}
	return mappings
}

func lookupSystem(pc uintptr) (lib string, base uintptr, sym string, ok bool) {
	return "", 0, "", false
}
