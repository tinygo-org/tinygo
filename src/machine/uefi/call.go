package uefi

//go:nosplit
//export uefiCall0
func UefiCall0(fn uintptr) EFI_STATUS

//go:nosplit
//go:export uefiCall1
func UefiCall1(fn uintptr, a uintptr) EFI_STATUS

//go:nosplit
//go:export uefiCall2
func UefiCall2(fn uintptr, a uintptr, b uintptr) EFI_STATUS

//go:nosplit
//go:export uefiCall3
func UefiCall3(fn uintptr, a uintptr, b uintptr, c uintptr) EFI_STATUS

//go:nosplit
//go:export uefiCall4
func UefiCall4(fn uintptr, a uintptr, b uintptr, c uintptr, d uintptr) EFI_STATUS

//go:nosplit
//go:export uefiCall5
func UefiCall5(fn uintptr, a uintptr, b uintptr, c uintptr, d uintptr, e uintptr) EFI_STATUS

//go:nosplit
//go:export uefiCall6
func UefiCall6(fn uintptr, a uintptr, b uintptr, c uintptr, d uintptr, e uintptr, f uintptr) EFI_STATUS
