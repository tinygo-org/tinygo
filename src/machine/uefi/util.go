//go:build uefi

package uefi

import (
	"unicode/utf16"
	"unicode/utf8"
	"unsafe"
)

var systemTable *EFI_SYSTEM_TABLE
var imageHandle uintptr

//go:nobound
func Init(argImageHandle uintptr, argSystemTable uintptr) {
	systemTable = (*EFI_SYSTEM_TABLE)(unsafe.Pointer(argSystemTable))
	imageHandle = argImageHandle
}

func ST() *EFI_SYSTEM_TABLE {
	return systemTable
}

func BS() *EFI_BOOT_SERVICES {
	return systemTable.BootServices
}

func GetImageHandle() EFI_HANDLE {
	return EFI_HANDLE(imageHandle)
}

func UTF16ToString(input []CHAR16) string {
	return UTF16PtrLenToString(&input[0], len(input))
}

func UTF16PtrToString(input *CHAR16) string {
	pointer := uintptr(unsafe.Pointer(input))
	length := 0
	for *(*CHAR16)(unsafe.Pointer(pointer)) != 0 {
		length++
		pointer += 2
	}
	return UTF16PtrLenToString(input, length)
}

func UTF16PtrLenToString(input *CHAR16, length int) string {
	var output []byte = make([]byte, 0, length)
	var u8buf [4]byte
	inputSlice := unsafe.Slice((*uint16)(unsafe.Pointer(input)), length)
	decodedRunes := utf16.Decode(inputSlice)
	for _, r := range decodedRunes {
		chunkLength := utf8.EncodeRune(u8buf[:], r)
		output = append(output, u8buf[:chunkLength]...)
	}
	return string(output)
}

var hexConst [16]CHAR16 = [16]CHAR16{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'}

var printBuf [256]CHAR16

//go:nobounds
//go:noinline
func DebugPrint(text string, val uint64) {
	var i int
	for i = 0; i < len(text); i++ {
		printBuf[i] = CHAR16(text[i])
	}
	printBuf[i] = CHAR16(':')
	i++
	for j := 0; j < (64 / 4); j++ {
		t := int(val>>((15-j)*4)) & 0xf
		printBuf[i] = hexConst[t]
		i++
	}
	printBuf[i] = '\r'
	i++
	printBuf[i] = '\n'
	i++
	printBuf[i] = 0

	systemTable.ConOut.OutputString(&printBuf[0])
}

func convertBoolean(b BOOLEAN) uintptr {
	if b {
		return 1
	}
	return 0
}
