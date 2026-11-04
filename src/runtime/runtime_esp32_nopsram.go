//go:build esp32 && !numa_psram_qspi

package runtime

func initPSRAM() {}

func reportPSRAM() {}
