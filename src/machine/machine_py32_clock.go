//go:build py32

package machine

var cpuFrequencyHz uint32 = defaultCPUFrequencyHz

func CPUFrequency() uint32 {
	return cpuFrequencyHz
}

func setCPUFrequency(frequency uint32) {
	cpuFrequencyHz = frequency
}
