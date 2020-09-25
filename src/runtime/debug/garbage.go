package debug

import "time"

type GCStats struct {
	PauseTotal time.Duration
	NumGC      int64
}

func ReadGCStats(stats *GCStats) {
	data := [2]int64{}
	readGCStats(&data)

	stats.NumGC = data[0]
	stats.PauseTotal = time.Duration(data[1])
}
