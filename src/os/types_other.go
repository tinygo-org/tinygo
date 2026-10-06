//go:build baremetal || (tinygo.wasm && !wasip1 && !wasip2 && !js) || nintendoswitch

package os

// SameFile reports whether fi1 and fi2 describe the same file. Stat is not
// implemented on this target, so no FileInfo here is the os package's own, and
// upstream Go answers false for any other kind. bbolt calls it.
func SameFile(fi1, fi2 FileInfo) bool {
	return false
}
