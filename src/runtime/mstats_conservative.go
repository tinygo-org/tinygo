//go:build gc.conservative
// +build gc.conservative

package runtime

// ReadMemStats populates m with memory statistics.
//
// The returned memory statistics are up to date as of the
// call to ReadMemStats. This would not do GC implicitly for you.
func ReadMemStats(m *MemStats) {
	m.HeapIdle = 0
	m.HeapInuse = 0
	for block := gcBlock(0); block < endBlock; block++ {
		bstate := block.state()
		if bstate == blockStateFree {
			m.HeapIdle += uint64(bytesPerBlock)
		} else {
			m.HeapInuse += uint64(bytesPerBlock)
		}
	}
	m.HeapReleased = 0 // always 0, we don't currently release memory back to the OS.
	m.HeapSys = m.HeapInuse + m.HeapIdle
	m.GCSys = uint64(heapEnd - uintptr(metadataStart))
	m.Sys = uint64(heapEnd - heapStart)
}
