//go:build py32

package machine

import "device/py32"

var cpuFrequencyHz = initialCPUFrequency()

func initialCPUFrequency() uint32 {
	if py32.CPU == "CM4" {
		return 8_000_000
	}
	return 24_000_000
}

func CPUFrequency() uint32 {
	return cpuFrequencyHz
}

func setCPUFrequency(frequency uint32) {
	cpuFrequencyHz = frequency
}
