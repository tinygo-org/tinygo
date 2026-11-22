//go:build esp32p4 && scheduler.cores

package runtime

import (
	"device/esp"
	"unsafe"
)

// Start core 1 at call_start_cpu1, following start_other_core in ESP-IDF
// components/esp_system/port/cpu_start.c and esp32p4 hal/cpu_utility_ll.h.
func startCore1() {
	// Unstall core 1 and wait until it runs.
	esp.PMU.CPU_SW_STALL.ReplaceBits(0xff, 0xff, esp.PMU_CPU_SW_STALL_HPCORE1_SW_STALL_CODE_Pos)
	for esp.HP_SYS.CPU_CORESTALLED_ST.HasBits(esp.HP_SYS_CPU_CORESTALLED_ST_REG_CORE1_CORESTALLED_ST) {
	}

	// Enable the clock of core 1 and release it from reset.
	esp.HP_SYS_CLKRST.SOC_CLK_CTRL0.SetBits(esp.HP_SYS_CLKRST_SOC_CLK_CTRL0_REG_CORE1_CPU_CLK_EN)
	esp.HP_SYS_CLKRST.HP_RST_EN0.ClearBits(esp.HP_SYS_CLKRST_HP_RST_EN0_REG_RST_EN_CORE1_GLOBAL)

	// The ROM on core 1 waits for this address and then jumps to it.
	ets_set_appcpu_boot_addr(uintptr(unsafe.Pointer(&call_start_cpu1)))
}

//export ets_set_appcpu_boot_addr
func ets_set_appcpu_boot_addr(addr uintptr)
