//go:build stm32f7

package machine

// eraseBlockSize returns the smallest erasable unit for the STM32F7 internal
// flash. The first sectors are 32 KB; return that as the nominal page size.
// Flash write/erase via machine.Flash is not implemented for STM32F7; this
// stub satisfies the flash.go interface so that BlockDevice (needed by MSC)
// compiles on this target.
func eraseBlockSize() int64 { return 32768 }
