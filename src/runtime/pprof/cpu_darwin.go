//go:build darwin

package pprof

func executable() (path string, start, end uintptr)
func lookupSystem(pc uintptr) (lib string, base uintptr, sym string, ok bool)

func fileMappings() []fileMapping {
	path, start, end := executable()
	return []fileMapping{{start: start, end: end, file: path}}
}
