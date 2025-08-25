package main

import "device"

// Test the specific bug case: AsmFull should use call-time input register value (42), not post-call value (44)
func testAsmFullBug() uintptr {
	place := make(map[string]interface{})
	place["input"] = uint32(42)

	// This should use input=42 at call time, return the moved value
	result := device.AsmFull("mov {}, {input}", place)

	// This update should NOT affect the AsmFull result (was the bug)
	place["input"] = uint32(44)

	return result
}

func main() {
	_ = testAsmFullBug()
}
