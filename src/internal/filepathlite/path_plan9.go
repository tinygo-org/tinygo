package filepathlite

// IsAbs reports whether the path is absolute.
func IsAbs(path string) bool {
	return len(path) > 0 && (path[0] == '/' || path[0] == '#')
}
