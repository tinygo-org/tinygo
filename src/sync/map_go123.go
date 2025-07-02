//go:build go1.23

package sync

// Go 1.23 added the Clear() method. The clear() function is added in Go 1.21,
// so this method can be moved to map.go once we drop support for Go 1.20 and
// below.

func (m *Map) Clear() {
	m.lock.Lock()
	defer m.lock.Unlock()
	clear(m.m)
}
