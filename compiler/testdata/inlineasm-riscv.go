package main

import "device/riscv"

func csrGet() uintptr {
	return riscv.MSTATUS.Get()
}

func csrSet(value uintptr) {
	riscv.MSTATUS.Set(value)
}

func csrSetBits(mask uintptr) uintptr {
	return riscv.MSTATUS.SetBits(mask)
}

func csrClearBits(mask uintptr) uintptr {
	return riscv.MSTATUS.ClearBits(mask)
}

func main() {}
