//go:build uefi

package uefi

// ResetType specifies the type of system reset.
type ResetType uint32

const (
	// ResetCold causes a system-wide reset that resets all processors and devices.
	ResetCold ResetType = iota
	// ResetWarm causes a system-wide initialization without resetting processors.
	ResetWarm
	// ResetShutdown causes the system to enter a power state equivalent to ACPI G2/S5 or G3.
	ResetShutdown
	// ResetPlatformSpecific causes a platform-specific reset type.
	ResetPlatformSpecific
)

// Reset performs a system reset.
// resetType specifies the type of reset (ResetCold, ResetWarm, ResetShutdown, ResetPlatformSpecific).
// status is the status code for the reset (typically 0 for Success).
// This function does not return on success.
func Reset(resetType ResetType, status Status) {
	st := GetSystemTable()
	if st == nil || st.RuntimeServices == nil || st.RuntimeServices.ResetSystem == 0 {
		return
	}
	Call(
		st.RuntimeServices.ResetSystem,
		uintptr(resetType),
		uintptr(status),
		0, // DataSize
		0, // ResetData (NULL)
	)
	// Should not return
}

// Reboot performs a cold reset of the system.
// This function does not return on success.
func Reboot() {
	Reset(ResetCold, Success)
}

// Shutdown powers off the system.
// This function does not return on success.
func Shutdown() {
	Reset(ResetShutdown, Success)
}

// SetWatchdogTimer sets, resets, or disables the watchdog timer.
// timeout is the number of seconds to wait before the watchdog fires (0 disables it).
// By default UEFI sets a 5-minute watchdog; call SetWatchdogTimer(0) to disable it.
func SetWatchdogTimer(timeout uintptr) Status {
	st := GetSystemTable()
	if st == nil || st.BootServices == nil || st.BootServices.SetWatchdogTimer == 0 {
		return ErrUnsupported
	}
	// SetWatchdogTimer(Timeout, WatchdogCode, DataSize, WatchdogData)
	return Call(
		st.BootServices.SetWatchdogTimer,
		timeout,
		0, // WatchdogCode
		0, // DataSize
		0, // WatchdogData (NULL)
	)
}

// DisableWatchdog disables the UEFI watchdog timer.
func DisableWatchdog() Status {
	return SetWatchdogTimer(0)
}

// Exit terminates the UEFI application with the specified exit code.
func Exit(code int) {
	st := GetSystemTable()
	if st == nil || st.BootServices == nil {
		return
	}
	Call(
		st.BootServices.Exit,
		uintptr(ImageHandle()),
		uintptr(code),
		0,
		0,
	)
}
