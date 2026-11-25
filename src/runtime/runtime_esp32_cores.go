//go:build esp32 && scheduler.cores

package runtime

import (
	"device/esp"
	"unsafe"
)

// Set (1) or clear (0) the FROM_CPU interrupt of the given core.
func signalCore(core, value uint32) {
	if core == 0 {
		esp.DPORT.CPU_INTR_FROM_CPU_0.Set(value)
	} else {
		esp.DPORT.CPU_INTR_FROM_CPU_1.Set(value)
	}
}

func crossCoreChipHandler(core uint32) {
}

func initCore1() {
	esp.DPORT.PRO_CPU_INTR_FROM_CPU_0_MAP.Set(crossCoreCPUInterrupt)
	esp.DPORT.APP_CPU_INTR_FROM_CPU_1_MAP.Set(crossCoreCPUInterrupt)
	initCore1Cache()
}

func resetCore1() {
	esp.DPORT.SetAPPCPU_CTRL_B_APPCPU_CLKGATE_EN(1)
	esp.DPORT.SetAPPCPU_CTRL_C_APPCPU_RUNSTALL(0)
	esp.DPORT.SetAPPCPU_CTRL_A_APPCPU_RESETTING(1)
	esp.DPORT.SetAPPCPU_CTRL_A_APPCPU_RESETTING(0)
}

const (
	flashPageSize = 0x10000
	dromVaddr     = 0x3F400000
	iromVaddr     = 0x400D0000
)

// call_start_cpu0 maps flash only for the PRO CPU cache. Map it for the APP CPU as
// esp-idf does in bootloader_support/src/esp32/bootloader_esp32.c and bootloader_utility.c.
func initCore1Cache() {
	Cache_Read_Disable(1)
	Cache_Flush(1)
	esp.DPORT.SetAPP_CACHE_CTRL1_APP_CACHE_MMU_IA_CLR(1)
	mmu_init(1)
	esp.DPORT.SetAPP_CACHE_CTRL1_APP_CACHE_MMU_IA_CLR(0)

	paddr := dromFlashAddr
	if n := flashPages(&rodataStart, &rodataEnd); n != 0 {
		cache_flash_mmu_set(1, 0, dromVaddr, paddr, 64, n)
		paddr += n * flashPageSize
	}
	if n := flashPages(&textStart, &textEnd); n != 0 {
		cache_flash_mmu_set(1, 0, iromVaddr, paddr, 64, n)
	}

	esp.DPORT.SetAPP_CACHE_CTRL1_APP_CACHE_MASK_IRAM0(0)
	esp.DPORT.SetAPP_CACHE_CTRL1_APP_CACHE_MASK_DROM0(0)
	Cache_Flush(1)
	Cache_Read_Enable(1)
}

func flashPages(start, end *[0]byte) uint32 {
	size := uint32(uintptr(unsafe.Pointer(end)) - uintptr(unsafe.Pointer(start)))
	return (size + flashPageSize - 1) / flashPageSize
}

//go:extern _drom_flash_addr
var dromFlashAddr uint32

//go:extern _rodata_start
var rodataStart [0]byte

//go:extern _rodata_end
var rodataEnd [0]byte

//go:extern _text_start
var textStart [0]byte

//go:extern _text_end
var textEnd [0]byte

//export Cache_Read_Disable
func Cache_Read_Disable(cpu uint32)

//export Cache_Read_Enable
func Cache_Read_Enable(cpu uint32)

//export Cache_Flush
func Cache_Flush(cpu uint32)

//export mmu_init
func mmu_init(cpu uint32)

//export cache_flash_mmu_set
func cache_flash_mmu_set(cpu, pid, vaddr, paddr, psize, num uint32) int32
