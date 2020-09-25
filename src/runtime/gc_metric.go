// +build !gc.debugmetric

package runtime

const (
	gcDebugMetric = false
)

func gcDebugIncrementNumGC() {}

func gcDebugGCPauseStart() {}

func gcDebugGCPauseEnd() {}

//go:linkname readGCStats runtime/debug.readGCStats
func readGCStats(data *[2]int64) {
	panic("unimplemented: debug.readGCStats is only available with 'gc.debugmetric' build tag")
}
