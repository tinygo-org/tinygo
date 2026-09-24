//go:build esp32s3 && scheduler.cores

package runtime

import (
	"device"
	"device/esp"
	"internal/task"
	"runtime/interrupt"
	"runtime/volatile"
	"sync/atomic"
	"unsafe"
)

const numCPU = 2

const crosscoreCPUInt = 12

const (
	crosscoreReasonWake = 1 << iota
	crosscoreReasonGC
)

var (
	printLock     spinLock
	schedulerLock spinLock
	atomicsLock   spinLock
	futexLock     spinLock
)

var sleepingCore uint8 = 0xff
var waitingCores uint8
var cpu1Started atomic.Uint32
var crosscoreReason [numCPU]atomic.Uint32
var gcSignalWait volatile.Register8

func hasSleepingCore() bool {
	return sleepingCore != 0xff
}

func sleepTicksMulticore(d timeUnit) {
	sleepingCore = uint8(currentCPU())
	schedulerLock.Unlock()
	sleepTicks(d)
	schedulerLock.Lock()
	sleepingCore = 0xff
}

func interruptSleepTicksMulticore(wakeup timeUnit) {
	_ = wakeup
	schedulerWake()
}

func schedulerUnlockAndWait() {
	core := currentCPU()
	waitingCores |= uint8(1 << core)
	schedulerLock.Unlock()
	device.Asm("waiti 0")
	schedulerLock.Lock()
	waitingCores &^= uint8(1 << core)
}

func schedulerWake() {
	if waitingCores == 0 {
		return
	}
	core := currentCPU() ^ 1
	if waitingCores&(1<<core) == 0 {
		core ^= 1
	}
	sendCrosscoreInterrupt(core, crosscoreReasonWake)
}

func currentCPU() uint32 {
	prid := uintptr(device.AsmFull("rsr.prid {}", nil))
	return uint32((prid >> 13) & 1)
}

func startSecondaryCores() {
	initCrosscoreInterrupt(0)

	esp.RTC_CNTL.SetOPTIONS0_SW_STALL_APPCPU_C0(0)
	esp.RTC_CNTL.SetSW_CPU_STALL_SW_STALL_APPCPU_C1(0)

	esp.SYSTEM.SetCORE_1_CONTROL_0_CONTROL_CORE_1_CLKGATE_EN(1)
	esp.SYSTEM.SetCORE_1_CONTROL_0_CONTROL_CORE_1_RUNSTALL(0)
	esp.SYSTEM.SetCORE_1_CONTROL_0_CONTROL_CORE_1_RESETING(1)
	esp.SYSTEM.SetCORE_1_CONTROL_0_CONTROL_CORE_1_RESETING(0)

	etsSetAppCPUBootAddr(uint32(uintptr(unsafe.Pointer(&callStartCPU1))))

	for i := 0; i < 1000000 && cpu1Started.Load() == 0; i++ {
		spinLoopWait()
	}
}

func gcPauseCore(core uint32) {
	sendCrosscoreInterrupt(core, crosscoreReasonGC)
}

func gcSignalCore(core uint32) {
	gcSignalWait.Set(1)
	sendCrosscoreInterrupt(core, crosscoreReasonGC)
}

func coreStackTop(core uint32) uintptr {
	switch core {
	case 0:
		return uintptr(unsafe.Pointer(&stackTopSymbol))
	case 1:
		return uintptr(unsafe.Pointer(&stack1TopSymbol))
	default:
		runtimePanic("unexpected core")
		return 0
	}
}

func spinLoopWait() {
	device.Asm("nop")
}

//export tinygo_runCore1
func runCore1() {
	interruptInit()
	initCrosscoreInterrupt(1)
	etsSetAppCPUBootAddr(0)
	cpu1Started.Store(1)
	schedulerLock.Lock()
	scheduler(false)
	schedulerLock.Unlock()
	exit(0)
}

func initCrosscoreInterrupt(core uint32) {
	if core == 0 {
		esp.INTERRUPT_CORE0.SetCPU_INTR_FROM_CPU_0_MAP(crosscoreCPUInt)
	} else {
		esp.INTERRUPT_CORE1.SetCPU_INTR_FROM_CPU_1_MAP(crosscoreCPUInt)
	}
	intr := interrupt.New(crosscoreCPUInt, crosscoreInterruptHandler)
	_ = intr.Enable()
}

func crosscoreInterruptHandler(interrupt.Interrupt) {
	handleCrosscoreInterrupt(currentCPU())
}

func sendCrosscoreInterrupt(core uint32, reason uint32) {
	crosscoreReason[core].Or(reason)
	if core == 0 {
		esp.SYSTEM.SetCPU_INTR_FROM_CPU_0(1)
	} else {
		esp.SYSTEM.SetCPU_INTR_FROM_CPU_1(1)
	}
}

func clearCrosscoreInterrupt(core uint32) {
	if core == 0 {
		esp.SYSTEM.SetCPU_INTR_FROM_CPU_0(0)
	} else {
		esp.SYSTEM.SetCPU_INTR_FROM_CPU_1(0)
	}
}

func handleCrosscoreInterrupt(core uint32) {
	clearCrosscoreInterrupt(core)
	reason := crosscoreReason[core].Swap(0)
	if reason&crosscoreReasonGC != 0 {
		gcInterruptHandler(core)
	}
}

func gcInterruptHandler(hartID uint32) {
	gcScanState.Add(1)
	for gcSignalWait.Get() == 0 {
		spinLoopWait()
	}
	gcSignalWait.Set(0)

	scanCurrentStack()
	if !task.OnSystemStack() {
		markRoots(task.SystemStack(), coreStackTop(hartID))
	}

	gcScanState.Store(1)
	for gcSignalWait.Get() == 0 {
		spinLoopWait()
	}
	gcSignalWait.Set(0)
	gcScanState.Add(1)
}

type spinLock struct {
	atomic.Uint32
}

func (l *spinLock) Lock() {
	for !l.CompareAndSwap(0, 1) {
		spinLoopWait()
	}
}

func (l *spinLock) Unlock() {
	if schedulerAsserts && l.Load() != 1 {
		runtimePanic("unlock of unlocked spinlock")
	}
	l.Store(0)
}

//go:extern _stack1_top
var stack1TopSymbol [0]uint32

//go:extern call_start_cpu1
var callStartCPU1 [0]uint32

//go:linkname etsSetAppCPUBootAddr ets_set_appcpu_boot_addr
func etsSetAppCPUBootAddr(addr uint32)
