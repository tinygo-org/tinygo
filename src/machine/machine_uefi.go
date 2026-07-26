//go:build uefi

package machine

import (
	deviceuefi "device/uefi"
)

const deviceName = "UEFI"

type (
	EFI_STATUS            = deviceuefi.EFI_STATUS
	EFI_TIME              = deviceuefi.EFI_TIME
	EFI_TIME_CAPABILITIES = deviceuefi.EFI_TIME_CAPABILITIES
	TextOutput            = deviceuefi.TextOutput

	Error = deviceuefi.Error
)

const (
	EFI_SUCCESS              = deviceuefi.EFI_SUCCESS
	EFI_LOAD_ERROR           = deviceuefi.EFI_LOAD_ERROR
	EFI_INVALID_PARAMETER    = deviceuefi.EFI_INVALID_PARAMETER
	EFI_UNSUPPORTED          = deviceuefi.EFI_UNSUPPORTED
	EFI_BAD_BUFFER_SIZE      = deviceuefi.EFI_BAD_BUFFER_SIZE
	EFI_BUFFER_TOO_SMALL     = deviceuefi.EFI_BUFFER_TOO_SMALL
	EFI_NOT_READY            = deviceuefi.EFI_NOT_READY
	EFI_DEVICE_ERROR         = deviceuefi.EFI_DEVICE_ERROR
	EFI_WRITE_PROTECTED      = deviceuefi.EFI_WRITE_PROTECTED
	EFI_OUT_OF_RESOURCES     = deviceuefi.EFI_OUT_OF_RESOURCES
	EFI_VOLUME_CORRUPTED     = deviceuefi.EFI_VOLUME_CORRUPTED
	EFI_VOLUME_FULL          = deviceuefi.EFI_VOLUME_FULL
	EFI_NO_MEDIA             = deviceuefi.EFI_NO_MEDIA
	EFI_MEDIA_CHANGED        = deviceuefi.EFI_MEDIA_CHANGED
	EFI_NOT_FOUND            = deviceuefi.EFI_NOT_FOUND
	EFI_ACCESS_DENIED        = deviceuefi.EFI_ACCESS_DENIED
	EFI_NO_RESPONSE          = deviceuefi.EFI_NO_RESPONSE
	EFI_NO_MAPPING           = deviceuefi.EFI_NO_MAPPING
	EFI_TIMEOUT              = deviceuefi.EFI_TIMEOUT
	EFI_NOT_STARTED          = deviceuefi.EFI_NOT_STARTED
	EFI_ALREADY_STARTED      = deviceuefi.EFI_ALREADY_STARTED
	EFI_ABORTED              = deviceuefi.EFI_ABORTED
	EFI_ICMP_ERROR           = deviceuefi.EFI_ICMP_ERROR
	EFI_TFTP_ERROR           = deviceuefi.EFI_TFTP_ERROR
	EFI_PROTOCOL_ERROR       = deviceuefi.EFI_PROTOCOL_ERROR
	EFI_INCOMPATIBLE_VERSION = deviceuefi.EFI_INCOMPATIBLE_VERSION
	EFI_SECURITY_VIOLATION   = deviceuefi.EFI_SECURITY_VIOLATION
	EFI_CRC_ERROR            = deviceuefi.EFI_CRC_ERROR
	EFI_END_OF_MEDIA         = deviceuefi.EFI_END_OF_MEDIA
	EFI_END_OF_FILE          = deviceuefi.EFI_END_OF_FILE
	EFI_INVALID_LANGUAGE     = deviceuefi.EFI_INVALID_LANGUAGE
	EFI_COMPROMISED_DATA     = deviceuefi.EFI_COMPROMISED_DATA
	EFI_IP_ADDRESS_CONFLICT  = deviceuefi.EFI_IP_ADDRESS_CONFLICT
	EFI_HTTP_ERROR           = deviceuefi.EFI_HTTP_ERROR
)

var (
	ErrLoadError           = deviceuefi.ErrLoadError
	ErrInvalidParameter    = deviceuefi.ErrInvalidParameter
	ErrUnsupported         = deviceuefi.ErrUnsupported
	ErrBadBufferSize       = deviceuefi.ErrBadBufferSize
	ErrBufferTooSmall      = deviceuefi.ErrBufferTooSmall
	ErrNotReady            = deviceuefi.ErrNotReady
	ErrDeviceError         = deviceuefi.ErrDeviceError
	ErrWriteProtected      = deviceuefi.ErrWriteProtected
	ErrOutOfResources      = deviceuefi.ErrOutOfResources
	ErrVolumeCorrupted     = deviceuefi.ErrVolumeCorrupted
	ErrVolumeFull          = deviceuefi.ErrVolumeFull
	ErrNoMedia             = deviceuefi.ErrNoMedia
	ErrMediaChanged        = deviceuefi.ErrMediaChanged
	ErrNotFound            = deviceuefi.ErrNotFound
	ErrAccessDenied        = deviceuefi.ErrAccessDenied
	ErrNoResponse          = deviceuefi.ErrNoResponse
	ErrNoMapping           = deviceuefi.ErrNoMapping
	ErrTimeout             = deviceuefi.ErrTimeout
	ErrNotStarted          = deviceuefi.ErrNotStarted
	ErrAlreadyStarted      = deviceuefi.ErrAlreadyStarted
	ErrAborted             = deviceuefi.ErrAborted
	ErrICMPError           = deviceuefi.ErrICMPError
	ErrTFTPError           = deviceuefi.ErrTFTPError
	ErrProtocolError       = deviceuefi.ErrProtocolError
	ErrIncompatibleVersion = deviceuefi.ErrIncompatibleVersion
	ErrSecurityViolation   = deviceuefi.ErrSecurityViolation
	ErrCRCError            = deviceuefi.ErrCRCError
	ErrEndOfMedia          = deviceuefi.ErrEndOfMedia
	ErrEndOfFile           = deviceuefi.ErrEndOfFile
	ErrInvalidLanguage     = deviceuefi.ErrInvalidLanguage
	ErrCompromisedData     = deviceuefi.ErrCompromisedData
	ErrIPAddressConflict   = deviceuefi.ErrIPAddressConflict
	ErrHTTPError           = deviceuefi.ErrHTTPError
)

func GetTime() (EFI_TIME, EFI_STATUS) {
	return deviceuefi.GetTime()
}

func ConsoleOut() *TextOutput {
	return deviceuefi.ConsoleOut()
}

func StandardError() *TextOutput {
	return deviceuefi.StandardError()
}

func StatusError(status EFI_STATUS) *Error {
	return deviceuefi.StatusError(status)
}

func (Pin) Configure(PinConfig) {
}

func (Pin) Set(bool) {
}

func (Pin) Get() bool {
	return false
}
