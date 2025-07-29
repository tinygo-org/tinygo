//go:build uefi

package syscall

import (
	"errors"
	"machine/uefi"
	"unsafe"
)

// EnvVendor is the UUID used to specify VendorGUID for Getting and Setting
// non-volatile EFI variables. Override this value to have a private variable
// space for an EFI application.
//
// The default value was generated with
//
//	uuidgen -N TinyGo -n @oid --sha1
var EnvVendor = uefi.EFI_GUID{
	0xaabc54d8, 0x7b8e, 0x5680,
	[...]byte{0x9f, 0x6c, 0x68, 0xda, 0x0d, 0xbb, 0xcd, 0xbf}}

func Environ() []string {
	// our implementation of runtime envs is already a freshly-made copy
	// so just return that copy, rather than make a second, redudant copy.
	return runtime_envs()
}

func Getenv(key string) (value string, found bool) {
	var data []byte
	keyC16 := uefi.StringToUTF16(key)
	data, found = getkey(keyC16)
	if found {
		return string(data), true
	}
	return "", false
}

func getkey(key []uefi.CHAR16) (value []byte, found bool) {
	var (
		dataSize uefi.UINTN
		data     []byte = make([]byte, 1)
	)
	// call with dataSize = 0 so we can be told how big the buffer needs to be; this is passed through dataSize
	status := uefi.ST().RuntimeServices.GetVariable(
		(*uefi.CHAR16)(unsafe.Pointer(&key[0])), // Variable Name
		&EnvVendor,                              // Vendor GUID
		nil,                                     // optional attributes
		&dataSize,                               // Before call: Size of Buffer; after call: how much was written to the buffer
		(*uefi.VOID)(unsafe.Pointer(&data[0])))
	switch status {
	case uefi.EFI_NOT_FOUND:
		return nil, false
	case uefi.EFI_BUFFER_TOO_SMALL:
		data = make([]byte, int(dataSize))
		status = uefi.ST().RuntimeServices.GetVariable(
			(*uefi.CHAR16)(unsafe.Pointer(&key[0])), // Variable Name
			&EnvVendor,                              // Vendor GUID
			nil,                                     // optional attributes
			&dataSize,                               // Before call: Size of Buffer; after call: how much was written to the buffer
			(*uefi.VOID)(unsafe.Pointer(&data[0])))
		if status != uefi.EFI_SUCCESS {
			return nil, false
		}
		return data, true
	default:
		return nil, false
	}
}

func Setenv(key, val string) error {
	if key == "" {
		return errors.New("key must not be empty")
	}

	keyC16 := uefi.StringToUTF16(key)
	data := []byte(val)

	status := uefi.ST().RuntimeServices.SetVariable(
		(*uefi.CHAR16)(unsafe.Pointer(&keyC16[0])),
		&EnvVendor,
		uefi.EFI_VARIABLE_NON_VOLATILE|
			uefi.EFI_VARIABLE_BOOTSERVICE_ACCESS|
			uefi.EFI_VARIABLE_RUNTIME_ACCESS,
		uefi.UINTN(len(data)),
		(*uefi.VOID)(unsafe.Pointer(&data[0])),
	)

	if status == uefi.EFI_SUCCESS {
		return nil
	}

	return uefi.StatusError(status)
}

func Unsetenv(key string) (err error) {
	// Setting a key to a dataSize of 0 deletes it
	keyC16 := uefi.StringToUTF16(key)

	status := uefi.ST().RuntimeServices.SetVariable(
		(*uefi.CHAR16)(unsafe.Pointer(&keyC16[0])),
		&EnvVendor,
		uefi.EFI_VARIABLE_NON_VOLATILE|
			uefi.EFI_VARIABLE_BOOTSERVICE_ACCESS|
			uefi.EFI_VARIABLE_RUNTIME_ACCESS,
		uefi.UINTN(0),
		nil,
	)

	if status == uefi.EFI_SUCCESS {
		return nil
	}

	return uefi.StatusError(status)
}

func Clearenv() (err error) {
	// stub for now
	return ENOSYS
}

func runtime_envs() (env []string) {
	// GetNextVariableName is a bit screwy, per the spec:
	// To start, you specify a pointer to a null, that is,
	// you send a null-terminated, zero-length string
	// You supply the last returned variable to GetNextVariableName
	// to get the next variable. status is set to
	// uefi.EFI_NOT_FOUND when there are no more variables
	// to return. No filtering happens via vendorGUID, despite
	// the spec saying this is both IN and OUT, it is only
	// OUT.

	var (
		varKey     = make([]uefi.CHAR16, 1)
		varKeySize = uefi.UINTN(2)
		status     uefi.EFI_STATUS
		vendorGUID uefi.EFI_GUID
	)
	for {
		status = uefi.ST().RuntimeServices.GetNextVariableName(
			&varKeySize,
			(*uefi.CHAR16)(unsafe.Pointer(&varKey[0])),
			&vendorGUID)

		switch status {
		case uefi.EFI_BUFFER_TOO_SMALL:
			// buffer was too small, the size needed will be in
			// varKeySize.
			newVarKey := make([]uefi.CHAR16, varKeySize)
			copy(newVarKey, varKey)
			varKey = newVarKey[:varKeySize/2]
			continue
		case uefi.EFI_SUCCESS:
			// we read a variable name, if it's ours,  get it
			// and append it to our list
			if vendorGUID != EnvVendor {
				continue
			}
			keyString := uefi.UTF16ToString(varKey[:varKeySize/2])
			val, _ := getkey(varKey[:varKeySize/2])
			env = append(env, keyString+"="+string(val))
		case uefi.EFI_NOT_FOUND:
			// all done!
			return
		case uefi.EFI_INVALID_PARAMETER:
			// this gets its own branch because it means something in this function
			// was done incorrectly.
			panic("invalid parameter passed to GetNextVariableName")
		default:
			// something went wrong; cheese it
			panic(uefi.StatusError(status))
		}
	}
}
