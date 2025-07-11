package uefi

const (
	uintnSize = 32 << (^uintptr(0) >> 63) // 32 or 64nt
	errorMask = 1 << uintptr(uintnSize-1)
)

const (
	EFI_SUCCESS              EFI_STATUS = 0
	EFI_LOAD_ERROR           EFI_STATUS = errorMask | 1
	EFI_INVALID_PARAMETER    EFI_STATUS = errorMask | 2
	EFI_UNSUPPORTED          EFI_STATUS = errorMask | 3
	EFI_BAD_BUFFER_SIZE      EFI_STATUS = errorMask | 4
	EFI_BUFFER_TOO_SMALL     EFI_STATUS = errorMask | 5
	EFI_NOT_READY            EFI_STATUS = errorMask | 6
	EFI_DEVICE_ERROR         EFI_STATUS = errorMask | 7
	EFI_WRITE_PROTECTED      EFI_STATUS = errorMask | 8
	EFI_OUT_OF_RESOURCES     EFI_STATUS = errorMask | 9
	EFI_VOLUME_CORRUPTED     EFI_STATUS = errorMask | 10
	EFI_VOLUME_FULL          EFI_STATUS = errorMask | 11
	EFI_NO_MEDIA             EFI_STATUS = errorMask | 12
	EFI_MEDIA_CHANGED        EFI_STATUS = errorMask | 13
	EFI_NOT_FOUND            EFI_STATUS = errorMask | 14
	EFI_ACCESS_DENIED        EFI_STATUS = errorMask | 15
	EFI_NO_RESPONSE          EFI_STATUS = errorMask | 16
	EFI_NO_MAPPING           EFI_STATUS = errorMask | 17
	EFI_TIMEOUT              EFI_STATUS = errorMask | 18
	EFI_NOT_STARTED          EFI_STATUS = errorMask | 19
	EFI_ALREADY_STARTED      EFI_STATUS = errorMask | 20
	EFI_ABORTED              EFI_STATUS = errorMask | 21
	EFI_ICMP_ERROR           EFI_STATUS = errorMask | 22
	EFI_TFTP_ERROR           EFI_STATUS = errorMask | 23
	EFI_PROTOCOL_ERROR       EFI_STATUS = errorMask | 24
	EFI_INCOMPATIBLE_VERSION EFI_STATUS = errorMask | 25
	EFI_SECURITY_VIOLATION   EFI_STATUS = errorMask | 26
	EFI_CRC_ERROR            EFI_STATUS = errorMask | 27
	EFI_END_OF_MEDIA         EFI_STATUS = errorMask | 28
	EFI_END_OF_FILE          EFI_STATUS = errorMask | 31
	EFI_INVALID_LANGUAGE     EFI_STATUS = errorMask | 32
	EFI_COMPROMISED_DATA     EFI_STATUS = errorMask | 33
	EFI_IP_ADDRESS_CONFLICT  EFI_STATUS = errorMask | 34
	EFI_HTTP_ERROR           EFI_STATUS = errorMask | 35
)

var errMap = make(map[EFI_STATUS]*Error)

var (
	ErrLoadError           = newError(EFI_LOAD_ERROR, "image failed to load")
	ErrInvalidParameter    = newError(EFI_INVALID_PARAMETER, "a parameter was incorrect")
	ErrUnsupported         = newError(EFI_UNSUPPORTED, "operation not supported")
	ErrBadBufferSize       = newError(EFI_BAD_BUFFER_SIZE, "buffer size incorrect for request")
	ErrBufferTooSmall      = newError(EFI_BUFFER_TOO_SMALL, "buffer too small; size returned in parameter")
	ErrNotReady            = newError(EFI_NOT_READY, "no data pending")
	ErrDeviceError         = newError(EFI_DEVICE_ERROR, "physical device reported an error")
	ErrWriteProtected      = newError(EFI_WRITE_PROTECTED, "device is write-protected")
	ErrOutOfResources      = newError(EFI_OUT_OF_RESOURCES, "out of resources")
	ErrVolumeCorrupted     = newError(EFI_VOLUME_CORRUPTED, "filesystem inconsistency detected")
	ErrVolumeFull          = newError(EFI_VOLUME_FULL, "no more space on filesystem")
	ErrNoMedia             = newError(EFI_NO_MEDIA, "device contains no medium")
	ErrMediaChanged        = newError(EFI_MEDIA_CHANGED, "medium changed since last access")
	ErrNotFound            = newError(EFI_NOT_FOUND, "item not found")
	ErrAccessDenied        = newError(EFI_ACCESS_DENIED, "access denied")
	ErrNoResponse          = newError(EFI_NO_RESPONSE, "server not found or no response")
	ErrNoMapping           = newError(EFI_NO_MAPPING, "no device mapping exists")
	ErrTimeout             = newError(EFI_TIMEOUT, "timeout expired")
	ErrNotStarted          = newError(EFI_NOT_STARTED, "protocol not started")
	ErrAlreadyStarted      = newError(EFI_ALREADY_STARTED, "protocol already started")
	ErrAborted             = newError(EFI_ABORTED, "operation aborted")
	ErrICMPError           = newError(EFI_ICMP_ERROR, "ICMP error during network operation")
	ErrTFTPError           = newError(EFI_TFTP_ERROR, "TFTP error during network operation")
	ErrProtocolError       = newError(EFI_PROTOCOL_ERROR, "protocol error during network operation")
	ErrIncompatibleVersion = newError(EFI_INCOMPATIBLE_VERSION, "requested version incompatible")
	ErrSecurityViolation   = newError(EFI_SECURITY_VIOLATION, "security violation")
	ErrCRCError            = newError(EFI_CRC_ERROR, "CRC error detected")
	ErrEndOfMedia          = newError(EFI_END_OF_MEDIA, "beginning or end of media reached")
	ErrEndOfFile           = newError(EFI_END_OF_FILE, "end of file reached")
	ErrInvalidLanguage     = newError(EFI_INVALID_LANGUAGE, "invalid language specified")
	ErrCompromisedData     = newError(EFI_COMPROMISED_DATA, "data security status unknown or compromised")
	ErrIPAddressConflict   = newError(EFI_IP_ADDRESS_CONFLICT, "IP address conflict detected")
	ErrHTTPError           = newError(EFI_HTTP_ERROR, "HTTP error during network operation")
)

type Error struct {
	code EFI_STATUS
	msg  string
}

func newError(code EFI_STATUS, msg string) *Error {
	err := &Error{
		code: code,
		msg:  msg,
	}
	errMap[code] = err
	return err
}

func (e *Error) Error() string {
	return e.msg
}

// StatusError returns the error object given by status. These
// can be checked/managed with errors.Is() and the like.
func StatusError(status EFI_STATUS) *Error {
	if status == 0 {
		return nil
	}
	err, ok := errMap[status]
	if !ok {
		return newError(status, "unknown EFI error")
	}
	return err
}
