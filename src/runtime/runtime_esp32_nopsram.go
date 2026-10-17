//go:build esp32s3 && !numa_psram_octal

package runtime

// initPSRAM is a no-op when PSRAM support is disabled.
func initPSRAM() {}
