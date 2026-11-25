package main

import "device/arm"

func asm() {
	arm.Asm("nop")
}

func asmFull() uintptr {
	return arm.AsmFull("mrs {}, PRIMASK", nil)
}

func asmFullInput(state uintptr) {
	arm.AsmFull("msr PRIMASK, {state}", map[string]interface{}{
		"state": state,
	})
}

func main() {}
