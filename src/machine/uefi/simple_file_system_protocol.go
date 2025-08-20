package uefi

import (
	"unsafe"
)

//---------------------------------------------------------------------------
//  GUIDs (§13.5 File Protocol and related info GUIDs)
//---------------------------------------------------------------------------

// EFI_SIMPLE_FILE_SYSTEM_PROTOCOL_GUID = {964E5B22-6459-11D2-8E39-00A0C969723B}
var EFI_SIMPLE_FILE_SYSTEM_PROTOCOL_GUID = EFI_GUID{
	0x964E5B22, 0x6459, 0x11D2,
	[8]uint8{0x8e, 0x39, 0x00, 0xa0, 0xc9, 0x69, 0x72, 0x3b},
}

// gEfiFileInfoGuid = {09576E92-6D3F-11D2-8E39-00A0C969723B}
var EFI_FILE_INFO_ID = EFI_GUID{
	0x09576e92, 0x6d3f, 0x11d2,
	[8]uint8{0x8e, 0x39, 0x00, 0xa0, 0xc9, 0x69, 0x72, 0x3b},
}

// gEfiFileSystemInfoGuid = {09576E93-6D3F-11D2-8E39-00A0C969723B}
var EFI_FILE_SYSTEM_INFO_ID = EFI_GUID{
	0x09576e93, 0x6d3f, 0x11d2,
	[8]uint8{0x8e, 0x39, 0x00, 0xa0, 0xc9, 0x69, 0x72, 0x3b},
}

// gEfiFileSystemVolumeLabelInfoIdGuid = {DB47D7D3-FE81-11D3-9A35-0090273FC14D}
var EFI_FILE_SYSTEM_VOLUME_LABEL_ID = EFI_GUID{
	0xdb47d7d3, 0xfe81, 0x11d3,
	[8]uint8{0x9a, 0x35, 0x00, 0x90, 0x27, 0x3f, 0xc1, 0x4d},
}

//---------------------------------------------------------------------------
//  Constants (§13.5.1)
//---------------------------------------------------------------------------

// Open modes are 64-bit flags
const (
	EFI_FILE_MODE_READ   uint64 = 0x0000000000000001
	EFI_FILE_MODE_WRITE  uint64 = 0x0000000000000002
	EFI_FILE_MODE_CREATE uint64 = 0x8000000000000000
)

// Attribute flags (also 64-bit)
const (
	EFI_FILE_READ_ONLY uint64 = 0x0000000000000001
	EFI_FILE_HIDDEN    uint64 = 0x0000000000000002
	EFI_FILE_SYSTEM    uint64 = 0x0000000000000004
	EFI_FILE_RESERVED  uint64 = 0x0000000000000008
	EFI_FILE_DIRECTORY uint64 = 0x0000000000000010
	EFI_FILE_ARCHIVE   uint64 = 0x0000000000000020
)

// Optional: known protocol revision values (not strictly required to use)
const (
	EFI_FILE_PROTOCOL_REVISION  = 0x00010000
	EFI_FILE_PROTOCOL_REVISION2 = 0x00020000
	EFI_FILE_PROTOCOL_LATEST    = EFI_FILE_PROTOCOL_REVISION2
)

//---------------------------------------------------------------------------
//  Protocols (§13.4 Simple File System, §13.5 File Protocol)
//---------------------------------------------------------------------------

// EFI_SIMPLE_FILE_SYSTEM_PROTOCOL (§13.4.2)
// Function table: Revision, OpenVolume
// OpenVolume returns a handle to the volume root directory (EFI_FILE_PROTOCOL*)

type EFI_SIMPLE_FILE_SYSTEM_PROTOCOL struct {
	Revision   uint64
	openVolume uintptr // (this, **EFI_FILE_PROTOCOL)
}

func (p *EFI_SIMPLE_FILE_SYSTEM_PROTOCOL) OpenVolume(root **EFI_FILE_PROTOCOL) EFI_STATUS {
	return UefiCall2(
		p.openVolume,
		uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(root)),
	)
}

// EFI_FILE_PROTOCOL (§13.5.2)
// Function table order: Revision, Open, Close, Delete, Read, Write, GetPosition, SetPosition, GetInfo, SetInfo, Flush

type EFI_FILE_PROTOCOL struct {
	Revision    uint64
	open        uintptr // (this, **newHandle, *FileName, OpenMode, Attributes)
	close       uintptr // (this)
	delete      uintptr // (this)
	read        uintptr // (this, *BufferSize, Buffer)
	write       uintptr // (this, *BufferSize, Buffer)
	getPosition uintptr // (this, *Position)
	setPosition uintptr // (this, Position)
	getInfo     uintptr // (this, *InfoType, *BufferSize, Buffer)
	setInfo     uintptr // (this, *InfoType, BufferSize, Buffer)
	flush       uintptr // (this)
}

// Open opens a file or directory relative to this handle (which may be a root or directory handle).
// FileName must be a NUL-terminated UTF-16 path using '\\' separators. (§13.5.3)
func (p *EFI_FILE_PROTOCOL) Open(newHandle **EFI_FILE_PROTOCOL, fileName *CHAR16, openMode uint64, attributes uint64) EFI_STATUS {
	return UefiCall6(
		p.open,
		uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(newHandle)),
		uintptr(unsafe.Pointer(fileName)),
		uintptr(openMode),
		uintptr(attributes),
		0,
	)
}

// Close closes the file handle. (§13.5.4)
func (p *EFI_FILE_PROTOCOL) Close() EFI_STATUS {
	return UefiCall1(p.close, uintptr(unsafe.Pointer(p)))
}

// Delete deletes the file opened by this handle. (§13.5.5)
func (p *EFI_FILE_PROTOCOL) Delete() EFI_STATUS {
	return UefiCall1(p.delete, uintptr(unsafe.Pointer(p)))
}

// Read reads from the file into Buffer. BufferSize is in/out: on entry, size of Buffer; on return, bytes read. (§13.5.6)
func (p *EFI_FILE_PROTOCOL) Read(bufferSize *UINTN, buffer unsafe.Pointer) EFI_STATUS {
	return UefiCall3(
		p.read,
		uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(bufferSize)),
		uintptr(buffer),
	)
}

// Write writes to the file from Buffer. BufferSize is in/out: on entry, bytes to write; on return, bytes written. (§13.5.7)
func (p *EFI_FILE_PROTOCOL) Write(bufferSize *UINTN, buffer unsafe.Pointer) EFI_STATUS {
	return UefiCall3(
		p.write,
		uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(bufferSize)),
		uintptr(buffer),
	)
}

// GetPosition returns the current file position in bytes. (§13.5.8)
func (p *EFI_FILE_PROTOCOL) GetPosition(position *uint64) EFI_STATUS {
	return UefiCall2(
		p.getPosition,
		uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(position)),
	)
}

// SetPosition sets the current file position. Setting to 0xFFFFFFFFFFFFFFFF seeks to end-of-file. (§13.5.9)
func (p *EFI_FILE_PROTOCOL) SetPosition(position uint64) EFI_STATUS {
	return UefiCall2(
		p.setPosition,
		uintptr(unsafe.Pointer(p)),
		uintptr(position),
	)
}

// GetInfo retrieves metadata identified by InformationType. Common GUIDs: EFI_FILE_INFO_ID, EFI_FILE_SYSTEM_INFO_ID, EFI_FILE_SYSTEM_VOLUME_LABEL_ID. (§13.5.10)
func (p *EFI_FILE_PROTOCOL) GetInfo(infoType *EFI_GUID, bufferSize *UINTN, buffer unsafe.Pointer) EFI_STATUS {
	return UefiCall4(
		p.getInfo,
		uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(infoType)),
		uintptr(unsafe.Pointer(bufferSize)),
		uintptr(buffer),
	)
}

// SetInfo sets metadata identified by InformationType. (§13.5.11)
func (p *EFI_FILE_PROTOCOL) SetInfo(infoType *EFI_GUID, bufferSize UINTN, buffer unsafe.Pointer) EFI_STATUS {
	return UefiCall4(
		p.setInfo,
		uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(infoType)),
		uintptr(bufferSize),
		uintptr(buffer),
	)
}

// Flush flushes file data and metadata to the device. (§13.5.12)
func (p *EFI_FILE_PROTOCOL) Flush() EFI_STATUS {
	return UefiCall1(p.flush, uintptr(unsafe.Pointer(p)))
}

//---------------------------------------------------------------------------
//  Info structures (§13.5.13, §13.5.15)
//---------------------------------------------------------------------------

// EFI_FILE_INFO head (FileName is a variable-length CHAR16[] immediately following this struct)
// Layout must exactly match the spec; FileName is retrieved from the backing buffer used with GetInfo.

type EFI_FILE_INFO struct {
	Size             uint64
	FileSize         uint64
	PhysicalSize     uint64
	CreateTime       EFI_TIME
	LastAccessTime   EFI_TIME
	ModificationTime EFI_TIME
	Attribute        uint64
	// CHAR16 FileName[] follows
}

// EFI_FILE_SYSTEM_INFO head (VolumeLabel is a variable-length CHAR16[] after the struct)

type EFI_FILE_SYSTEM_INFO struct {
	Size       uint64
	ReadOnly   bool
	_          [7]byte // pad to 8-byte alignment (bool is 1 byte)
	VolumeSize uint64
	FreeSpace  uint64
	BlockSize  uint32
	_          uint32 // padding to keep next CHAR16[] aligned as in C
	// CHAR16 VolumeLabel[] follows
}

// EFI_FILE_SYSTEM_VOLUME_LABEL is just a CHAR16[] volume label; retrieved via GetInfo with EFI_FILE_SYSTEM_VOLUME_LABEL_ID.
// Represented implicitly by reading a buffer of UTF-16 data.

//---------------------------------------------------------------------------
//  Helpers: enumerate volumes and open root
//---------------------------------------------------------------------------

// EnumerateFileSystems locates all Simple File System handles and returns their protocol pointers.
func EnumerateFileSystems() ([]*EFI_SIMPLE_FILE_SYSTEM_PROTOCOL, error) {
	var (
		handleCount  UINTN
		handleBuffer *EFI_HANDLE
	)
	st := BS().LocateHandleBuffer(ByProtocol, &EFI_SIMPLE_FILE_SYSTEM_PROTOCOL_GUID, nil, &handleCount, &handleBuffer)
	if st == EFI_NOT_FOUND {
		return nil, nil
	}
	if st != EFI_SUCCESS {
		return nil, StatusError(st)
	}
	handles := unsafe.Slice((*EFI_HANDLE)(unsafe.Pointer(handleBuffer)), int(handleCount))
	out := make([]*EFI_SIMPLE_FILE_SYSTEM_PROTOCOL, 0, len(handles))
	for _, h := range handles {
		var sfs *EFI_SIMPLE_FILE_SYSTEM_PROTOCOL
		if st = BS().HandleProtocol(h, &EFI_SIMPLE_FILE_SYSTEM_PROTOCOL_GUID, unsafe.Pointer(&sfs)); st == EFI_SUCCESS {
			out = append(out, sfs)
		}
	}
	return out, nil
}

// OpenRoot opens the volume root directory for a given SFS.
func OpenRoot(sfs *EFI_SIMPLE_FILE_SYSTEM_PROTOCOL) (*EFI_FILE_PROTOCOL, EFI_STATUS) {
	var root *EFI_FILE_PROTOCOL
	st := sfs.OpenVolume(&root)
	if st != EFI_SUCCESS {
		return nil, st
	}
	return root, EFI_SUCCESS
}
