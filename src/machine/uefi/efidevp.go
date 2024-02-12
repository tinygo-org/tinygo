package uefi

// Device Path structures - Section C

// EFI_DEVICE_PATH_PROTOCOL structure
type EFI_DEVICE_PATH_PROTOCOL struct {
	Type    uint8
	SubType uint8
	Length  [2]uint8
}
