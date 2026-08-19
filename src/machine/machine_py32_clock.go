//go:build py32

package machine

func CPUFrequency() uint32 {
	return cpuFrequencyHz
}
