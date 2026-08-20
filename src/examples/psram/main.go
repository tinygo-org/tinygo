package main

// This example demonstrates how to use external PSRAM (Octal or Quad SPI) on
// supported targets (e.g., -target=esp32s3-psram-qspi).
//
// Variables placed in the `.psram` section via the `//go:section .psram`
// pragma are stored in external PSRAM rather than internal SRAM.
// Note: importing "unsafe" is required to enable `//go:section`.

import (
	"fmt"
	"time"
	"unsafe"
)

// Allocate a 1MB buffer explicitly in external PSRAM.
//
//go:section .psram
var psramBuffer [1024 * 1024]byte

func main() {
	time.Sleep(2 * time.Second)

	fmt.Println("=== PSRAM Example ===")
	fmt.Printf("psramBuffer start address: 0x%X\n", uintptr(unsafe.Pointer(&psramBuffer[0])))
	fmt.Printf("psramBuffer size:          %d bytes\n", len(psramBuffer))

	fmt.Println("Writing test pattern to PSRAM...")
	for i := range psramBuffer {
		psramBuffer[i] = byte(i ^ 0xAA)
	}

	fmt.Println("Verifying test pattern...")
	errors := 0
	for i := range psramBuffer {
		if psramBuffer[i] != byte(i^0xAA) {
			errors++
		}
	}

	if errors == 0 {
		fmt.Println("SUCCESS: PSRAM 1MB read/write verification passed!")
	} else {
		fmt.Printf("FAIL: %d readback errors detected\n", errors)
		printed := 0
		for i := range psramBuffer {
			if psramBuffer[i] != byte(i^0xAA) {
				fmt.Printf("  mismatch at idx %d (offset 0x%X): got 0x%02X, want 0x%02X\n", i, i, psramBuffer[i], byte(i^0xAA))
				printed++
				if printed >= 8 {
					break
				}
			}
		}
	}

	for {
		time.Sleep(time.Second)
	}
}
