// +build gc.debugmetric

package runtime

const (
	gcDebugMetric = true
)

var (
	numGC                 int64
	pauseTotalNanoseconds int64
	start                 int64
)

func gcDebugIncrementNumGC() {
	numGC++
}

func gcDebugGCPauseStart() {
	procPin()
	start = int64(ticks())
	procUnpin()
}

func gcDebugGCPauseEnd() {
	procPin()
	pauseTotalNanoseconds += int64(ticks()) - start
	start = 0
	procUnpin()
}

//go:linkname readGCStats runtime/debug.readGCStats
func readGCStats(data *[2]int64) {
	procPin()
	data[0] = numGC
	data[1] = pauseTotalNanoseconds
	procUnpin()
}
