//go:build !tinygo.wasm && !(nrf || (stm32 && !(stm32f103 || stm32l0x1 || stm32g0 || stm32u0 || stm32f401)) || (sam && atsamd51) || (sam && atsame5x) || esp32 || esp32c3 || esp32c6 || esp32s3 || tkey || (tinygo.riscv32 && virt))

package unix

import "syscall"

type GetRandomFlag uintptr

const (
	GRND_NONBLOCK GetRandomFlag = 0x0001
	GRND_RANDOM   GetRandomFlag = 0x0002
)

func GetRandom(p []byte, flags GetRandomFlag) (n int, err error) {
	// Not supported on most TinyGo targets.
	// On real Linux the sysrand package will fall back to /dev/urandom.
	return 0, syscall.ENOSYS
}
